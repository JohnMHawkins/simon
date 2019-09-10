package webber

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)


/*
Explanation of the config entries:

Port :	this is the port the server will listen on.  Should be of the form ":8080".  Default is ":80"

WWWRoot : This is the path, from the working directory, to where files should be served from.  No files outside 
		this directory will be directly accessible through the file hander.  If "", then no file server will
		be created.  Defauls is "wwwroot"

DefaultFile :  The name of the file served up if there is a file request for the root.  Default is "index.html"

ApiBase : This is the base path for api calls handled by the server.  Any call outside of this path will be 
		considered a file request.  For example, if ApiBase was "rest", then calls to "www.foo.com/rest/bar" would
		generate a call to "bar".   The default is "api".

FileBase : This is the base url path for file calls.  This will be removed from the url path before looking for 
		files in the WWWRoot.  For example, if FileBase = "files" and WWWRoot = "wwwwroot", then a request for 
		"www.foo.com/files/img/treasuremap.png"  would look for the file "wwwroot/img/treasuremap.png"
	
*/
type ServerConfig struct {
	PublicUrl string 		
	Port string
	WWWRoot string			// path to wwwroot on server for fileserver.  If empty, no file server
	DefaultFile string		// default html file, e.g. index.html
	ApiBase string			// base url path to start of api, e.g. /api
	FileBase string			// base url path to files, e.g. /files

	DBPath string			// path to the db we should use

	SessionCollName string	// name of the collection used for session info in the DB

	// optional, used for app ID
	AppName string			// the name of the server app, for logging and id purposes
	AppVersion string		// the version of the server app
	
	// optional, used for AWS id
    AWSRegion string		// AWS region it is launched in
    AWSProfile string		// profile in  ~/.aws/credentials to use for auth to aWS
    LoggerFirehoseDeliveryStream string 	// name of the firehose delivery stream to use
}


// DefaultConfig creates a Server config with all defaults
//
func DefaultConfig () *ServerConfig {
	config := new(ServerConfig)

	config.Port = ":80"
	config.WWWRoot = "wwwroot"
	config.DefaultFile = "index.html"
	config.ApiBase = "api"
	config.FileBase = "/"

	return config
}

// LoadConfig reads a json config file, parses it, and fills in any defaults.  
//
// Parameters:
//	configFile string : path to file on the server to pull config information from.  if empty, will
//						use the default of "config.json"
//
// Returns:
//	*ServerConfig : the server config created
//
func LoadConfig (configFile string) *ServerConfig {
	config := DefaultConfig()

	// load configuration from file
	if (len(configFile) == 0) {
		configFile = "config.json"
	}

	// parse the json and return the ServerConfig contained in it
	configJson, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Failed to load app server config file " + configFile + ":", err)
	}

	jsonerr := json.Unmarshal(configJson, config )
	if jsonerr != nil {
		fmt.Println("Failed to parse file server config file : ", jsonerr)
	} else {

		// ensure FileBase ends with a slash
		if len(config.FileBase) > 0 && config.FileBase[len(config.FileBase)-1:] != "/" {
			config.FileBase += "/"
		}
		fmt.Println("root = " + config.WWWRoot)
		fmt.Println("apibase = " + config.ApiBase)
		fmt.Println("filebase = ", config.FileBase)
	}

	return config
}