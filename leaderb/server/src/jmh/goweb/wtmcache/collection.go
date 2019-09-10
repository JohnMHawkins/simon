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
	"fmt"
	"time"
	"reflect"
	"errors"
	"github.com/patrickmn/go-cache"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"jmh/goweb/logger"
)


// Struct that holds information about the cached collection
// 
type Collection struct {
	Parentdb *Db	// the db we are attached to
	Name     string			// the colleciton name
	KeyField string			// the name of the key 
	Dbc      *mgo.Collection	// the actuall collection
	C        *cache.Cache		// the cache for this collection
}


// internal method to create the collection, called from Db class.
//
//	Db - the parent db
//	collectionName string:	the name of the collection with the db
//	keyField string:		the name of the field that should be used as the cache key
//	cacheDur time.Duration	expiration time of the cache.  Typically specified by something like 10 * time.Minutes/\.
//							if this value is < 1 or NoExpiration, then the cache never expires
// 	cleanupInterval time.Duration	how frequently expired items are cleaned up in the cache.
//
func newCollection ( db *Db, collectionName string, keyField string, cacheDur time.Duration, cleanupInterval time.Duration) *Collection {
	wtcc := new(Collection)
	wtcc.Parentdb = db
	wtcc.Name = collectionName
	wtcc.KeyField = keyField
	wtcc.Dbc = wtcc.Parentdb.Db.C(collectionName)
	wtcc.C = cache.New(cacheDur, cleanupInterval)
	return wtcc	
}

// Read is a method that will read the document associated with the supplied key parameter, reading it from the cache if
// it is present, and fetching from the db (and writing to the cache) if it is not already in the cache.
//
// Parameters:
//	key string : the key value to use for the lookup
//	result interface{} :  defines the expected structure of the fetched document
//	 
// Returns:
//	interface{} :	the document read or fetched
//	error :			an error if there is a failure, or nil if it worked
//
// NOTE:  result *might* have the read result put into it, but you cannot rely on this.  You 
// should always use the return value and not the passed in parameter.
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcachepersist", "id", 5*time.Minute, 5*time.Minute)
//	var fetchDoc = DocTest{}
//	r, err2 := cColl.Read("123", &fetchDoc)
//	fetchDoc, ok := r.(DocTest)
//
func (w *Collection) Read (key string, result interface{}) (interface{}, error) {
	r, present := w.C.Get(key)
	if present {
		err := bson.Unmarshal(r.([]byte), result)
		return result, err
	} else {
		// fetch from db
		err := w.Dbc.Find(bson.M{w.KeyField: key}).One(result)
		if err == nil {
			// load it in the cache
			w.WriteFast(key, result)
		}
		return result, err
	}
	
}


// read directly from the db, bypassing the cache (does not set the cache either)
// Read is a method that will read the document associated with the supplied key parameter, always fetching from the db.
// Does not write it to the cache
//
// Parameters:
//	key string : the key value to use for the lookup
//	result interface{} :  defines the expected structure of the fetched document
//	 
// Returns:
//	interface{} :	the document read or fetched
//	error :			an error if there is a failure, or nil if it worked
//
// NOTE:  result *might* have the read result put into it, but you cannot rely on this.  You 
// should always use the return value and not the passed in parameter.
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcachepersist", "id", 5*time.Minute, 5*time.Minute)
//	var fetchDoc = DocTest{}
//	r, err2 := cColl.ReadNoCache("123", &fetchDoc)
//	fetchDoc, ok := r.(DocTest)
//
func (w *Collection) ReadNoCache (key string, result interface{}) (interface{}, error) {
	// fetch from db
	err := w.Dbc.Find(bson.M{w.KeyField: key}).One(result)
	return result, err
	
}

