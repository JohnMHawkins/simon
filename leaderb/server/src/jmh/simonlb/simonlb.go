// App - Main file for the server application
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




package main

import (
	"net/http"
	"jmh/goweb/webber"
	"jmh/goweb/logger"
	"jmh/goweb/wtmcache"
	"gopkg.in/mgo.v2"
	"os"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
)

// uncomment to enable profiling on the /debug/pprof/ endpoint
import _ "net/http/pprof"


// Global instantiations
var ourConfig *webber.ServerConfig

var cDb *wtmcache.Db					// the db we will use

var httpClient *webber.HttpClient

// this is the AWS session for any api calls we need to make
var awsSession *session.Session


// this is the struct we use for keeping data about our logged in user
//
type UserSessionData struct {
	Username string		`json:"username"`
}


///////////////////////
// helper functions
//

// returns the current session, if one exists.  Otherwise writes a 404 to the response
func GetUserSessionHandleError(w http.ResponseWriter, r *http.Request) *UserSessionData {
	session := UserSessionData{}
	bHasSession, _ := webber.GetSession(r, &session)
	if ( bHasSession ) {
		return &session
	} else {
		http.Error(w, "NoSession", http.StatusUnauthorized)
		return nil
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
	AWSProfile := flag.String("awsprofile", "", "What AWS security profile from the user/.aws/credentials file to use")
	DBPath := flag.String("dbpath", "", "Path to the db")
	flag.Parse()

	// read our config
	ourConfig = webber.LoadConfig(*configFile)

	//set any cmd line overrides
	if len(*AWSRegion) > 0 {
		ourConfig.AWSRegion = *AWSRegion
	}
	if len(*AWSProfile) > 0 {
		ourConfig.AWSProfile = *AWSProfile
	}
	if len(*DBPath) > 0 {
		ourConfig.DBPath = *DBPath
	}

	
	////////////////////////////
	// set up our logger
	// fill out the AppInfo struct so the logger knows who it is writting logs for:
	app := logger.AppInfo{Name:ourConfig.AppName, Version:ourConfig.AppVersion, Instance:*AppInstance,Cluster:*AppCluster}
	fhLogger := logger.NewFirehoseLogger(app, ourConfig.AWSRegion, ourConfig.AWSProfile, ourConfig.LoggerFirehoseDeliveryStream)
	logger.StdLogger = fhLogger
	logger.StdLogger.StdOutOn(true)
	logger.StdLogger.LOG(logger.INFO, "", "SimonLB is starting up", nil)

	// connect to our db
	dbSession, dbErr := mgo.Dial(ourConfig.DBPath)
	if ( dbErr != nil ) {
		// can't connect to our db
		logger.StdLogger.LOG(logger.CRITICAL, "", fmt.Sprintf("Can't connect to db: %s", dbErr), nil)
		os.Exit(1)
	}
	cDb = wtmcache.NewDb(dbSession, "simonlb")
	webber.CreateSessionDbCollection(cDb, ourConfig.SessionCollName)
	wtmcache.CreateAutoIncDbCollection(cDb)
	wtmcache.EnsureAutoIncrement("simonusers")
	wtmcache.EnsureAutoIncrement("tourneys")
	wtmcache.EnsureAutoIncrement("teams")
	CreateUserDbCollection(cDb)
	CreateTeamDbCollection(cDb)
	CreateTourneyDbCollection(cDb)

	httpClient = webber.NewHttpClient(nil);

	// create an App Server
	as := webber.NewAppServer(ourConfig)

	//////////////////////////////////
	// create a couple of handlers

	// create our auth handler and assign it to <apibase>/auth
	auths := NewAuthServer(ourConfig.ApiBase + "/auth")
	as.RegisterHandler(auths)

	// create our teams handler and assign it to <apibase>/teams
	teams := NewTeamServer(ourConfig.ApiBase + "/teams")
	as.RegisterHandler(teams)

	// create an AWS session
	var awsErr error
	awsSession, awsErr = session.NewSession(&aws.Config{Region:aws.String(ourConfig.AWSRegion)})
	if (awsErr != nil ) {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Unable to create AWS session: %s", awsErr.Error()), nil)
	}

	// now start the server
	http.HandleFunc("/", as.Handler)
	http.ListenAndServe(ourConfig.Port, nil)
	
}