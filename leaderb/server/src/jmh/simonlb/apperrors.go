// Starterkit - code file for error handling in the starter kit server application
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

)

// These are the errors we can return for either 400 or 500 errors
const (
	ERR_ACCOUNT_ALREADY_EXISTS    	= "ACCOUNT_ALREADY_EXISTS"
	ERR_EMAIL_ALREADY_EXISTS    	= "EMAIL_ALREADY_EXISTS"
	ERR_COULDNT_ADD_TASK		  	= "COULD_NOT_ADD_TASK"
	ERR_USER_DOESNT_EXIST			= "USER_DOESNT_EXIST"
	ERR_READ_USER_FAILURE			= "READ_USER_FAILURE"
	ERR_WRITE_USER_FAILURE			= "WRITE_USER_FAILURE"
	ERR_INTERNAL_ERROR				= "INTERNAL_ERROR"
	ERR_TASK_NOT_FOUND				= "TASK_NOT_FOUND"
	ERR_WRONG_STATUS_FOR_ARCHIVE	= "WRONG_STATUS_FOR_ARCHIVE"
	ERR_TEAMNAME_ALREADY_TAKEN		= "TEAMNAME_ALREADY_TAKEN"
	ERR_TOURNEYNAME_ALREADY_EXISTS	= "TOURNEYNAME_ALREADY_EXISTS"
)

type AppError struct {
	Code string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func MakeError(code string, message string) *AppError {
	e := new(AppError)
	e.Code = code
	e.Message = message
	return e
}