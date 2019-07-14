// Starterkit - code file for auth handling in the taskdeck server application
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
	"time"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"

)


// This is our api/auth handler, which will do a simple login and save session auth info
//

type AuthServer struct {
	basePath string
}

func NewAuthServer(basePath string) *AuthServer {
	f := new(AuthServer)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h AuthServer) Name() string {
	return "AuthServer"
}

func (h AuthServer) BasePath() string {
	return h.basePath
}

func (h AuthServer) Handler ( w http.ResponseWriter, r *http.Request) { 
	webber.DispatchMethod(h, w, r);
}


////////////////////////////////////////
//
// Get commands
//
// /info
// /logout
// 

type AuthInfoResponse struct {
	Username string		`json:"username"`
	Email string		`json:"email"`
	AccountType string	`json:"account_type"`
}


func (h AuthServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathVars := map[int]string{1:"a"}
	logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("AuthServer GET handler called for %s", apiPath), nil)
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, pathVars )

	switch pathParts[0] {
	case "info":
		// get the session if one exists
		// NOTE: this only works if we're a single instance.  If we are multiple instances,
		// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
		//var session UserSessionData
		session := UserSessionData{}
		bHasSession, _ := webber.GetSession(r, &session)
		if ( bHasSession ) {
			u, _ := GetUser(session.Username)
			if ( u != nil ) {
				resp := AuthInfoResponse{Username:u.Username, Email:u.Email, AccountType:u.AccountType}
				webber.ReturnJson(w,resp)
				return
			}
		} else {
			http.Error(w, "NoSession", http.StatusUnauthorized)
		}
	case "logout":
		// in this case, we would clear the session entry in the db and the header itself
		webber.ClearSession(w)
		fmt.Fprintf(w, "ok")
	
	}
	
}

//////////////////////////////
//
// POST commands
//
// /register
// /login
// requestreset
// resetpassword
// update profile
// 

// Webhandler - Registers a new user.  
// Request params:  
// 	username - the username for the new use
//	email - email to use
//	password - password
//	confirm - test to confirm matching password
// Response:
//	200 OK if success
//	400 Bad Request if:
//		username too short:	username must be at least 6 chars
//		password too short: password must be at least 6 chars
//		password doesn't match
//		ERR_ACCOUNT_ALREADY_EXISTS
//		
func doRegisterUser (w http.ResponseWriter, r *http.Request ) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	// validate everything
	if len(username) < 6 {
		http.Error(w, "username too short", http.StatusBadRequest)
		return
	}
	if len(password) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}
	if password != confirm {
		http.Error(w, "password doesn't match", http.StatusBadRequest)
		return
	}

	// go ahead and create the user
	pwBytes, bErr := bcrypt.GenerateFromPassword([]byte(password), 14)
	if bErr != nil {
		http.Error(w, bErr.Error(), http.StatusBadRequest)
		return
	}
	pwHash := string(pwBytes)
	// registering a free account
	u, cErr := CreateNewUser(r, username, pwHash, email, 0, "free")
	if cErr != nil {
		if (cErr.Code == ERR_ACCOUNT_ALREADY_EXISTS) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Attempt to create existing username: %s", username), nil)
			http.Error(w, cErr.Code, http.StatusBadRequest)
		} else {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Error creating user: %s : %s", username, cErr.Error()), nil)
			http.Error(w, cErr.Code, http.StatusInternalServerError)
		}
		return
	}

	// okay, everything worked, create a session
	sessionData := UserSessionData{Username:u.Username}
	_, err := webber.MakeSession(w, sessionData);
	if err != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("unable to create session: %s", err.Error()), nil)
		http.Error(w, "Please Log in", http.StatusUnauthorized)
	}
	webber.ReturnJson(w,u)
	return		

}

func doLoginUser (w http.ResponseWriter, r *http.Request ) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// fetch user with username, compare
	u, _ := GetUser(username)
	if ( u == nil ) {
		logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Invalid login for username: %s", username), nil)
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if ( err != nil) {
		logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Invalid login for username: %s", username), nil)
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	// create a session
	sessionData := UserSessionData{Username:u.Username}
	_, err2 := webber.MakeSession(w, sessionData);
	if err2 != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("unable to create session: %s", err2.Error()), nil)
		http.Error(w, "Please Log in", http.StatusUnauthorized)
		return
	}
	u.LastLogin = time.Now().Unix()
	UpdateUser(u)
	webber.ReturnJson(w,u)
	return		
	
}

func doSendResetEmail(destEmail string, resetCode string) {
	svc := ses.New(awsSession)
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(destEmail),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String("Here is the password reset code you requested: Go to  <a class=\"ulink\" href=\"http://" + ourConfig.PublicUrl + "/reset.html\" target=\"_blank\">http://" + ourConfig.PublicUrl + "/reset.html</a> and enter " + resetCode + " as your code."),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String("Here is the password reset code you requested: Go to http://" + ourConfig.PublicUrl + "/reset.html</a> and enter " + resetCode + " as your code."),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("Password Reset from Taskdeck"),
			},
		},
		//ReturnPath:    aws.String(""),
		//ReturnPathArn: aws.String(""),
		Source:        aws.String("jmhawkins@msn.com"),
		//SourceArn:     aws.String(""),
	}

	result, err := svc.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			case ses.ErrCodeConfigurationSetSendingPausedException:
				fmt.Println(ses.ErrCodeConfigurationSetSendingPausedException, aerr.Error())
			case ses.ErrCodeAccountSendingPausedException:
				fmt.Println(ses.ErrCodeAccountSendingPausedException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)

}

