// starterkit - code file for example api handling in the starterkit server application
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
	"fmt"
)

type TaskStatus int

const (
	STATUS_Open = TaskStatus(iota)   // 
	STATUS_Started
	STATUS_OnHold
	STATUS_InReview
	STATUS_Finished
)


type Task struct {
	Id int64			`bson:"_id" json:"id"` 
	UserId int64   		`json:"userid"`
	Title string		`json:"title"`
	Description string  `json:"description"`
	Impact int	 		`json:"impact"` 		// 0 = undefined, -1 = low impact, 1 = high impact
	DoOrDelegate int 	`json:"doordelegate"` 	// 0 = undefined, -1 = delegate, 1 = do self
	Priority int		`json:"priority"` 		// 0 = undefined, 1+ = priority in stack, 1 = highest
	Status TaskStatus   `json:"status"`
	Assigned int64		`json:"assigned"`		// 0 = unassigned, 1+ staffid a delegated task is assigned to

}


var taskArchiveColl *wtmcache.Collection


// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
func CreateTaskArchiveDbCollection ( cDb *wtmcache.Db)  {
	taskArchiveColl = cDb.NewCollection("taskarchive", "id", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

// CreateNewtask creates a new task with the title and description, using the next autoinc for taskid and returns it
//
func CreateNewTask(r *http.Request, user *User, title string, description string ) (*Task) {
	logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Creating New Task for User: %s", user.Username), nil)


	id := wtmcache.GetNextAutoIncValue("tasks")

	t := new(Task)
	t.Id = id
	t.UserId = user.Id
	t.Title = title
	t.Description = description
	t.Impact = 0
	t.DoOrDelegate = 0
	t.Priority = 0
	t.Status = STATUS_Open
	t.Assigned = 0
	
	return t
}

// AddTaskToEndOfArray taks ta task and adds it to the end of the aray referenced by slice, 
// growing it if needed.
// Returns the new array/slice
//
func AddTaskToEndOfArray ( slice []Task, add Task) []Task {
	n := len(slice)
	if cap(slice) == len(slice) {
		// full, we need to make it bigger
		// grow by 1 - no need to grow by more.  We're always serializing to the db, and the 
		// deserialization always sets cap == size, so growing by more than one doesn't do us
		// any good.
		newslice := make([]Task, len(slice), cap(slice)+1)
		copy(newslice, slice)
		slice = newslice[0:n+1]
		slice[n] = add
	} else {
		slice = slice[0 : n+1]
		slice[n] = add
	}

	return slice

}


// InsertTaskIntoArray inserts the task at the index-1 specified
// Returns the new array/slice
//
func InsertTaskIntoArray ( slice []Task, task *Task, index int64) ([]Task) {
	n := len(slice)
	if cap(slice) == len(slice) {
		// full, we need to make it bigger
		// grow by 1 - no need to grow by more.  We're always serializing to the db, and the 
		// deserialization always sets cap == size, so growing by more than one doesn't do us
		// any good.
		biggerSlice := make([]Task, len(slice), cap(slice)+1)
		copy(biggerSlice, slice)
		slice = biggerSlice[0:n+1]
	}
	if (index > int64(len(slice))) {
		index = int64(len(slice))
	}
	newSlice := append(slice[:index], append([]Task{*task}, slice[index:]...)...)
	for i := range newSlice {
		newSlice[i].Priority = i+1
	}
	return newSlice
}


// RemoveTaskIdFromArray removes the task with the id from the slice, returns the new slice and the task
//
func RemoveTaskIdFromArray ( slice []Task, taskId int64) ([]Task, *Task) {
	var t Task
	for i := range slice {
		t = slice[i]
		if (t.Id == taskId) {
			newslice := append(slice[:i], slice[i+1:]...)
			fmt.Println(newslice)
			return newslice, &t
		}
	}
	return slice, nil
}

// FindtaskInArray searches for the task with the id from the slice, returns the task
//
func FindTaskInArray ( slice []Task, taskId int64) (*Task) {
	var t Task
	for i := range slice {
		t = slice[i]
		if (t.Id == taskId) {
			return &t
		}
	}
	return nil
}

// UpdateTaskStatusInArray updates the status of a task in the array
func UpdateTaskStatusInArray ( slice []Task, taskId int64, newStatus TaskStatus) {
	if (newStatus > STATUS_Finished ) {
		newStatus = STATUS_Finished
	}
	for i := range slice {
		t := slice[i]
		if (t.Id == taskId) {
			slice[i].Status = newStatus
			return
		}
	}

}

func AddTaskToArchive ( task *Task) *AppError {

	err := taskArchiveColl.Write(*task)	
	if (err != nil) {
		return MakeError(ERR_INTERNAL_ERROR, err.Error())
	} else {
		return nil
	}
}


// This is our api/task handler
//

type TaskServer struct {
	basePath string
}

func NewTaskServer(basePath string) *TaskServer {
	f := new(TaskServer)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h TaskServer) Name() string {
	return "TaskServer"
}

func (h TaskServer) BasePath() string {
	return h.basePath
}

func (h TaskServer) Handler ( w http.ResponseWriter, r *http.Request) { 
	webber.DispatchMethod(h, w, r);
}

//////////////////////////////
//
// GET handlers
//

func (h TaskServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathVars := map[int]string{1:"taskid"}
	logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("TaskServer GET handler called for %s", apiPath), nil)
	pathParts, params := webber.ParsePathAndQueryFlat(r, apiPath, pathVars )

	switch pathParts[0] {
	case "new":
		doGetTasks(w,r,params, "new")
	default:
		http.Error(w, "NYI", http.StatusNotImplemented)
	
	}
	
}

func doGetTasks(w http.ResponseWriter, r *http.Request, params map[string]string, fromList string) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}

		// here is where we would get a task list from the user, but we will just return NYI
		http.Error(w, "NYI", http.StatusNotImplemented)
	}
}

/////////////////////////
//
// POST handlers
//

func (h TaskServer) HandlePost (w http.ResponseWriter, r *http.Request) {	
	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )

	if (len(pathParts) > 0) {
		switch pathParts[0] {
		case "newtask":
			doAddNewTaskForUser(w, r)
		default:
			http.Error(w, "NYI", http.StatusNotImplemented)
		}
	}

}


func doAddNewTaskForUser( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		// This is where you would add a task to the user, but we will just return NYI
		/*
		username := session.Username
		title := r.FormValue("title")
		description := r.FormValue("description")
		*/

		http.Error(w, "NYI", http.StatusNotImplemented)
	}
	// else we're already returned an error

	
}



