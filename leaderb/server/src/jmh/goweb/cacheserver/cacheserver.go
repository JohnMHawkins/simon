// cacheserver - http based cache server for reading and writing singleton caches used by multiple services
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
// 
//
//

package main

import (
	"net/http"
	"jmh/goweb/webber"
	"jmh/goweb/logger"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	//"os"
	"flag"
	"time"
	"fmt"
)

// uncomment to enable profiling on the /debug/pprof/ endpoint
import _ "net/http/pprof"

// TODO - make these configurable
var cacheDur time.Duration = 14*60*24*time.Minute
var cleanupInterval time.Duration = 14*60*24*time.Minute

// our caches
var CacheMap map[string]*cache.Cache

func getCache(dbName string, collName string, keyName string) *cache.Cache {
	mapKey := dbName + "." + collName + "." + keyName
	c, ok := CacheMap[mapKey]
	if (!ok) {
		// create it
		newC := cache.New(cacheDur, cleanupInterval)
		CacheMap[mapKey] = newC
		c = newC
	}
	return c
}



// This is our api/cache handler
//

/*
APIs

GET /<dbname>/<collname>/<keyname>/<keyvalue>
retrieves the entry stored for <keyvalue> under <dbname>/<collname>/<keyname> cache.  

POST /<dbname>/<collname>/<keyname>/<keyvalue>  -d {bson data}
stores the bson data entry under <keyvalue> in the <dbname>/<collname>/<keyname> cache

*/


type CacheHandler struct {
	basePath string
}

func NewCacheHandler(basePath string) *CacheHandler {
	f := new(CacheHandler)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h CacheHandler) Name() string {
	return "CacheHandler"
}

func (h CacheHandler) BasePath() string {
	return h.basePath
}

func (h CacheHandler) Handler ( w http.ResponseWriter, r *http.Request) { 
	webber.DispatchMethod(h, w, r);
}

func (h CacheHandler) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )
	if ( len(pathParts) == 4 ) {
		c := getCache(pathParts[0], pathParts[1], pathParts[2])
		if ( c != nil) {
			r, present := c.Get(pathParts[3])
			if ( present) {
				w.Header().Set("Content-type", "application/json")
				w.Write(r.([]byte))
			} else {
				http.Error(w, "", http.StatusNotFound)
			}
		} else {
			logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Cannot read cache for %s / %s / %s", pathParts[0], pathParts[1], pathParts[2]), nil)
			http.Error(w, "Cannot read Cache", http.StatusInternalServerError)
		}
	} else {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Invalid path specified: %s", apiPath), nil)
		http.Error(w, "Invalid path specified", http.StatusBadRequest)
	}
}

func (h CacheHandler) HandlePost (w http.ResponseWriter, r *http.Request) {

	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )

	if ( len(pathParts) == 4 ) {

		c := getCache(pathParts[0], pathParts[1], pathParts[2])
		if ( c != nil) {
			body, err := ioutil.ReadAll(r.Body)
			if (err == nil ) {
				c.Set(pathParts[3], body, cache.DefaultExpiration)	
				fmt.Fprintf(w, "%d bytes written", len(body))
			} else {
				logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Error reading POST body: %s", err.Error()), nil)
				http.Error(w, "Error ready data:" + err.Error(), http.StatusBadRequest)
			}
		} else {
			logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Cannot create cache for %s / %s / %s", pathParts[0], pathParts[1], pathParts[2]), nil)
			http.Error(w, "Cannot Create Cache", http.StatusInternalServerError)
		}
	} else {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Invalid path specified: %s", apiPath), nil)
		http.Error(w, "Invalid path specified", http.StatusBadRequest)
	}

}





// main func - we'll load our config, set up a logger,create an AppServer, and add our cache handler
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
	logger.StdLogger.LOG(logger.INFO, "", "cache-server starting up", nil)

	// initialize the map of caches
	CacheMap = make(map[string]*cache.Cache)

	// create an App Server
	as := webber.NewAppServer(config)

	//////////////////////////////////
	// create a couple of handlers

	// create our cache handler and assign it to <apibase>/cache
	ch := NewCacheHandler(config.ApiBase + "/cache")
	as.RegisterHandler(ch)

	// now start the server
	http.HandleFunc("/", as.Handler)
	http.ListenAndServe(config.Port, nil)
	
}