package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/* -------------------------------------------------------------------------- */
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
	statement, _ := db.Prepare(`CREATE TABLE IF NOT EXISTS Battles 
		(id INTEGER PRIMARY KEY, player1ID INTEGER, player2ID INTEGER, turn INTEGER)`)
	statement.Exec()
	statement, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Boards 
		(id INTEGER PRIMARY KEY, boardName TEXT, userID INTEGER, gameID INTEGER, 
		 created DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	statement.Exec()
	statement, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Positions 
		(id INTEGER PRIMARY KEY, boardID INTEGER, battleshipID INTEGER, 
		 playerID INTEGER, coordX INTEGER, coordY INTEGER, pinColor TEXT)`)
	statement.Exec()
	statement, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Ships 
		(id INTEGER PRIMARY KEY, shipType TEXT, shipLength INTEGER)`)
	statement.Exec()
	statement, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Users 
		(id INTEGER PRIMARY KEY, screenName TEXT, isActive BOOLEAN, 
		 lastLogin DATETIME)`)
	statement.Exec()

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

// check relationship between current user and a resource
func (app *application) checkRelationship(resourceID int) bool {
	return true
}

// add default data to board info screens
func (app *application) addDefaultDataBoard(td *templateDataBoard, r *http.Request) *templateDataBoard {
	if td == nil {
		td = &templateDataBoard{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}

// add default data to player info screens
func (app *application) addDefaultDataPlayer(td *templateDataPlayer, r *http.Request) *templateDataPlayer {
	if td == nil {
		td = &templateDataPlayer{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}

// cache templates for Boards
func (app *application) renderBoards(w http.ResponseWriter, r *http.Request, name string, td *templateDataBoard) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the current year
	err := ts.Execute(buf, app.addDefaultDataBoard(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// cache templates for Players
func (app *application) renderPlayers(w http.ResponseWriter, r *http.Request, name string, td *templateDataPlayer) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the current year
	err := ts.Execute(buf, app.addDefaultDataPlayer(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

/* -------------------------------------------------------------------------- */
// Error handling

// The serverError helper writes an error message and stack trace to the errorLog, // then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description // to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to // the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
