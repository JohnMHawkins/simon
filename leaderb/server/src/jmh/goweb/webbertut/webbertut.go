// webbertut - Tutorial for using the webber package in Golang
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



//////////////////////////////////////////////////////
// This demonstrates use of the webber library
//
//

package main

import (
	"net/http"
	"jmh/goweb/webber"
	"jmh/goweb/logger"
	"jmh/goweb/wtmcache"
	"gopkg.in/mgo.v2"
	"os"
	"flag"
//	"time"
	"encoding/json"
	"io/ioutil"
//	"math/rand"
//	"gopkg.in/mgo.v2/bson"
	"fmt"
)

// uncomment to enable profiling on the /debug/pprof/ endpoint
import _ "net/http/pprof"


// Global instantiations

var cDb *wtmcache.Db					// the db we will use

var httpClient *webber.HttpClient


// this is the struct we use for keeping data about our logged in user
// It's only a sample, so it doesn't store much, but you can add more information, such as
// permissions or preferences.
//
type UserSessionData struct {
	Username string		`json:"username"`
}

// This is our api/auth handler, which will do a simple login and save session auth info
//

type AuthServer struct {
	basePath string
}

func NewAuthServer(basePath string) *AuthServer {
	f := new(AuthServer)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h AuthServer) Name() string {
	return "AuthServer"
}

func (h AuthServer) BasePath() string {
	return h.basePath
}

func (h AuthServer) Handler ( w http.ResponseWriter, r *http.Request) { 
	webber.DispatchMethod(h, w, r);
}

func (h AuthServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathVars := map[int]string{1:"a"}
	logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("AuthServer GET handler called for %s", apiPath), nil)
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, pathVars )

	switch pathParts[0] {
	case "check":
		// get the session if one exists
		// NOTE: this only works if we're a single instance.  If we are multiple instances,
		// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
		//var session UserSessionData
		session := UserSessionData{}
		bHasSession, sessionKey := webber.GetSession(r, &session)


		if ( bHasSession ) {
			//  log it and write back a page.  
			logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("found sesssion %s", session), nil)
			fmt.Fprintf(w, "<html><body>The session key is %s for username %s</body></html>", sessionKey, session.Username)
		} else {
			fmt.Fprintf(w, "<html><body>No active session found</body></html>")
		}
	case "logout":
		// in this case, we would clear the session entry in the db and the header itself
		webber.ClearSession(w)
		fmt.Fprintf(w, "ok")
	
	}
	
}

func (h AuthServer) HandlePost (w http.ResponseWriter, r *http.Request) {
	parseErr := r.ParseForm()
	if parseErr != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("error parsing login form: %s", parseErr), nil)
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "dog" && password == "bark" {
		// that's our login!  Go ahead and make a session
		// store it in the db and set the header.
		
		// this is the session data we want to store.  For this example, it's just the username, but
		// it could be anything we need to keep track of or check for each call, such as permissions,
		// preferences, etc.
		sessionData := UserSessionData{Username:username}
		sk, err := webber.MakeSession(w, sessionData);
		fmt.Println("Created a session with the key ", sk, " err = ", err)
		fmt.Fprintf(w, "Success")
		return		
	} else {
		logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Invalid login for username: %s", username), nil)
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
	}

}



type HikeInfo struct {
	Name string
	Length int
	Description string
}


// This is our api/hike handler, which will tell us simple things about hikes
//

type HikeServer struct {
	basePath string
}

func NewHikeServer(basePath string) *HikeServer {
	f := new(HikeServer)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h HikeServer) Name() string {
	return "HikeServer"
}

func (h HikeServer) BasePath() string {
	return h.basePath
}

func (h HikeServer) Handler ( w http.ResponseWriter, r *http.Request) { 
	apiPath := r.URL.Path[len(h.basePath):]
	logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("HikeServer Handler  called for %s", apiPath), nil)
	webber.DispatchMethod(h, w, r);
}