// WriteFast
// Writes the supplied document to the db with the specified key, and also places in the cache under the key
// Is slightly faster than Write, but requires the caller to specify the correct value of the key.
//
// Parameters:
//	key string : the key to store the document under
//	document interface{} : the document to write
//
// Returns:
//	error : nil if no error
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
//	err := cColl.Write("xyz", doc)
//
func (w *Collection) WriteFast(key string, document interface{}) error {

	_, err := w.Dbc.Upsert(bson.M{w.KeyField: key}, document)
	if err == nil {
		d, err2 := bson.Marshal(document)
		w.C.Set(key, d, cache.DefaultExpiration)
		err = err2
	}
	return err
}



// Write
// Writes the supplied document to the db with the specified key, and also places in the cache under the key
//
// Parameters:
//	document interface{} : the document to write
//
// Returns:
//	error : nil if no error
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
//	err := cColl.Write(doc)
//
func (w *Collection) Write(document interface{}) error {
	// fetch json key via reflection
	var key string
	v := reflect.ValueOf(document)
	for i := 0; i < v.Type().NumField(); i++ {
		if v.Type().Field(i).Tag.Get("json") == w.KeyField {
			key = v.Field(i).String()
			break
		}
	}

	if key == "" {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("DB write called but keyfield %s not found in supplied document", w.KeyField), nil)
		return errors.New("No key field " + w.KeyField + " found in document " )
	}
	return w.WriteFast(key, document)
}

// Deletes the entry identified by key from both the cache and the underlying collection
//
// Parameters:
//	key string : the key of the document to delete
//
// Returns:
//	error : nil if no error
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//  err := cColl.Delete("bbb")
//
func (w *Collection) Delete (key string) error {
	// remove from the cache
	w.C.Delete(key)

	// remove from the db
	err := w.Dbc.Remove(bson.M{w.KeyField: key})
	return err
}

// Query reads directly from the database, using a mongo document query supplied as a parameter.
//
// Parameters:
//	query interface{} :	mongo query object, as specified by https://godoc.org/gopkg.in/mgo.v2#Query
//	results interface{} : a slice of the form []<doctype> to receive the queried results
//
// Returns:
//	error : nil if no error
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//	var results []DocTest
//	var query = bson.M{"foo" : "xxx"}
//	err := cColl.Query(query, &results)
//	for i, v := range results {
//		... i is the index, v is a DocTest object ...
//	}
//
func (w *Collection) Query ( query interface{}, results interface{} ) error {
	return w.Dbc.Find(query).All(results)
}


// Queryandsort reads directly from the database, using a mongo document query supplied as a parameter, and sorts
// the results by the field defined by sortBy.
//
// Parameters:
//	query interface{} :	mongo query object, as specified by https://godoc.org/gopkg.in/mgo.v2#Query
//	results interface{} : a slice of the form []<doctype> to receive the queried results
//	sortBy string : a string to sort by, preceeded by a minus sign to sort in reverse
//
// Returns:
//	error : nil if no error
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//	var results []DocTest
//	var query = bson.M{"foo" : "xxx"}
//	sortBy := "-name"
//	err := cColl.QueryAndSort(query, &results, sortBy)
//	for i, v := range results {
//		... i is the index, v is a DocTest object ...
//	}
//
func (w *Collection) QueryAndSort ( query interface{}, results interface{}, sortBy string ) error {
	return w.Dbc.Find(query).Sort(sortBy).All(results)
}

// RawQuery reads directly from the database, using a mongo document query supplied as a parameter and returning
// the mgo Query object.
//
// Parameters:
//	query interface{} :	mongo query object, as specified by https://godoc.org/gopkg.in/mgo.v2#Query
//	
// Returns:
//	*Query : pointer to a mgo query object as specified by https://godoc.org/gopkg.in/mgo.v2#Query
//
// Example:
//	cDb :=  NewDb(session, "test")
//	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
//	var query = bson.M{"foo" : "xxx"}
//	result, err := cColl.RawQuery(query)
//	... now result can be manipulated as defined at https://godoc.org/gopkg.in/mgo.v2#Query
//
func (w *Collection) RawQuery ( query interface{} ) *mgo.Query {
	return w.Dbc.Find(query) 
}
