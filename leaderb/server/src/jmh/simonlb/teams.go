// teams - code file for team api handling in the simon leaderboard (simonlb) server application
//
// Copyright (c) 2019 - John M. Hawkins <jmhawkins@msn.com>
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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"strconv"
	"fmt"
)

/*
type TaskStatus int

const (
	STATUS_Open = TaskStatus(iota)   // 
	STATUS_Started
	STATUS_OnHold
	STATUS_InReview
	STATUS_Finished
)
*/

type Team struct {
	Id int64			`bson:"_id" json:"id"` 
	TeamId string		`json:"teamid"`
	TourneyId int64		`json:"tourneyid"`
	Name string			`json:"name"`
	Sponsor string  	`json:"sponsor"`
	PoolScore int64		`json:"poolscore"` 		// highest score (number of successful sequences) achieved
	FinalsScore int64	`json:"finalsscore"`
	PoolSlot int64 		`json:"poolslot"` 	// where they are in the pool schedule
	FinalsSlot int64	`json:"finalsslot"`
}

var teamColl *wtmcache.Collection


// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
func CreateTeamDbCollection ( cDb *wtmcache.Db)  {
	teamColl = cDb.NewCollection("teams", "teamid", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

// CreateNewTeam creates a new task with the title and description, using the next autoinc for taskid and returns it
//
func CreateNewTeam(r *http.Request, user *User, tourneyId int64, name string, sponsor string ) (*Team, *AppError) {
	logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Creating New Team, by User: %s", user.Username), nil)

	// check that an account with that username doesn't already exist
	var docTemplate Team
	_, dbErr := teamColl.Read(name, &docTemplate) 
	if ( dbErr == nil ) {
		// already exists
		return nil, MakeError(ERR_TEAMNAME_ALREADY_TAKEN, "Team already exists")
	}


	id := wtmcache.GetNextAutoIncValue("teams")

	t := new(Team)
	t.Id = id
	t.TeamId = strconv.FormatInt(id, 10)
	t.TourneyId = tourneyId
	t.Name = name
	t.Sponsor = sponsor
	t.PoolScore = 0
	t.FinalsScore = 0
	t.PoolSlot = 0
	t.FinalsSlot = 0
	
	teamColl.Write(*t)

	return t, nil
}

// AddTeamToEndOfArray takes a team and adds it to the end of the aray referenced by slice, 
// growing it if needed.
// Returns the new array/slice
//
func AddTeamToEndOfArray ( slice []Team, add Team) []Team {
	n := len(slice)
	if cap(slice) == len(slice) {
		// full, we need to make it bigger
		// grow by 1 - no need to grow by more.  We're always serializing to the db, and the 
		// deserialization always sets cap == size, so growing by more than one doesn't do us
		// any good.
		newslice := make([]Team, len(slice), cap(slice)+1)
		copy(newslice, slice)
		slice = newslice[0:n+1]
		slice[n] = add
	} else {
		slice = slice[0 : n+1]
		slice[n] = add
	}

	return slice

}


// InsertTeamIntoArray inserts the team at the index-1 specified
// Returns the new array/slice
//
func InsertTeamIntoArray ( slice []Team, team *Team, index int64) ([]Team) {
	n := len(slice)
	if cap(slice) == len(slice) {
		// full, we need to make it bigger
		// grow by 1 - no need to grow by more.  We're always serializing to the db, and the 
		// deserialization always sets cap == size, so growing by more than one doesn't do us
		// any good.
		biggerSlice := make([]Team, len(slice), cap(slice)+1)
		copy(biggerSlice, slice)
		slice = biggerSlice[0:n+1]
	}
	if (index > int64(len(slice))) {
		index = int64(len(slice))
	}
	newSlice := append(slice[:index], append([]Team{*team}, slice[index:]...)...)
	return newSlice
}


// RemoveTeamIdFromArray removes the Team with the id from the slice, returns the new slice and the Team
//
func RemoveTeamIdFromArray ( slice []Team, teamId int64) ([]Team, *Team) {
	var t Team
	for i := range slice {
		t = slice[i]
		if (t.Id == teamId) {
			newslice := append(slice[:i], slice[i+1:]...)
			fmt.Println(newslice)
			return newslice, &t
		}
	}
	return slice, nil
}

// FindTeamkInArray searches for the Team with the id from the slice, returns the Team
//
func FindTeamInArray ( slice []Team, taskId int64) (*Team) {
	var t Team
	for i := range slice {
		t = slice[i]
		if (t.Id == taskId) {
			return &t
		}
	}
	return nil
}

// UpdateTeamHighScoreInArray updates the high score of a team in the array
func UpdateTeamHighScoreInArray ( slice []Team, teamId int64, newHighScore int64) {
	for i := range slice {
		t := slice[i]
		if (t.Id == teamId) {
			slice[i].PoolScore = newHighScore
			return
		}
	}

}

// Game state (tourney = tournament) handler
//

type Tourney struct {
	Id int64			`bson:"_id" json:"id"` 
	TourneyId string	`json:"tourneyid"`
	Name string			`json:"name"`	// name of the tournament
	NumTeams int64		`json:"numteams"`	// how many teams in the tourney
	NumFinals int64		`json:"numfinals"`	// how many teams make the finals
	PoolPlay int64		`json:"pool"`  	// 0 = pool play not going, >0 = cur team in pool play
	Finals int64		`json:"finals"`	// 0 = finals not going, >0 = cur team in finals play
}

var tourneyColl *wtmcache.Collection


// NOTE: this only works if we're a single instance.  If we are multiple instances,
// we need to use the wtmdb (still TBD) to ensure we don't have divergent caches.
func CreateTourneyDbCollection ( cDb *wtmcache.Db)  {
	tourneyColl = cDb.NewCollection("tourney", "tourneyid", 14*60*24*time.Minute, 14*60*24*time.Minute)
}

func StartNewTourney(name string, numFinals int64) (*Tourney, *AppError) {
	var docTemplate Tourney
	dbErr := tourneyColl.Dbc.Find(bson.M{"name": name}).One(&docTemplate) 
	if (dbErr == nil ) {
		// tourney with that name already exists
		return nil, MakeError(ERR_TOURNEYNAME_ALREADY_EXISTS, "A tournament with that name already exists")
	}

	id := wtmcache.GetNextAutoIncValue("tourneys")

	t := new(Tourney)
	t.Id = id
	t.TourneyId = strconv.FormatInt(id, 10)
	t.Name = name
	t.NumTeams = 0
	t.NumFinals = numFinals
	t.PoolPlay = 0
	t.Finals = 0

	tourneyColl.Write(*t)

	return t, nil

}


func GetCurTourneyStatus() (*Tourney, *AppError){
	var tourneys []Tourney
	dbErr := tourneyColl.QueryAndSort(bson.M{},&tourneys, "-_id")
	if ( dbErr == nil && len(tourneys) > 0) {
		return &(tourneys[0]), nil
	} else {
		return nil, MakeError(ERR_INTERNAL_ERROR, "No tournaments found")
	}



}

func IncNumTeamsInCurTourney(tourneyId int64) (int64){
	changeToApply := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"numteams": 1}},
		ReturnNew: true,
	}
	var result bson.M
	_, err :=tourneyColl.Dbc.Find(bson.M{"_id": tourneyId}).Apply(changeToApply, &result)
	if ( err != nil ) {
		return 0;
	}
	return result["numteams"].(int64)

}


