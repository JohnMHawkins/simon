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
	"fmt"
	"time"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)


type FirehoseLogger struct {
	AWSSession *session.Session
	FirehoseClient *firehose.Firehose
	DeliveryStreamName string
	AlsoToStdout bool  // if true, also outputs to stdout with fmt.Println
	App AppInfo
}



// NewFirehoseLogger creates an AWSSession and then a Kinesis Firehose Client.
//
// It requires that a Deliverystream has already been created, and does not create one itself.  The
// reason for this is that AWS allows a limited number of deliverystreams by default, so it is potentially
// dangerous to allow programatic creation - you could accidentally DDOS yourself.  
//
// parameters:
//	app : the AppInfo structure describing the service
//	region : the AWS region to create everything in.  
//	profileName : The profile stored in ~/.aws/credentials that provides the creds for accessing AWS
//	deliveryStream : The name of the firehose delivery stream to use for the log
//
// Returns:
//	a pointer to an FirehoseLogger struct
//
func NewFirehoseLogger(app AppInfo, region string, profileName string, deliveryStreamName string) *FirehoseLogger {
	fmt.Println("FireHoseLogger creating new AWS Session for region ", region, " and profile ", profileName)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: profileName, 
		Config: aws.Config{Region: aws.String(region)},
	}))
	return NewFirehoseLoggerFromSession(app, sess, deliveryStreamName)

}

// NewFirehoseLoggerFromSession uses an existing AWSSession and creates a Kinesis Firehose Client.
//
// It requires that a Deliverystream has already been created, and does not create one itself.  The
// reason for this is that AWS allows a limited number of deliverystreams by default, so it is potentially
// dangerous to allow programatic creation - you could accidentally DDOS yourself.  
//
// parameters:
//	app : the AppInfo structure describing the service
//	session : the AWS.Session object already created.  
//	deliveryStream : The name of the firehose delivery stream to use for the log
//
// Returns:
//	a pointer to an FirehoseLogger struct
//
func NewFirehoseLoggerFromSession(app AppInfo, sess *session.Session, deliveryStreamName string) *FirehoseLogger {
	fmt.Println("NewAWSLogger being created for region ", *(sess.Config.Region), " on deliverystream ", deliveryStreamName)
	f := new(FirehoseLogger)
	f.AWSSession = sess
	f.FirehoseClient = firehose.New(f.AWSSession)
	f.DeliveryStreamName = deliveryStreamName
	f.AlsoToStdout = false
	f.App = app

	// check for existing delivery streams.  
	dl, err := f.FirehoseClient.ListDeliveryStreams(nil)
	if ( err == nil ) {
		// look for the stream we're supposed to use:
		for _, n := range dl.DeliveryStreamNames {
			if *n == deliveryStreamName {
				// found it
				fmt.Println("...delivery stream found")
				return f
			}
		}
		// if we got here, we didn't find it, write to stdout 
		fmt.Println("...ERROR delivery stream ", deliveryStreamName, " not found, streams found: ", dl)
		// we don't error out however, as perhaps the strem will show up?

	} else {
		fmt.Println("...error getting delivery streams :", err)
	}
	return f 

}

func (l *FirehoseLogger) StdOutOn(alsoToStdOut bool) {
	l.AlsoToStdout = alsoToStdOut
}

func (l *FirehoseLogger) LOG (level LogLevel, correlationid string, msg string, keys map[string]string ) {

	entry := LogEntry{Level: level, Timestamp:time.Now().UnixNano()/1000000, CorrelationId:correlationid, Message:msg, App:l.App, Keys:keys}
	data, _ := json.Marshal(entry)

	var r firehose.Record
	r.SetData([]byte(data))

	var p firehose.PutRecordInput
	p.SetDeliveryStreamName(l.DeliveryStreamName)
	p.SetRecord(&r)

	_, err := l.FirehoseClient.PutRecord(&p)
	if ( l.AlsoToStdout) {
		fmt.Println(entry)
	}

	if ( err != nil ) {
		// what do we do now?  Log an error ha ha?
		fmt.Println("Failed to PutRecord to firehose client: ", err)
	}

}