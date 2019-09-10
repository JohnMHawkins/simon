// wtmcache - Write-Through Mongo cache for Go
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

package wtmcache

import (
	"time"
	"gopkg.in/mgo.v2"
)

// WTMCachedDb is the struct that holds information on the database itself
//
type Db struct {
	Db *mgo.Database
}

// GetSession establishes a session with the mongo cluster identified by 
// the url, using the mgo.v2 driver and returning a pointer to a mgo.Session 
// 
// If the port is not specified in the url, it will default to 27017.  See
// documentation for gopkg.in/mgo.v2 for further details on the session
// object.
//
// Call session.Close() on the returned object to close it
//
// Example:
// session, err := mgo.Dial("127.0.0.1:27017")
//
func GetSession ( url string) (*mgo.Session, error) {
	return mgo.Dial(url)
}


// NewWTMCachedDb takes a mgo.v2 session object and a db name, and returns a WTMCachedDb
// object.
//
// Example:
// session, err := mgo.Dial("127.0.0.1:27017")
// wtmCachedDb := NewWTMCachedDb(session, "test")
//
func NewDb ( session *mgo.Session, dbName string) *Db {
	w := new(Db)
	w.Db = session.DB(dbName)
	return w
}

// NewCollection is a Method for Db that takes a collection name, keyfield and cache curation values
// and returns a pointer to a cache-backed collection object (WTMCachedCollection) object.
//
// Parameters:
//	collectionName string:	the name of the collection with the db
//	keyField string:		the name of the field that should be used as the cache key
//	cacheDur time.Duration	expiration time of the cache.  Typically specified by something like 10 * time.Minutes/\.
//							if this value is < 1 or NoExpiration, then the cache never expires
// 	cleanupInterval time.Duration	how frequently expired items are cleaned up in the cache.
//
// Example:
// session, err := mgo.Dial("127.0.0.1:27017")
// wtmCachedDb := NewDb(session, "test")
// wtmCachedColl := wtmCachedDb.NewCollection("foo", "id", 59 * time.Minute, 59 * time.Minute)
//

func (w *Db) NewCollection(collectionName string, keyField string, cacheDur time.Duration, cleanupInterval time.Duration) *Collection {
	return newCollection (w, collectionName, keyField, cacheDur, cleanupInterval)
	
}