// This is our api/task handler
//

type TeamServer struct {
	basePath string
}

func NewTeamServer(basePath string) *TeamServer {
	f := new(TeamServer)
	f.basePath = "/" + basePath + "/"
	return f 
}

func (h TeamServer) Name() string {
	return "TeamServer"
}

func (h TeamServer) BasePath() string {
	return h.basePath
}

func (h TeamServer) Handler ( w http.ResponseWriter, r *http.Request) { 
	webber.DispatchMethod(h, w, r);
}

//////////////////////////////
//
// GET handlers
//

func (h TeamServer) HandleGet (w http.ResponseWriter, r *http.Request) {
	apiPath := r.URL.Path[len(h.basePath):]
	pathVars := map[int]string{1:"taskid"}
	logger.StdLogger.LOG(logger.INFO, "", fmt.Sprintf("TeamServer GET handler called for %s", apiPath), nil)
	pathParts, params := webber.ParsePathAndQueryFlat(r, apiPath, pathVars )

	switch pathParts[0] {
	case "fetch":
		doGetTeams(w,r,params )
	case "startquals":
		doStartQuals(w,r,params )
	case "startfinals":
		doStartFinals(w,r,params )
	case "endgame":
		doEndGame(w,r)
	case "leaderboard":
		doGetLeaderboard(w,r,params )
	case "champion":
		doGetChampion(w,r,params )
	default:
		http.Error(w, "NYI", http.StatusNotImplemented)
	
	}
	
}

