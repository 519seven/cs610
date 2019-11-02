package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime/debug"
	"time"
	"strings"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/justinas/nosurf"							 // csrf prevention
)

/* ------------------------------------------------------------------------- */
// database

func initializeDB(dsn string, initdb bool) (*sql.DB, error) {
	// Do we need to remove the existing file before we begin?
	if initdb == true {
		// in this sense, initdb means start fresh
		if !os.IsNotExist(os.Remove(dsn)) {
			fmt.Println("==> database has been deleted; starting over")
		}
	}
	// in this sense, we are initializing the connection to the database
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	// for sqlite3 this doesn't matter, it only allows one connection at a time
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(2) // 2 is default

	if err = db.Ping(); err != nil {
		return nil, err
	}
	// create the tables if they don't exist
	// in sqlite3, a unique, auto-incrementing rowid is automatically created
	stmt, _ := db.Prepare(`CREATE TABLE IF NOT EXISTS Battles 
		(player1ID INTEGER, player2ID INTEGER, turn INTEGER)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Boards 
		(boardName TEXT, userID INTEGER, gameID INTEGER, 
		 created DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Players 
		(screenName TEXT, emailAddress TEXT NOT NULL UNIQUE, 
		 hashedPassword TEXT, created DATETIME, loggedIn BOOLEAN, 
		 inBattle BOOLEAN, lastLogin DATETIME)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Positions 
		(boardID INTEGER, shipID INTEGER, 
		 userID INTEGER, coordX INTEGER, coordY TEXT, pinColor TEXT)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Ships 
		(shipType TEXT, shipLength INTEGER)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`INSERT INTO Ships (shipType, shipLength) VALUES 
		('carrier', 5), ('battleship', 4), ('cruiser', 3), ('submarine', 3), ('destroyer', 2)`)
	stmt.Exec()

	return db, nil
}

/* -------------------------------------------------------------------------- */
// Method checking
/*
func checkMethod(m string, w http.ResponseWriter, r *http.Request) (http.ResponseWriter, bool) {
	if m == "POST" {
		// restrict this handler to HTTP POST methods only
		if r.Method != http.MethodPost {
			// Change response header map before WriteHeader or Write
			w.Header().Set("Allow", http.MethodPost)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			// WriteHeader must be called explicity before any call to Write
			//w.WriteHeader(405)
			//w.Write([]byte("Method Not Allowed\r\n"))
			// http.Error handles both WriteHeader and Write
			return w, false
		}
		return w, true
	} else if m == "GET" {
		// restrict this handler to HTTP POST methods only
		if r.Method != http.MethodGet {
			// Change response header map before WriteHeader or Write
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			// WriteHeader must be called explicity before any call to Write
			//w.WriteHeader(405)
			//w.Write([]byte("Method Not Allowed\r\n"))
			// http.Error handles both WriteHeader and Write
			return w, false
		}
		return w, true
	}
	return w, true
}
*/

// -----------------------------------------------------------------------------
// General

// Clean form data
func cleanCoordinates(coordinates string) string {
	cleanString := strings.Replace(coordinates, "\t", "", -1)
	cleanString = strings.Replace(cleanString, "\n", "", -1)
	return strings.Replace(cleanString, " ", "", -1)
}

// Check relationship between current user and a resource
func (app *application) checkRelationship(resourceID int) bool {
	return true
}

// For Authorization
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// Pre-processing HTML/template data
// - Draw the board and pass in the HTML
func (app *application) preprocessBoard(r *http.Request) template.HTML {
	gameBoard := "<table><th>&nbsp;</th>"
	for _, col := range "ABCDEFGHIJ" {
		gameBoard += fmt.Sprintf("<th>%s</th>", string(col))
	}
	for row := 1; row < 11; row++ {
		gameBoard += fmt.Sprintf("<tr><td>%d</td>", row)
		rowStr := strconv.Itoa(row)
		for _, col := range "ABCDEFGHIJ" {
			gameBoard += fmt.Sprintf(
				"<td><input type='text'	maxlength=1 size=6 name=\"shipXY%d%s\" value=\"%s\"></td>", 
				row, string(col), r.PostForm.Get("shipXY"+rowStr+string(col)))
		}
		gameBoard += "</tr>"
	}
	gameBoard += "</table>"
	return template.HTML(gameBoard)
}

// ----------------------------------------------------------------------------
// Create template data helpers so we can add items.  This information is 
// automatically available each time we render a template
// ----------------------------------------------------------------------------

// Add default data to create board interface
func (app *application) addDefaultDataBoard(td *templateDataBoard, r *http.Request) *templateDataBoard {
	if td == nil {
		td = &templateDataBoard{}
	}
	td.ActiveBoardID = app.session.GetInt(r, "boardID")
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	// Positions
	//td.PositionsOnBoard = app.positionsOnBoard(r)
	//td.PositionsOnBoards = app.positionsOnBoard(r)
	// Add a new string containing processed template snippet
	td.MainGrid = app.preprocessBoard(r)
	return td
}

// Add default data to list of boards screens
func (app *application) addDefaultDataBoards(td *templateDataBoards, r *http.Request) *templateDataBoards {
	if td == nil {
		td = &templateDataBoards{}
	}
	boardID, err := strconv.Atoi(app.session.GetString(r, "boardID"))
	if err != nil {
		fmt.Println("[INFO] boardID is", err)
	}
	td.ActiveBoardID = boardID
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Add default data to login screens
func (app *application) addDefaultDataLogin(td *templateDataLogin, r *http.Request) *templateDataLogin {
	if td == nil {
		td = &templateDataLogin{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Add default data to player info screens
func (app *application) addDefaultDataPlayer(td *templateDataPlayer, r *http.Request) *templateDataPlayer {
	if td == nil {
		td = &templateDataPlayer{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Add default data to player info screens
func (app *application) addDefaultDataPlayers(td *templateDataPlayers, r *http.Request) *templateDataPlayers {
	if td == nil {
		td = &templateDataPlayers{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Add default data to signup screens
func (app *application) addDefaultDataSignup(td *templateDataSignup, r *http.Request) *templateDataSignup {
	if td == nil {
		td = &templateDataSignup{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Cache Templates
// ----------------------------------------------------------------------------

// Board
func (app *application) renderBoard(w http.ResponseWriter, r *http.Request, name string, td *templateDataBoard) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// Write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// Execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataBoard(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// Boards
func (app *application) renderBoards(w http.ResponseWriter, r *http.Request, name string, td *templateDataBoards) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataBoards(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// Login
func (app *application) renderLogin(w http.ResponseWriter, r *http.Request, name string, td *templateDataLogin) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataLogin(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// Player
func (app *application) renderPlayer(w http.ResponseWriter, r *http.Request, name string, td *templateDataPlayer) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataPlayer(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// Players
func (app *application) renderPlayers(w http.ResponseWriter, r *http.Request, name string, td *templateDataPlayers) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataPlayers(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// Signup
func (app *application) renderSignup(w http.ResponseWriter, r *http.Request, name string, td *templateDataSignup) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataSignup(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

/* -------------------------------------------------------------------------- */
// Error handling

// The serverError helper writes an error message and stack trace to the errorLog
//  - Sends a generic 500 Internal Server Error response to the user
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and description to the user
// - Like 400 "Bad Request" when there's a problem with the request that the user sent
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// And, a notFound helper for consistency
// - A convenience wrapper around clientError
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
