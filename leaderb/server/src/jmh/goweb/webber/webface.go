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
	"strings"
	"encoding/json"
	"math/rand"
	"jmh/goweb/logger"
)

const CORRELATION_ID_HEADER = "correlation-id"

// WebHandler is the base interface for all web server types.
// It only supports the common methods GET and POST.  
//
type WebHandler interface {
	HandleGet(w http.ResponseWriter, r *http.Request)
	HandlePost(w http.ResponseWriter, r *http.Request)
	BasePath() string
	Name() string
}

func getCorrelationId ( r *http.Request) (string) {
	h := r.Header.Get(CORRELATION_ID_HEADER)
	if (len(h) == 0 ) {
		// create and add one
		h = fmt.Sprintf("%d", rand.Uint64())
		r.Header.Add(CORRELATION_ID_HEADER, h)
	}
	return h
}

func GetCorrelationId ( r *http.Request) (string) {
	h := r.Header.Get(CORRELATION_ID_HEADER)
	return h
}

// formats the request for useful log information
func formatReqForLog(r *http.Request) (string) {
	s := fmt.Sprintf("%s %s Headers:{", r.Method, r.URL)
	for k, v := range r.Header {
		s += k + ":[" 
		for _, vv := range v { s += vv + ","}
		s += "], "
	}
	s += "}"
	return s
}


// root dispatcher called by all WebHandlers to determine Method and dispatch to appropriate case handler
func DispatchMethod(h WebHandler, w http.ResponseWriter, r *http.Request) {

	logger.StdLogger.LOG(logger.INFO, getCorrelationId(r), fmt.Sprintf("Inbound request %s", formatReqForLog(r)), nil)
	
	switch {
		case r.Method == "GET":
			h.HandleGet(w, r)
		case r.Method == "POST":
			h.HandlePost(w, r)

		case r.Method == "HEAD":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "TRACE":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "OPTIONS":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "PUT":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "PATCH":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "DELETE":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		case r.Method == "CONNECT":
			logger.StdLogger.LOG(logger.WARN, "", fmt.Sprintf("Unsupported method %s called", r.Method), nil)
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}

}

// ParsePathAndQueryFlat parses the path into an array of path elements (delimited by /) and 
// a map[string]string of query string parameters.  The difference between ParsePathAndQuery and 
// ParsePathAndQueryFlat is the flat version only returns a single value for each query param, while
// the non-flat returns a list for each (even if there is only one item in the list)
//
// Parameters:
//	r		: http.request, used to fetch query and post params
//	path 	: string that is the path as modified by the caller
//	pathVars: optional, map of indicies into the path that should be extracted
//			  from the pathParts and puts it into the query params
//
// Returns:
//	two values, a []string of path parts and a map[string]string of query/post
//		params.  
//
// Example:
//	if the path is "foo/bar?x=1&y=2" then this will return
//	["foo","bar"] , ["x" = ["1"], "y" = ["2"]}
//
//
func ParsePathAndQueryFlat (r *http.Request, path string, pathVars map[int]string) ([]string, map[string]string) {

	var pathParts []string
	qParams := r.URL.Query()
	queryParams := make(map[string]string)
	for k, v := range qParams {
		queryParams[k] = v[0]
	}

	if len(path) > 0 {
		pathParts = strings.Split(path, "/")
	}

	// pull out any vars
	i := 0	// this keeps track of the number of vars we've pulled out so we can
			// offset the now-smaller remaining index since we extract as we go
	for k, v := range pathVars {
		if len(pathParts) > (k-i) {
			queryParams[v] = pathParts[k-i]
			pathParts = append(pathParts[:k-i], pathParts[k-i+1:]...)
			i++
		}
	}


	return pathParts, queryParams

}


// ParsePathAndQuery parses the path into an array of path elements (delimited by /) and 
// a map[string][]string of query string parameters.
//
// Parameters:
//	r		: http.request, used to fetch query and post params
//	path 	: string that is the path as modified by the caller
//	pathVars: optional, map of indicies into the path that should be extracted
//			  from the pathParts and puts it into the query params
//
// Returns:
//	two values, a []string of path parts and a map[string][]string of query/post
//		params.  
//
// Example:
//	if the path is "foo/bar?x=1&y=2" then this will return
//	["foo","bar"] , ["x" = ["1"], "y" = ["2"]}
//
//
func ParsePathAndQuery (r *http.Request, path string, pathVars map[int]string) ([]string, map[string][]string) {

	var pathParts []string
	queryParams := r.URL.Query()

	if len(path) > 0 {
		pathParts = strings.Split(path, "/")
	}

	// pull out any vars
	i := 0	// this keeps track of the number of vars we've pulled out so we can
			// offset the now-smaller remaining index since we extract as we go
	for k, v := range pathVars {
		if len(pathParts) > (k-i) {
			queryParams.Add(v, pathParts[k-i])
			pathParts = append(pathParts[:k-i], pathParts[k-i+1:]...)
			i++
		}
	}

	return pathParts, queryParams

}

// ReturnJson returns the supplied struct marshalled to json, setting the 
// appropriate http headers.
//
// Parameters:
//	w :	the responseWriter to use
//	jsonDoc : a structure that can be marshalled as json
//
// Returns:
//	an error if the struct could not be marshalled
//
func ReturnJson ( w http.ResponseWriter, jsonDoc interface{} ) error  {
	jsonStr, jsonerr := json.Marshal(jsonDoc)
	if jsonerr == nil {
		w.Header().Set("Content-type", "application/json")
		w.Write(jsonStr)
	} else {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Error %s returning json %s", jsonerr.Error(), jsonStr), nil)
		http.Error(w, jsonerr.Error(), http.StatusInternalServerError)
	}
	return jsonerr

}

