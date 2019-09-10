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
)

// AppServer is a webserver intended to support web applications by providing both a file server and an
// Api Server
type AppServer struct {
	Config *ServerConfig
	FileServerInst* FileServer
	Handlers map[string]WebHandler
}

// NewAppServer creates a new appserver with configuration information supplied by a ServerConfig object.  Will
// also create a FileServer that handles any paths not handled by the api server, if a wwwroot is specified
//
// Parameters:
//	config *ServerConfig : struct with configuration information for the server
//
// Returns:
//	*AppServer : the app server created
//
func NewAppServer(config *ServerConfig) *AppServer {
	f := new(AppServer)

	// set our config, then see if we need to create a FileServer
	f.Config = config
	if len(f.Config.WWWRoot) > 0 {
		f.FileServerInst = NewFileServer(f.Config.FileBase, f.Config.WWWRoot, f.Config.DefaultFile)
	}

	// initialize our map of handlers
	f.Handlers = make(map[string]WebHandler)
	return f
}

// GetName returns the name of the server (AppServers are always 
// named "AppServer")
//
// Parmaeters: 
//	none
//
// Returns:
//	string : Always just returns "AppServer"
//
func (h AppServer) Name() string {
	// AppServers don't have a unique name
	return "AppServer"
}

// BasePath returns the base path of the server (AppServers don't have base paths)
// they start at the root and have other handlers set underneath them
//
// Parmaeters: 
//	none
//
// Returns:
//	string : the api base
//
func (h AppServer) BasePath() string {
	return h.Config.ApiBase
}

// Handler - the base handler for the AppServer.  Our hptt server will call this directly
//
func (h AppServer) Handler (w http.ResponseWriter, r *http.Request) {
	wasHandled := false
	urlPath := r.URL.Path
	l := len(urlPath)
	if l > 0 {
		if urlPath[l-1:l] != "/" {
			// tack on a trailing slash
			urlPath = urlPath + "/"
		}
		fmt.Println("appServer handler path=", urlPath)
		
		for p := range h.Handlers {
			if len(urlPath) >= len(p) &&	urlPath[:len(p)] == p {
				wasHandled = true
				phf := h.Handlers[p]
				DispatchMethod(phf, w, r)
			} 
		}
	}
	if !wasHandled {
		// not specific handler, assume it's a file
		if h.FileServerInst != nil {
			DispatchMethod(h.FileServerInst, w, r)
		} else {
			http.Error(w, "File not Found", http.StatusNotFound)
		}
	}

}


// RegisterHandler will add a new handler to the appServer
//
// Parameters:
//	handler WebHandler : the handler to add, it should implement the WebHandler interface
//
// Returns:
//	none
//
func (h AppServer) RegisterHandler(handler WebHandler) {
	basePath := handler.BasePath()
	h.Handlers[basePath] = handler
}



func (h AppServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unknown path", http.StatusNotFound)

}

func (h AppServer) HandlePost (w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unknown path", http.StatusNotFound)

}