func GetTourneyId(params map[string]string) (tourneyId int64) {
	if tidStr, ok := params["tourney"]; ok {
		// return the int64 of the string
		tid, err := strconv.ParseInt(tidStr, 10, 64)
		if (err == nil) {
			return tid
		}
	} else {
		// get the current one
		t, err := GetCurTourneyStatus()
		if (err == nil ) {
			return t.Id
		}
	}
	return 0
}

func GetCurTourneyId() (tourneyId int64) {
	// get the current one
	t, err := GetCurTourneyStatus()
	if (err == nil ) {
		return t.Id
	}
	return 0
}

func doGetTeams(w http.ResponseWriter, r *http.Request, params map[string]string) {
	session := GetUserSessionHandleError(w,r)
	// we actually don't care about the user yet, but we might later if we want to decorate the UI with
	// capabilities.  
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}
	}

	tourneyId := GetTourneyId(params)
	if (tourneyId == 0) {
		http.Error(w, "No tourney", http.StatusInternalServerError);
	}
	var teams []Team
	err2 := teamColl.Query(bson.M{"tourneyid":tourneyId},&teams) 
	if (err2 == nil) {
		webber.ReturnJson(w,teams)
	} else {
		http.Error(w, "foo", http.StatusInternalServerError)
		return
	}


}


func doStartQuals(w http.ResponseWriter, r *http.Request, params map[string]string) {
	session := GetUserSessionHandleError(w,r)
	// we actually don't care about the user yet, but we might later if we want to decorate the UI with
	// capabilities.  
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}
	}

	tourney, tErr := GetCurTourneyStatus()
	if (tErr != nil ) {
		logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get current tourney, err = %s", tErr.Error()), nil)
		http.Error(w, tErr.Code, http.StatusInternalServerError)
		return
	}

	if (tourney.PoolPlay > 0 ) {
		// we already started
		logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Pool play already started" ), nil)
		http.Error(w, ERR_POOL_PLAY_ALREADY_STARTED, http.StatusInternalServerError)
		return
	}

	tourney.PoolPlay = 1
	tourneyColl.Write(*tourney)

	webber.ReturnJson(w,tourney)

}


func doStartFinals(w http.ResponseWriter, r *http.Request, params map[string]string) {
	session := GetUserSessionHandleError(w,r)
	// we actually don't care about the user yet, but we might later if we want to decorate the UI with
	// capabilities.  
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}
	}

	tourney, tErr := GetCurTourneyStatus()
	if (tErr != nil ) {
		logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get current tourney, err = %s", tErr.Error()), nil)
		http.Error(w, tErr.Code, http.StatusInternalServerError)
		return
	}

	if (tourney.Finals > 0 ) {
		// we already started
		logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Finals already started" ), nil)
		http.Error(w, ERR_POOL_PLAY_ALREADY_STARTED, http.StatusInternalServerError)
		return
	}

	// go through teams and tag all the ones in the finals with their new schedule
	// TBD

	tourney.PoolPlay = 0
	tourney.Finals = 1
	tourneyColl.Write(tourney)

	webber.ReturnJson(w,tourney)

}



