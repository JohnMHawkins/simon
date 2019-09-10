package webber

import (
	"fmt"
	"net/http"
	"io/ioutil"	
	"path"
	"path/filepath"
)

type FileServer struct {
	Config ServerConfig
	WWWRootPath string 	// file path to where the www root directory is on the server.  Files served from here 
	basePath string     // url path to where we start serving files from.  e.g. "/files", but usually "/"
	DefaultFile string	// name of the default file served up for the root.  Usually Index.html
}

func NewFileServerWithConfig(basePathToHere string, config ServerConfig) *FileServer {
	f := new(FileServer)

	f.basePath = basePathToHere
	f.Config = config

	return f
}

func NewFileServer(basePathToHere string, wwwRoot string, defaultFile string) *FileServer {
	f := new(FileServer)

	f.basePath = basePathToHere 
	f.WWWRootPath = wwwRoot

	// if they specified a default file, use it, otherwise use index.html as the default
	if len(defaultFile) > 0 {
		f.DefaultFile = defaultFile
	} else {
		f.DefaultFile = "index.html"
	}
	
	return f
}


func (h FileServer) Name() string {
	return "FileServer"
}

func (h FileServer) BasePath() string {
	return h.basePath;
}

func AddMimeType(w http.ResponseWriter, filename string) {
	ext := filepath.Ext(filename)
	fmt.Println("AddMimetype called for ", ext)
	switch(ext) {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".htm":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	//case ".ttf":
	//	w.Header().Set("Content-Type", "text/html")
	//case ".otf":
	//	w.Header().Set("Content-Type", "text/html")
	case ".png":
		w.Header().Set("Content-Type", "image/png")

	}

}

func (h FileServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	ourPath := r.URL.Path[len(h.basePath):]
	fmt.Println("fileserver handleGet of ", ourPath)
	if len(ourPath) == 0 {
		ourPath = h.DefaultFile
	}
	filename := path.Join(h.WWWRootPath,  ourPath)
	fmt.Println("...fileserver handleGet looking for ", filename)
	body, err := ioutil.ReadFile(filename)
	if err == nil {
		AddMimeType(w, filename)
		w.Write(body)
	} else {
		// try adding html
		filename = filename + ".html"
		body, err := ioutil.ReadFile(filename)
		if err == nil {
			AddMimeType(w, filename)
			w.Write(body)
		} else {
			http.Error(w, "File not Found", http.StatusNotFound)
		}

	}

}


func (h FileServer) HandlePost (w http.ResponseWriter, r *http.Request) {

	// todo:  add uploader support

}

