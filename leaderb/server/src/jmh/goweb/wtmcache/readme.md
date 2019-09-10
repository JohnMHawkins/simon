WTMCache 
========

wtmcache stands for Write-Through Mongo Cache.  It is a golang write-through cache to a mongo database.  It uses caching code from https://patrickmn.com/projects/go-cache/ and the mgo.v2 (https://gopkg.in/mgo.v2) package as the mongo driver.

Operation is based on two main classes:

*Db* : A cached connection to a mongo database

*Collection* : A write-through cached connection to mongo collection.  

## Installing dependencies:
go get github.com/patrickmn/go-cache

go get gopkg.in/mgo.v2

## TODO

-Allow use of cacheserver instead of inherent cache


## Usage

Establish a session with the mongo server by:

    session, err := mgo.Dial("127.0.0.1:27017")

Then use the session to obtain a Db object:

    cDb := NewDb(session, "<dbname>")

And a Collection object:

    cColl := cDb.NewCollection("<collectionName", "<keyName>", <cacheDuration>, <cleanupTime> )
    
where 
    keyName is the name of the field from the collection documents you want to use as the cache key
    and the two time values at the end are something like 5*time.Minute


You can read from the collection by:

    var fetchDoc = <yourDocStructDTO>{}
	var ok bool
	r, err2 := cColl.Read("<keyValue>", &fetchDoc)
	if err2 == nil {
    	fetchDoc, ok = r.(<yourDocStructDTO>)
    }

And you can write to the collection by:

    doc := <yourDocStructDTO>{ID: "xyz", Name: "Joe", Foo: "bar"}
	err := cColl.Write("xyz", doc)

    

## License

wtmcache is covered by the MIT Licesne.  

Copyright (c) 2018 - John M. Hawkins <jmhawkins@msn.com>

All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and 
associated documentation files (the "Software"), to deal in the Software without restriction, 
including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, 
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial 
portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

    