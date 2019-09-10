cacheserver
=========

``cacheserver``` is a Golang http server providing unified cache service.

## TODO 
 -allow configuration of cache duration parameters
 -allow storage limits and cache eviction

## Usage

to put a value to the cache, call:

POST /<dbname>/<collname>/<keyname>/<keyvalue>  -d {data}

This stores the data entry under <keyvalue> in the <dbname>/<collname>/<keyname> cache

example:  curl --header "Content-Type: application/json" --request POST --data '{"id":"1", "foo":"bar"} http:localhost:8090/api/cache/test/foos/id/1

will store {"id":"1","foo":"bar"}   is a cache set up for db = test and collection=foos, indexed on id


To retrive a value, call:

GET /<dbname>/<collname>/<keyname>/<keyvalue>

retrieves the value stored under <keyvalue> in the <dbname>/<collname>/<keyname> cache

example:  curl --header "Content-Type: application/json" --request GET http:localhost:8090/api/cache/test/foos/id/1

will return the value stored above

Content-type is always application/json in both directons

## License

cacheserver is covered by the MIT Licesne.  

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


