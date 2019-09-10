// webber - WebServer package
//
// Copyright (c) 2018 - John M. Hawkins <jmhawkins@msn.com>
//
// All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and 
// associated documentation files (the "Software"), to deal in the Software without restriction, 
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, 
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial 
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
// NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//

package webber

import (
	"fmt"
	"net/http"
	"math/rand"
	"time"
	"jmh/goweb/wtmcache"
	"jmh/goweb/logger"
	"encoding/json"
)

var sessionColl *wtmcache.Collection
var sessionHeader string = "Session"
var sessionLength int = 0				// 0 - no expiration, otherwise seconds

type sessionDocument struct {
	SessionKey string `json:"sessionkey"`
	Data []byte  `json:"data"`
}

// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
func CreateSessionDbCollection ( cDb *wtmcache.Db, sessionCollName string)  {
	sessionColl = cDb.NewCollection(sessionCollName, "sessionkey", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

// MakeSession creates a session key, adds it as a cookie, writes any provided sessionData to
// the session collection (if one exists) and returns the sessionKey
// 
// Parameters:
//	w : the response writer, to add the session header to
//	sessionData : an interface{} for json-serializable structure to store as session data with the sesion
//
// Returns:
//	the sessionKey created
//
func MakeSession (w http.ResponseWriter, sessionData interface{}) (string, error) {
	sessionKey := fmt.Sprintf("%d", rand.Uint64())
	cookie := http.Cookie{Name: sessionHeader, Value: sessionKey, Path: "/", HttpOnly: true, MaxAge: sessionLength}
	http.SetCookie(w, &cookie)
	var err error
	if ( sessionData != nil && sessionColl != nil ) {
		dataJson, err := json.Marshal(sessionData)
		if ( err == nil) {
			doc := sessionDocument{SessionKey:sessionKey, Data:dataJson}
			logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintln("writing session to db ", doc), nil)
			sessionColl.Write(doc)
		}
	}
	return sessionKey, err
}

// GetSession returns a session if one exists 
//
// Params:
//	r :	the request to get header info from
//
// returns:
//	bool true if a session exists
//	string that is the session key
//	an interface{} object for any session data stored by MakeSessionKey
//
func GetSession ( r *http.Request, data interface{}) (bool, string) {
	session, err := r.Cookie(sessionHeader)	
	bHaveSession := false
	sessionKey := ""
	if err == nil  {
		// check if expired
		if session.Expires.IsZero() || session.Expires.After(time.Now()) {
			bHaveSession = true
			sessionKey = session.Value
			if  (sessionColl != nil) {
				var docTemplate sessionDocument
				doc, dbErr := sessionColl.Read(session.Value, &docTemplate) 
				if ( dbErr == nil ) {
					err := json.Unmarshal(doc.(*sessionDocument).Data, data)
					if ( err != nil) {
						logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintln("Error decoding session data ", err, " for session ", session.Value), nil)
					}
				}
			}
		}
	}
	return bHaveSession, sessionKey
}

// Clears any session
//
func ClearSession(w http.ResponseWriter) {
	cookie := http.Cookie{Name: sessionHeader, Value: "", Path: "/", HttpOnly: true, MaxAge: 0}
	http.SetCookie(w, &cookie)

}