func doRequestPasswordResets(w http.ResponseWriter, r *http.Request)  {
	email := r.FormValue("email")
    var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

    b := make([]rune, 10)
    for i := range b {
        b[i] = letter[rand.Intn(len(letter))]
    }
    code := string(b)
	u := GetUserByEmail(email)
	if ( u == nil ) {
		// no user

	}
	u.ResetCode = code
	u.ResetTime = time.Now().Unix()
	UpdateUser(u)
	doSendResetEmail(u.Email, code);
	fmt.Fprintf(w, "Success")
	return		 	

}

func doPasswordReset(w http.ResponseWriter, r *http.Request)  {
	username := r.FormValue("username")
	resetCode := r.FormValue("resetcode")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	u, _ := GetUser(username)
	if ( u == nil ) {
		// no user, but just say incorrect code so we don't leak usernames
		http.Error(w, "incorrect reset code", http.StatusBadRequest) 
		return
	}
	// reset code is only good for two hours 
	if (time.Now().Unix() > u.ResetTime + 60*60*2) {
		// return error, code expired
		http.Error(w, "reset code expired", http.StatusBadRequest)
		return
	}
	if (u.ResetCode != resetCode ) {
		// return error, code doesn't match
		http.Error(w, "incorrect reset code", http.StatusBadRequest)
		return
	}
	if len(password) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}
	if password != confirm {
		http.Error(w, "password doesn't match", http.StatusBadRequest)
		return
	}
	// okay, reset the password and make a session
	pwBytes, bErr := bcrypt.GenerateFromPassword([]byte(password), 14)
	if bErr != nil {
		http.Error(w, bErr.Error(), http.StatusBadRequest)
		return
	}
	pwHash := string(pwBytes)	
	u.Password = pwHash
	u.ResetCode = ""
	u.ResetTime = 0
	sessionData := UserSessionData{Username:u.Username}
	_, err2 := webber.MakeSession(w, sessionData);
	if err2 != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("unable to create session: %s", err2.Error()), nil)
		http.Error(w, "Please Log in", http.StatusUnauthorized)
		return
	}
	u.LastLogin = time.Now().Unix()	
	UpdateUser(u)
	fmt.Fprintf(w, "Success")
	return		 	

}


func doUpdateProfile (w http.ResponseWriter, r *http.Request ) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	session := UserSessionData{}
	bHasSession, _ := webber.GetSession(r, &session)
	if ( bHasSession ) {
		u, _ := GetUser(session.Username)
		if ( u == nil ) {
			http.Error(w, "NoUser", http.StatusBadRequest);
			return
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if ( err != nil) {
				logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Invalid login for username: %s", u.Username), nil)
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}
			// if we're here, we're okay
			u.Email = email
			UpdateUser(u)
			webber.ReturnJson(w,u)
			return		

		}
	} else {
		http.Error(w, "NoSession", http.StatusUnauthorized)
		return
	}


}


func doChangePassword (w http.ResponseWriter, r *http.Request ) {
	oldpassword := r.FormValue("oldpassword")
	newpassword := r.FormValue("newpassword")
	confirm := r.FormValue("confirm")

	if len(newpassword) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}
	if newpassword != confirm {
		http.Error(w, "password doesn't match", http.StatusBadRequest)
		return
	}

	session := UserSessionData{}
	bHasSession, _ := webber.GetSession(r, &session)
	if ( bHasSession ) {
		u, _ := GetUser(session.Username)
		if ( u == nil ) {
			http.Error(w, "NoUser", http.StatusBadRequest);
			return
		} else {
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldpassword))
			if ( err != nil) {
				logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("Invalid login for username: %s", u.Username), nil)
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}
			// if we're here, we're okay
			pwBytes, bErr := bcrypt.GenerateFromPassword([]byte(newpassword), 14)
			if bErr != nil {
				http.Error(w, bErr.Error(), http.StatusBadRequest)
				return
			}
			pwHash := string(pwBytes)	
			u.Password = pwHash
			u.LastLogin = time.Now().Unix()	
			UpdateUser(u)
			webber.ReturnJson(w,u)
			return		

		}
	} else {
		http.Error(w, "NoSession", http.StatusUnauthorized)
		return
	}



}


func (h AuthServer) HandlePost (w http.ResponseWriter, r *http.Request) {
	parseErr := r.ParseForm()
	if parseErr != nil {
		logger.StdLogger.LOG(logger.ERROR, "", fmt.Sprintf("error parsing login form: %s", parseErr), nil)
	}

	
	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )

	if (len(pathParts) > 0) {
		switch pathParts[0] {
		case "register":
			doRegisterUser(w, r)
		case "login":
			doLoginUser(w, r)
		case "requestreset":
			doRequestPasswordResets(w, r)
		case "resetpassword":
			doPasswordReset(w,r)
		case "updateprofile":
			doUpdateProfile(w,r)
		case "changepassword":
			doChangePassword(w,r)
		default:
			http.Error(w, "NYI", http.StatusNotImplemented)
		}
	}
	



}

