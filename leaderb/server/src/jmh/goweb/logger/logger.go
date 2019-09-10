// logger - GoWeb Logging helper package
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

package logger

import (
	"crypto/rand"
	"encoding/hex"
)

var StdLogger Logger

// TODO
// 
//


type LogLevel string
const (
	CRITICAL LogLevel = "CRITICAL"
	ERROR = "ERROR"
	WARN = "WARN"
	INFO = "INFO"
)

// the AppInfo struct is used for providing information about the app/service that is doing
// the logging. 
type AppInfo struct {
	Name string `json:"name"`			// name of the app/service, e.g. "Account-Service"
	Version string `json:"version"`		// version of the deployed service, e.g. "1.3.11"
	Cluster string `json:"cluster"`		// cluster or virtual environment it is deployed in, e.g. "PROD-West"
	Instance string `json:"instance"`	// instance of the service if there are more than one, e.g. "3"
}

// LOG :  instructs the logger to write a log.  
//			level is one of the LogLevel enums above
//			correlationid is meant to be used to correlate related log entries
//			msg is the message to be logged
//			keys is an optional map of additional variables that can be used for searching
//
//	StdOutOne : controls whether the logger will echo the log data to stdout on the machine (default is Off)
//		alsoToStdOut :  if true, will echo to stdout.  if false, will not echo to stdout
type Logger interface {
	LOG(level LogLevel, correlationid string, msg string, keys map[string]string)
	StdOutOn(alsoToStdOut bool)
}

// this is the data that will be sent to the remote log collector
type LogEntry struct {
	Level LogLevel  `json:"level"`
	Timestamp int64 `json:"@timestamp"`
	CorrelationId string `json:"correlation_id"`
	App AppInfo `json:"appinfo"`			
	Message string  `json:"message"`
	Keys map[string]string `json:"keys"`
}

///////////////////////////////////////////////////
// Helper funcs
//

// returns a randomly generated 16 digit hex value
func GenerateCorrelationId() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return ""
	}
	
	return hex.EncodeToString(u)
}