func doGetLeaderboard(w http.ResponseWriter, r *http.Request, params map[string]string) {
	session := GetUserSessionHandleError(w,r)
	// we actually don't care about the user yet, but we might later if we want to decorate the UI with
	// capabilities.  
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}
	}


	tourney, tErr := GetCurTourneyStatus()
	if (tErr != nil ) {
		logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get current tourney, err = %s", tErr.Error()), nil)
		http.Error(w, tErr.Code, http.StatusInternalServerError)
		return
	}
	
	var sortby string
	if (tourney.Finals > 0 ) {
		sortby = "-finalscore"
	} else {
		sortby = "-poolscore"
	}
	
	var teams []Team
//	err2 := teamColl.Query(bson.M{"tourneyid":tourney.Id},&teams) 
	err2 := teamColl.QueryAndSort(bson.M{"tourneyid":tourney.Id},&teams, sortby) 
	if (err2 == nil) {
		webber.ReturnJson(w,teams)
	} else {
		http.Error(w, "foo", http.StatusInternalServerError)
		return
	}


}



func doGetChampion(w http.ResponseWriter, r *http.Request, params map[string]string) {
	session := GetUserSessionHandleError(w,r)
	// we actually don't care about the user yet, but we might later if we want to decorate the UI with
	// capabilities.  
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}
	}


	// TBD return a list of teams...
	http.Error(w, "NYI", http.StatusNotImplemented)

}


/////////////////////////
//
// POST handlers
//

func (h TeamServer) HandlePost (w http.ResponseWriter, r *http.Request) {	
	apiPath := r.URL.Path[len(h.basePath):]
	pathParts, _ := webber.ParsePathAndQueryFlat(r, apiPath, nil )

	if (len(pathParts) > 0) {
		switch pathParts[0] {
		case "newtourney":
			doStartNewTourney(w,r, )
		case "newteam":
			doAddNewTeam(w, r)
		case "startgame":
			doStartGame(w,r)
		case "registerscore":
			doRegisterScore(w,r)
		default:
			http.Error(w, "NYI xx", http.StatusNotImplemented)
		}
	}

}


func doStartNewTourney( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}

		// add a tourney, validating the user has permission to add it and that the name doesn't already exist
		
		name := r.FormValue("name")
		numfinalsStr := r.FormValue("numfinals")
		numFinals, _ := strconv.ParseInt(numfinalsStr, 10, 64)
		
		t, tErr := StartNewTourney(name, numFinals);
		if tErr != nil {
			if (tErr.Code == ERR_TEAMNAME_ALREADY_TAKEN) {
				logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Attempt to create existing tourney: %s", name), nil)
				http.Error(w, tErr.Code, http.StatusBadRequest)
			} else {
				logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Error creating tourney: %s : %s", name, tErr.Error()), nil)
				http.Error(w, tErr.Code, http.StatusInternalServerError)
			}
			return
		}

		// okay, it worked
		webber.ReturnJson(w,t)
		return		
	}
	// else we're already returned an error

	
}

