# Starter Kit repository

## Overview
This is a basic jmh/goweb app server that includes basic account managmeent (creation, pword recovery, etc) using a cache-fronted mongo database.

Use this as the start of anything that requires user accounts, renaming files as needed and adding to the User struct appropriate elemetns for the service as needed.  It contains a www root directory and functions as a web server and an api server both - the webserver behavior can be turned off if it is to be run as a pure service.  It also contains an example handler to serve as an example of creating an api handler.

## Setting up
Remember to set goroot=[root path where you cloned this repo]
This project requires the following other github libraries be installed into the src directory (run these commands from the root):

* go get github.com/aws/aws-sdk-go
* go get github.com/patrickmn/go-cache
* go get gopkg.in/mgo.v2
* go get -u golang.org/x/crypto/...

## install goweb packages:

* go install jmh\goweb\webber
* go install jmh\goweb\logger
* go install jmh\goweb\wtmcache


you can install cacheserver.exe to the bin directory with:
go install jmh\goweb\cacheserver

or you can build it in the src directory directly and run it from there with the supplied config.json file (update any firehose settings you wish) But the current 
version of RAAS dosen't use cacheserver but rather writes directly to it's own cache-fronted mongo db.  When multiple instances are running, we'll partition by 
user cohort, so the same user will always use the same taskdeck instance so the cache will be fine. 

## Using the Starter Kit

- rename src/jmh/app and src/jmh/app/app.go to the name of your service.  
- update config.json for the correct app name
- update config.json for the right LoggerFirehoseDeliveryStream
- replace example.go (and it's use in app.go) with appropriate api handlers for your app
- change "appusers" in user.go CreateUserDbCollection to the appropriate name for your users (e.g. foousers)
- change "users" in the wtmcache.EnsureAutoIncrement() call in app.go main() and in wtmcache.GetNextAutoIncValue in users.go CreateNewUser to somthing unique to your app. 
- change the message in the logger.StdLogger.LOG call in main() to indicate your app's name
- change the db name in wtmcache.NewDb to something unique for your app

NOTE:  sending email for a password reset reminder requires validating the source email with AWS.  Testing can be done with test accounts to test accounts, but the AWS Simple Email Service (SES) will company validation before it can be used for real.  See https://aws.amazon.com/ses/faqs/ for details

{appname}.exe will launch the app and use the default config.json from the same directory. 

### launch options:
- {appname} -config foo.json   :  uses foo.json as the config file instead of the default config.json
- {appname} -instance foo  : gives the instance the name "foo" for use in any orchestration and logging
- {appname} -cluster foo  : gives the instance a cluster named "foo" for use in orchestration and logging
- {appname} -awsregion xxx : overrides the region setting from the config file
- {appname} -awsprofile xxx : overrides the profile setting from the config file
- {appname} -dbpath xxx : overrides the dbpath setting from the config file

### config file options
- Port : ":8090"  : will listen on port 8090
- "WWWRoot" : ""  : relative path from where exe runs to find the wwwroot directory for file serving
- "DefaultFile":"index.html"  : file name from wwwroot to serve as the default is just http://foo.bar.com/ is specified
- "ApiBase":"api"  : base url path for the rest or rpc calls.  e.g. this says the api starts at http://foo.bar.com/api/xxx
- "FileBase" : "/" : base url path for the fileserver
- "DBPath" : "127.0.0.1:27017"  : url to find the mongodb at.
- "SessionCollName" : "sessioncache" : what the collection in the db that stores session info is called
- "AppName" : "starterkit"  : the name of the app for orchestration and logging
- "AppVersion" : "0.1.1"  : the version of the app for orchestration and logging
- "AWSRegion" : "us-east-1"  : what AWS region this instance should use resources from
- "AWSProfile" : "default"  : AWS profile from the users .aws/credentials file to use (specifies aws keys etc)
- "LoggerFirehoseDeliveryStream" : "hawk-test-firehose1-useast-1"  : name of the AWS Kenisis Firehose stream to send logging to
- "PublicUrl" : "localhost:8080"  : the base url to use when constructing replies that need to include links


### starter kit APIs

Auth:
- POST /api/auth/register : registers for a new account 
- POST /api/auth/login : login, creating session and setting session key in response header
- POST /api/auth/requestreset : requests a reset code be sent to the registered user's email
- POST /api/auth/resetpassword : use the reset code to reset the password
- POST /api/auth/updateprofile : allows the user to update their profile
- GET /api/auth/logout : logout and end the session

Inventory:
- GET /api/inventory/get  : gets the inventory for the logged in user, optionally item={itemcode}, and tags={taglist}
        returns application/json { itemname : { count : x, properties: { k:v, k:v}}}
- GET /api/inventory/add  : adds a quantity of an item to a user's inventory, ?user={username}&item={itemcode}&qty={qty} 
- GET /api/inventory/remove  : removes a quantity of an item to a user's inventory, ?user={username}&item={itemcode}&qty={qty} 

Recipe:
- GET /api/recipe/get : gets one or more recipes, ?recipe={recipecode} or tags={tags}
- GET /api/recipe/make : executes a recipe for a user (removes inputs, adds output) ?user={username}&recipe={recipecode}
- GET /api/recipe/check : returns if a user has the requirements to make a recipe ?user={username}&recipe={recipecode}

- POST /api/recipe/new : creates a new recipe (TBD)



## NYI features
 
- detect lost connection to firehose

- Real web pages


