logger
=========

``logger``` is a Golang library that implements easy logging to a network log collector. It is meant to be used with webber and other jmh/goweb libraries

Currently implement destinations are:
-FirehoseLogger:  AWS ElasticSearch via an AWS Firehose delivery stream



## Design Notes

The AWS FirehoseLogger requires that a Firehose delivery stream be created before the logger is instantiated, and this deliverystream will define the Index and Type of the data delivered to ElasticSearch.  The AWS SDK does allow for programatic creation of a Firehose delivery stream, and this would allow the application to specify its own Index and Type and thought was given to implementing this for the FirehoseLogger.  However, by default AWS limits the number of Firehose delivery streams an account can create per region.  Because delivery streams are a potentially scarce resource, I decided not to enable programatic creation of them, and instead require they be set up manually outside the codebase.


## Usage

Step 1) create a logger upon initialization of your service:

    import {
        "jmh/goweb/logger"
    }

    // fill out the AppInfo struct so the logger knows who it is writting logs for:
	app := logger.AppInfo{Name:"test-server", Version:"1.13.11", Instance:"1",Cluster:"PROD-east"}

    // initialize a firehose logger for the us-east-1 region using the "default" AWS profile, and 
    // connect the logger to the firehose delivery stream named "test-firehose1-useast-1"
	fhLogger := logger.NewFirehoseLogger(app, "us-east-1", "default", "test-firehose1-useast-1")

    // alternately, if you already have an AWS Session in the correct region from another operation,
    // you can use it to create the logger
	fhLogger := logger.NewFirehoseLoggerFromSession(app, existingSession, "test-firehose1-useast-1")


    // if you want all your log output to ALSO go to the stdout on the machine, call this:
    fhLogger.StdOutOn(true)

    // and of course to stop output from also going to stdout, call the same function with false
    fhLogger.StdOutOn(false)


Step 2) whenever there is a message to log, call the LOG function on the logger:

    // generate a correlation id to let you find related entries in the logs.  For example, you might do this
    // on first receiving a call at your outer API and make sure every function and service involved in handling 
    // the call has the correlationId available to include with any logs.  Then you will easily be able to 
    // local all log entries related to handling that call
   	correlationId := logger.GenerateCorrelationId()

    // if you have additional fields you want to have available, put them in a keys map[string]string
    keys := map[string]string{"abc":"123","xyz":"kumquot"}

    // send a Loglevel CRITICAL log message
    fhLogger.Log(logger.CRITICAL, correlationId, "Your text log message goes here", keys)

    // send a Loglevel ERROR log message
    fhLogger.Log(logger.ERROR, correlationId, "Your text log message goes here", keys)

    // send a Loglevel WARN log message
    fhLogger.Log(logger.WARN, correlationId, "Your text log message goes here", keys)

    // send a Loglevel INFO log message
    fhLogger.Log(logger.INFO, correlationId, "Your text log message goes here", keys)


## ToDo

-Fix comments to be more gopherish

## License

logger is covered by the MIT Licesne.  

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