func doAddNewTeam( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}

		// add a team, validating the user has permission to add it and that the team name doesn't already exist
		
		name := r.FormValue("name")
		sponsor := r.FormValue("sponsor")
		tourney, tErr := GetCurTourneyStatus()
		if (tErr != nil) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Cant find current tourney info"), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}
		tourneyId := tourney.Id

		// if the finals round has already started, teams can't be added!
		if ( tourney.Finals > 0 ) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Attempted to add a team after finals round started"), nil)
			http.Error(w, "FINALS_IN_PROGRESS", http.StatusBadRequest)
			return
		}

		t, tErr := CreateNewTeam(r, u, tourneyId, name, sponsor);
		if tErr != nil {
			if (tErr.Code == ERR_TEAMNAME_ALREADY_TAKEN) {
				logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Attempt to create existing team: %s", name), nil)
				http.Error(w, tErr.Code, http.StatusBadRequest)
			} else {
				logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Error creating team: %s : %s", name, tErr.Error()), nil)
				http.Error(w, tErr.Code, http.StatusInternalServerError)
			}
			return
		}
		numTeams := IncNumTeamsInCurTourney(tourneyId)

		// if pool play has already started, give team next schedule id
		t.PoolSlot = numTeams
		errUpdate := teamColl.Write(*t)			
		if (errUpdate != nil) {
				// retry?
				teamColl.Write(*t)			
		}

		// okay, it worked
		webber.ReturnJson(w,t)
		return		
	}
	// else we're already returned an error

	
}


func doStartGame( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}



		http.Error(w, "NYI", http.StatusNotImplemented)
		return		
	}
	// else we're already returned an error

	
}


func doRegisterScore( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}

		tourney, tErr := GetCurTourneyStatus()
		if (tErr != nil) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Cant find current tourney info"), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}
		//tourneyId := tourney.Id
		var curTeamId string
		var bFinals bool
		if (tourney.Finals > 0 ) {
			curTeamId = strconv.FormatInt(tourney.Finals, 10)
			bFinals = true
		} else if ( tourney.PoolPlay > 0 ) {
			curTeamId = strconv.FormatInt(tourney.PoolPlay, 10)
			bFinals = false
		} else {
			// we haven't started yet...
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Tried to record score before starting tourney"), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}
		
		var docTemplate Team
		t, dbErr := teamColl.Read(curTeamId, &docTemplate) 
		if ( dbErr != nil ) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Error updating score, team=%s, err=%s", curTeamId, dbErr.Error()), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}
		team, _ := t.(*Team)

		scoreStr := r.FormValue("score")
		scoreInt, cErr := strconv.ParseInt(scoreStr, 10, 32)
		if ( cErr != nil ) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Error updating score, team=%s, err=%s", curTeamId, cErr.Error()), nil)
			http.Error(w, tErr.Code, http.StatusInternalServerError)
			return
		}
		if ( bFinals) {
			if ( scoreInt > team.FinalsScore) {
				team.FinalsScore = scoreInt
				teamColl.Write(*team)
			}
		} else {
			if ( scoreInt > team.PoolScore) {
				team.PoolScore = scoreInt
				teamColl.Write(*team)
			}
		}
		

		webber.ReturnJson(w,team)

	}
	// else we're already returned an error

	
}


func doEndGame( w http.ResponseWriter, r *http.Request) {
	session := GetUserSessionHandleError(w,r)
	if  (session != nil ) {
		u, err := GetUser(session.Username)
		if (err != nil ) {
			logger.StdLogger.LOG(logger.ERROR, webber.GetCorrelationId(r), fmt.Sprintf("Can't get user for userrname %s, err = %s", u.Username, err.Error()), nil)
			http.Error(w, err.Code, http.StatusInternalServerError)
			return
		}

		tourney, tErr := GetCurTourneyStatus()
		if (tErr != nil) {
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Cant find current tourney info"), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}
		//tourneyId := tourney.Id
		if (tourney.Finals > 0 ) {
			tourney.Finals = tourney.Finals + 1
		} else if ( tourney.PoolPlay > 0 ) {
			tourney.PoolPlay = tourney.PoolPlay + 1
		} else {
			// we haven't started yet...
			logger.StdLogger.LOG(logger.INFO, webber.GetCorrelationId(r), fmt.Sprintf("Tried to record score before starting tourney"), nil)
			http.Error(w, tErr.Code, http.StatusBadRequest)
			return
		}

		
		tourneyColl.Write(*tourney)

		webber.ReturnJson(w,tourney)
			


	}
	// else we're already returned an error

	
}