// fetch information about a hike from our cache server
func (h HikeServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	// api is in the form /api/hike/<hikename>/describe
	//                    /api/hike/<hikename>/length
	pathVars := map[int]string{0:"hike_name"}
	pathParts, vars := webber.ParsePathAndQueryFlat(r, apiPath, pathVars )

	url := "http://localhost:8090/api/cache/hikes/hikes/Name/" + vars["hike_name"]
	resp, rerr := httpClient.Get(url, r)
	fmt.Println("tried caache server, err=", rerr)



	hikeInfo := new(HikeInfo)

	json.NewDecoder(resp.Body).Decode(hikeInfo)

	if ( len(pathParts) == 0) {
		// just return the hike info
		webber.ReturnJson(w, hikeInfo)
	} else {
		switch pathParts[0] {
		case "describe":
			webber.ReturnJson(w, hikeInfo.Description)
		case "length":
			webber.ReturnJson(w, hikeInfo.Length)
		default:
			http.Error(w, "NYI", http.StatusNotImplemented)
		}
	
	}
	
}

// take information about a hike and store it to our cache server
func (h HikeServer) HandlePost (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )

	if (len(pathParts) > 0) {
		hikename := pathParts[0]
		body, err := ioutil.ReadAll(r.Body)
		if ( err == nil ) {
			url := "http://localhost:8090/api/cache/hikes/hikes/Name/" + hikename
			resp, rerr := httpClient.Post(url, body, r.Header.Get("Content-Type"), r)
			if ( rerr == nil) {
				fmt.Fprintf(w, "%d bytes written", len(body))
			} else {
				http.Error(w, rerr.Error(), resp.StatusCode)
			}
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	} else {
		http.Error(w, "no hikename specified", http.StatusBadRequest)
	}
					
}



// main func - we'll load our config, set up a logger,create an AppServer, and add two handlers
// one for auth, the other to return data about Hikes,  then start serving.
//
func main() {

	// get any cmd line flags
	configFile := flag.String("config", "config.json", "path to the config file")
	AppInstance := flag.String("instance", "", "instance name")
	AppCluster := flag.String("cluster", "", "Name for the cluster")
	AWSRegion := flag.String("awsregion", "", "What AWS region we should look for resources in")
	DBPath := flag.String("dbpath", "", "Path to the db")
	flag.Parse()

	// read our config
	config := webber.LoadConfig(*configFile)

	//set any cmd line overrides
	if len(*AWSRegion) > 0 {
		config.AWSRegion = *AWSRegion
	}
	if len(*DBPath) > 0 {
		config.DBPath = *DBPath
	}

	
	////////////////////////////
	// set up our logger
	// fill out the AppInfo struct so the logger knows who it is writting logs for:
	app := logger.AppInfo{Name:config.AppName, Version:config.AppVersion, Instance:*AppInstance,Cluster:*AppCluster}
	fhLogger := logger.NewFirehoseLogger(app, config.AWSRegion, config.AWSProfile, config.LoggerFirehoseDeliveryStream)
	logger.StdLogger = fhLogger
	logger.StdLogger.StdOutOn(true)
	logger.StdLogger.LOG(logger.INFO, "", "WebberTut starting up", nil)

	// connect to our db
	dbSession, dbErr := mgo.Dial(config.DBPath)
	if ( dbErr != nil ) {
		// can't connect to our db
		logger.StdLogger.LOG(logger.CRITICAL, "", fmt.Sprintf("Can't connect to db: %s", dbErr), nil)
		os.Exit(1)
	}
	cDb = wtmcache.NewDb(dbSession, "tutorial")
	webber.CreateSessionDbCollection(cDb, config.SessionCollName)

	httpClient = webber.NewHttpClient(nil);

	// create an App Server
	as := webber.NewAppServer(config)

	//////////////////////////////////
	// create a couple of handlers

	// create our auth handler and assign it to <apibase>/auth
	auths := NewAuthServer(config.ApiBase + "/auth")
	as.RegisterHandler(auths)

	// add a hike handler and assing it <apibase>/hike
	hikes := NewHikeServer(config.ApiBase + "/hike")
	as.RegisterHandler(hikes)

	// now start the server
	http.HandleFunc("/", as.Handler)
	http.ListenAndServe(config.Port, nil)
	
}