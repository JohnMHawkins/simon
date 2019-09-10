webber
======

``webber``` is a Golang library that implements several common webserver variations.

It provides a simple file server (filesvr), and easily extensible app server (appsvr) implementations.  Both are based on the Golang net/http library, but provide simple integration points to create a scalable web server with appropriate logging functionality.

## Design Notes

Why another Mux?  Why not just use an existing one?  The main purpose here was to create an easily versionable
set of apis for the app server.  So you can declare a type that implements WebHandler and assign it to /foo/v1  
(or /v1/foo if you'd rather version that way).  Now you have an easily contained api set.  

It doesn't support the kind of mix-n-match that other Mux's do, you have to pick a path and bundle all the handlers underneath it, but that's intentional - it keeps your code logically organized in the WebHandler implementations.

It only supports GET and POST methods.  An early version supported all the methods, but it became cumbersome declaring all the "not implementeds" for PUT, PATCH, TRACE etc, and I generally don't use these, so I just cut them out at the top.  If needed, they can always be put back, if you want to do a very formal REST api for instance.


## Usage

See jmh/goweb/webbertut for an example app server

## ToDo



## License

Webber is covered by the MIT Licesne.  

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


