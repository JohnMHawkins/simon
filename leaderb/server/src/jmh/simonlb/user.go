// Taskdeck - code file for user handling in the taskdeck server application
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


package main

import (
	"net/http"
	"jmh/goweb/webber"
	"jmh/goweb/logger"
	"jmh/goweb/wtmcache"
	"time"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)


type User struct {
	Id int64			`bson:"_id" json:"id"` 
	Username string		`json:"username"`
	Password string		`json:"password"`
	Email string		`json:"email"`
	AccountType string	`json:"account_type"`
	CreateTime int64	`json:"create_time"`	// Unix time, number of seconds elapsed sicne Jan1 1970
	LastLogin int64		`json:"last_login"`		// Unix time, number of seconds elapsed sicne Jan1 1970
	ResetCode string	`json:"reset_code"`
	ResetTime int64		`json:"reset_time"`		// Unix time when reset was issued	

}

var userColl *wtmcache.Collection


// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
func CreateUserDbCollection ( cDb *wtmcache.Db)  {
	userColl = cDb.NewCollection("simonusers", "username", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

func CreateNewUser(r *http.Request, username string, pwhash string, email string, orgId int64, accType string) (*User, *AppError) {
	logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Creating New User: %s - %s ", username, email), nil)

	// check that an account with that username doesn't already exist
	var docTemplate User
	_, dbErr := userColl.Read(username, &docTemplate) 
	if ( dbErr == nil ) {
		// already exists
		return nil, MakeError(ERR_ACCOUNT_ALREADY_EXISTS, "Account already exists")
	}

	// check that an account with that email doesn't already esits
	dbErr = userColl.Dbc.Find(bson.M{"email": email}).One(&docTemplate) 
	if (dbErr == nil ) {
		// email already exists
		return nil, MakeError(ERR_EMAIL_ALREADY_EXISTS, "An account with that email already exists")
	}

	id := wtmcache.GetNextAutoIncValue("simonusers")

	u := new(User)
	u.Id = id
	u.Username = username
	u.Password = pwhash
	u.Email = email
	u.AccountType = accType
	u.CreateTime = time.Now().Unix()
	u.LastLogin = u.CreateTime
	u.ResetTime = 0

	userColl.Write(*u)

	return u, nil
}

func GetUser(username string) (*User, *AppError) {
	var docTemplate User
	u, err := userColl.Read(username, &docTemplate) 
	if (err != nil ) {
		return nil, MakeError(ERR_READ_USER_FAILURE, err.Error())
	}
	user, _ := u.(*User)
	return user, nil
}

func GetUserByEmail(email string) *User {
	var docTemplate User
	err := userColl.Dbc.Find(bson.M{"email": email}).One(&docTemplate) 
	if (err != nil ) {
		logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Error fetch email: %s", email), nil)
		return nil
	}
	return &docTemplate
}

func UpdateUser(u *User) error {
	err := userColl.Write(*u)
	if err != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("Error %s writing user: %d", err.Error(), u.Id), nil)
	}
	return err
}


