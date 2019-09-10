// Taskdeck - code file for autoinc handling goweb server application stack
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
	"jmh/goweb/logger"
	"gopkg.in/mgo.v2"
	"time"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)


type AutoInc struct {
	Id string			`bson:"_id"`
	NextValue int64		`bson:"nextvalue"`
}


var autoIncColl *Collection


// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
//
// it creates an "autoinc" collection.  If you need to use that name for one of your own
// collections, you'll need to rename it here.
//
func CreateAutoIncDbCollection ( cDb *Db)  {
	autoIncColl = cDb.NewCollection("autoinc", "_id", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

// EnsureAutoIncrement makes sure there is an entry for the specified counter name
// and initializes it if needed.  Assumes there is a logger.StdLogger configured
//
// Params:
//	name :	the name of the autoincrement counter
//
// returns:
//	nothing
//
func EnsureAutoIncrement(name string) {
	var docTemplate AutoInc
	_, dbErr := autoIncColl.ReadNoCache(name, &docTemplate) 
	if (dbErr != nil ) {
		// doesn't exist yet, create it
		a := AutoInc{Id:name, NextValue:0}
		err := autoIncColl.WriteFast(name, a)
		if ( err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintln("Error creating autoinc for ", name, err), nil)
		}
	}

}

// GetNextAutoIncValue increments and returns the counter for the specified name
//
// Params:
//	name :	the name of the autoincrement counter to use
//
// returns:
//	int64 that is the next counter to use
//
func GetNextAutoIncValue(name string ) int64 {
	changeToApply := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"nextvalue": 1}},
		ReturnNew: true,
	}

	var result bson.M
	_, err :=autoIncColl.Dbc.Find(bson.M{"_id": name}).Apply(changeToApply, &result)
	if ( err != nil ) {
		return 0;
	}
	return result["nextvalue"].(int64)

}



