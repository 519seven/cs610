package main

// ----------------------------------------------------------------------------
// Copyright 2019 Peter J. Akey
// helpers.go
//
// Two pieces to every part
// - Template data helpers helps us add items that can be accessed within HTML
// - Page/template rendering
// ----------------------------------------------------------------------------

import (
	"bytes"
	"crypto/rand"
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
	"github.com/519seven/cs610/battleship/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// ----------------------------------------------------------------------------
// DATABASE

// Initialize the database - create tables, pupulate ship types
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
		(player1ID INTEGER, player1Accepted BOOLEAN, player1BoardID INTEGER,
		 player2ID INTEGER, player2Accepted BOOLEAN, player2BoardID INTEGER, 
		 player1SunkenShips INTEGER DEFAULT 0, player2SunkenShips INTEGER DEFAULT 0,
		 challengeDate DATETIME DEFAULT CURRENT_TIMESTAMP,
		 turn INTEGER, secretTurn STRING, winner INTEGER DEFAULT 0)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Boards 
		(boardName TEXT, playerID INTEGER, 
		 created DATETIME DEFAULT CURRENT_TIMESTAMP, isChosen BOOLEAN)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Players 
		(screenName TEXT NOT NULL UNIQUE, emailAddress TEXT, 
		 hashedPassword TEXT, created DATETIME, loggedIn BOOLEAN, 
		 inBattle BOOLEAN, lastLogin DATETIME)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Positions 
		(boardID INTEGER, shipID INTEGER, playerID INTEGER, 
		 coordX INTEGER, coordY TEXT, pinColor TEXT)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS Ships 
		(shipType TEXT, shipLength INTEGER)`)
	stmt.Exec()
	stmt, _ = db.Prepare(`INSERT INTO Ships (shipType, shipLength) VALUES 
		('carrier', 5), ('battleship', 4), ('cruiser', 3), ('submarine', 3), ('destroyer', 2)`)
	stmt.Exec()
	// sample accounts
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("B0mbs4way:("), 13)
	elvisPassword, err := bcrypt.GenerateFromPassword([]byte("P34nutButter76"), 13)
	stmt, _ = db.Prepare(`INSERT INTO Players (screenName, emailAddress, hashedPassword) VALUES
		('bob', 'bob@bob.com', ?), ('sue', 'sue@sue.com', ?), 
		('elvis', 'elvis@graceland.com', ?), ('maria', 'maria@presley.com', ?)`)
	stmt.Exec(hashedPassword, hashedPassword, elvisPassword, hashedPassword)
	// I created a board for Bob in battleship.db.sample
	// Running `make` will copy that db into place
	fmt.Println("Using sample database; If behavior is unpredictable, ")
	fmt.Println("rename battleship.db.sample to battleship.db.dontuse ")
	fmt.Println("and run `make clean && make` again...")

	return db, nil
}

// -----------------------------------------------------------------------------
// GENERAL HELPERS

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
	//fmt.Println("Inspecting context for authentication...")				// debug
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		//fmt.Println("isAuthenticated returned false!")					// debug
		return false
	}
	return isAuthenticated
}


// Pre-processing HTML/template data based on data from database
func (app *application) preprocessBoardFromData(p []*models.Position, battleID int, csrf_token string, permissions string) template.HTML {
	var boardID int = 0
	var pinColor string = ""
	var tableID string = ""
	var toolTip string = ""
	if (permissions == "hidden") {
		tableID = "opponent"
		toolTip = "These boxes represent your opponents board. Select one that is not gray (an unsuccessful strike) or red (a successful strike) to launch a missile towards your opponent."
	} else {
		tableID = "challenger"
		toolTip = "These boxes show you which coordinates have been launched against you already. There is nothing for you to do on this board."
	}

	gameBoard := fmt.Sprintf("<table id=\"%s\"><th>&nbsp;</th>", tableID)
	for _, col := range "ABCDEFGHIJ" {
		gameBoard += fmt.Sprintf("<th>%s</th>", string(col))
	}
	for row := 1; row < 11; row++ {
		gameBoard += "<tr>"
		gameBoard += fmt.Sprintf("<td>%d</td>", row)
		//rowStr := strconv.Itoa(row)
		for _, col := range "ABCDEFGHIJ" {
			var fieldHTML string; fieldHTML = ""
			var fieldValue string; fieldValue = "&nbsp;"
			var inputid string; inputid = ""
			var checked string; checked = ""
			var onclick string; onclick = ""
			var class string; class = ""
			for _, onePosition := range p {
				// This playerID ought to help us determine whose board these checkboxes belong to
				if (boardID == 0) { boardID = onePosition.ID; }
				//if (pinColor == "" || pinColor == "0") { pinColor = onePosition.PinColor; }
				if onePosition.CoordX == row && onePosition.CoordY == string(col) {
					if onePosition.ShipType.Valid {
						if strings.ToUpper(onePosition.ShipType.String) == "CRUISER" {
							fieldValue = "R"	// Cruiser is represented with an "R"
						} else {
							fieldValue = strings.ToUpper(onePosition.ShipType.String[0:1])
						}
					}
					//fmt.Println("pinColor:", onePosition.PinColor)
					if onePosition.PinColor != "" && onePosition.PinColor != "0" {
						checked = "checked"
						onclick = "onclick=\"return false;\""
						class = fmt.Sprintf("class='%sBattleBoard' ", onePosition.PinColor)
						pinColor = onePosition.PinColor;
					} else {
						checked = "";
						onclick = "";
						class = "";
						inputid = "";
						pinColor = "";
					}
					break
				}
			}
			fieldName := fmt.Sprintf("%d_shipXY%d%s", boardID, row, string(col))
			//gameBoard += fmt.Sprintf("<td id=\"%d_%s\">", playerID, fieldName)
			gameBoard += fmt.Sprintf("<td id=\"%s_%s\" style=\"background-color:%s\">", tableID, fieldName, pinColor)
			if permissions == "ro" {
				fieldHTML = fmt.Sprintf("<label id=\"%s\" class=\"container\">%s<input type='checkbox' name=\"%s\" %s value=\"%s\" %s %s><span class=\"checkmark\" title=\"%s\"></span></label>", 
					fieldName, fieldValue, fieldName, class, fieldValue, inputid, checked, toolTip)
			} else if permissions == "rw" {
				fieldHTML = fmt.Sprintf(
					"<label id=\"%s\" class=\"container\"><input type='text' maxlength=1 size=6 name=\"%s\" %s value=\"%s\" onclick=\"save_checkbox('%s');\" %s><span class=\"checkmark\"></span></label>", 
					fieldName, fieldName, class, fieldValue, fieldName, inputid)
			} else if permissions == "hidden" {
				fieldHTML = fmt.Sprintf("<label id=\"%s\" class=\"container\"><input class=\"striker\" type='checkbox' name=\"%s\" %s %s %s %s><span title=\"%s\" class=\"checkmark\"></span></label>", 
					fieldName, fieldName, class, onclick, inputid, checked, toolTip)
			}
			//fmt.Println(fieldName, ":", fieldHTML)						// debug
			gameBoard += fieldHTML + "</td>"
			pinColor = "";
		}
		gameBoard += "</tr>"
	}
	gameBoard += "</table>"
	return template.HTML(gameBoard)
}


// Pre-processing HTML/template data based on data found in the form request
func (app *application) preprocessBoardFromRequest(r *http.Request) template.HTML {
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


// Add default data to login screens
func (app *application) addDefaultDataLogin(td *templateDataLogin, r *http.Request) *templateDataLogin {
	if td == nil {
		td = &templateDataLogin{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}


// ----------------------------------------------------------------------------
// ABOUT

// Add default data to create about template
func (app *application) addDefaultDataAbout(td *templateDataAbout, r *http.Request) *templateDataAbout {
	if td == nil {
		td = &templateDataAbout{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}


// Render signup form
func (app *application) renderAbout(w http.ResponseWriter, r *http.Request, name string, td *templateDataAbout) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataAbout(td, r))
	
	// Remove session information
	if app.session != nil {
		app.session.Remove(r, "authenticatedPlayerID")
		app.session.Remove(r, "boardID")
		app.session.Remove(r, "battleID")
		//fmt.Println("Session information has been removed...")			// debug
	}

	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)}

// ----------------------------------------------------------------------------
// BATTLES

// Add default data to create battle interface
func (app *application) addDefaultDataBattle(td *templateDataBattle, r *http.Request) *templateDataBattle {
	if td == nil {
		td = &templateDataBattle{}
	}
	// Default the boardID to 0
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = app.session.GetInt(r, "boardID")
	} else {
		td.ActiveBoardID = 0
	}
	td.AuthenticatedPlayerID = app.session.GetInt(r, "authenticatedPlayerID")
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	if td.ChallengerPositions != nil {
		//fmt.Println("Positions is not nil.  We should build MainGrid here...", td.AuthenticatedPlayerID, td.ChallengerID)
		if td.AuthenticatedPlayerID == td.ChallengerID {
			td.ChallengerGrid = app.preprocessBoardFromData(td.ChallengerPositions, td.Battle.ID, td.CSRFToken, "ro")
		} else {
			td.ChallengerGrid = app.preprocessBoardFromData(td.OpponentPositions, td.Battle.ID, td.CSRFToken, "ro")
		}
	} else {
		td.ChallengerGrid = app.preprocessBoardFromRequest(r)
	}
	if td.OpponentPositions != nil {
		//fmt.Println("Positions is not nil.  We should build MainGrid here...", td.AuthenticatedPlayerID, td.ChallengerID)
		if td.AuthenticatedPlayerID == td.ChallengerID {
			td.OpponentGrid = app.preprocessBoardFromData(td.OpponentPositions, td.Battle.ID, td.CSRFToken, "hidden")
		} else {
			td.OpponentGrid = app.preprocessBoardFromData(td.ChallengerPositions, td.Battle.ID, td.CSRFToken, "hidden")
		}
	} else {
		td.OpponentGrid = app.preprocessBoardFromRequest(r)
	}
	return td
}


// renderBattle
func (app *application) renderBattle(w http.ResponseWriter, r *http.Request, name string, td *templateDataBattle) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// Write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// Execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataBattle(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}


// Add default data to create list of battles
func (app *application) addDefaultDataBattles(td *templateDataBattles, r *http.Request) *templateDataBattles {
	if td == nil {
		td = &templateDataBattles{}
	}
	// Default the boardID to 0
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = app.session.GetInt(r, "boardID")
	} else {
		td.ActiveBoardID = 0
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}


// renderBattles
func (app *application) renderBattles(w http.ResponseWriter, r *http.Request, name string, td *templateDataBattles) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataBattles(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}


// ----------------------------------------------------------------------------
// BOARDS

// Add default data to create board interface
func (app *application) addDefaultDataBoard(td *templateDataBoard, r *http.Request) *templateDataBoard {
	if td == nil {
		td = &templateDataBoard{}
	}
	if td.Positions != nil {
		//fmt.Println("Positions is not nil.  We should build MainGrid here...")
		td.MainGrid = app.preprocessBoardFromData(td.Positions, 0, "", "ro")
	} else {
		td.MainGrid = app.preprocessBoardFromRequest(r)
	}
	// Default the boardID to 0
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = app.session.GetInt(r, "boardID")
	} else {
		td.ActiveBoardID = 0
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}


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


// Add default data to list of boards screens
func (app *application) addDefaultDataBoards(td *templateDataBoards, r *http.Request) *templateDataBoards {
	if td == nil {
		td = &templateDataBoards{}
	}
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = activeBoardID
	} else {
		//fmt.Println("boardID is empty:", activeBoardID)
		td.ActiveBoardID = 0
	}
	td.AuthenticatedPlayerID = app.session.GetInt(r, "authenticatedPlayerID")
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}


// renderBoards
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


// ----------------------------------------------------------------------------
// STATUS

// Confirm Status
func (app *application) renderConfirmStatus(w http.ResponseWriter, r *http.Request, name string, td *templateDataBattle) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// write to buffer first to catch errors that may occur
	buf := new(bytes.Buffer)
	// execute template set, passing the dynamic data with the copyright year
	err := ts.Execute(buf, app.addDefaultDataBattle(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}


// ----------------------------------------------------------------------------
// JSON

// Return JSON
func (app *application) renderJson(w http.ResponseWriter, r *http.Request, out []byte) {
	// Convert challengerID to string so we can add more strings later
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(out)
	w.Write(out)
}


// ----------------------------------------------------------------------------
// LOGIN

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
	
	// Remove session information
	if app.session != nil {
		app.session.Remove(r, "authenticatedPlayerID")
		app.session.Remove(r, "boardID")
		app.session.Remove(r, "battleID")
		//fmt.Println("Session information has been removed...")
	}

	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}


// ----------------------------------------------------------------------------
// PLAYER

// Add default data to player info screens
func (app *application) addDefaultDataPlayer(td *templateDataPlayer, r *http.Request) *templateDataPlayer {
	if td == nil {
		td = &templateDataPlayer{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}

func (app *application) renderPlayer(w http.ResponseWriter, r *http.Request, name string, td *templateDataPlayer) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = activeBoardID
	} else {
		//fmt.Println("boardID is empty:", activeBoardID)
		td.ActiveBoardID = 0
	}
	td.AuthenticatedPlayerID = app.session.GetInt(r, "authenticatedPlayerID")
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
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


// Add default data to player info screens
func (app *application) addDefaultDataPlayers(td *templateDataPlayers, r *http.Request) *templateDataPlayers {
	if td == nil {
		td = &templateDataPlayers{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
	return td
}

func (app *application) renderPlayers(w http.ResponseWriter, r *http.Request, name string, td *templateDataPlayers) {
	// retrieve based on page name or call serverError helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}
	activeBoardID := app.session.GetInt(r, "boardID")
	if activeBoardID > 0 {
		//fmt.Println("activeBoardID = ", activeBoardID)
		td.ActiveBoardID = activeBoardID
	} else {
		//fmt.Println("boardID is empty:", activeBoardID)
		td.ActiveBoardID = 0
	}
	td.AuthenticatedPlayerID = app.session.GetInt(r, "authenticatedPlayerID")
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)
	td.ScreenName = app.session.GetString(r, "screenName")
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

// ----------------------------------------------------------------------------
// SIGNUP

// Add default data to signup screens
func (app *application) addDefaultDataSignup(td *templateDataSignup, r *http.Request) *templateDataSignup {
	if td == nil {
		td = &templateDataSignup{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	td.IsAuthenticated = app.isAuthenticated(r)
	return td
}

// Render signup form
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


// ----------------------------------------------------------------------------
// ERROR HANDLING

// The serverError helper writes an error message and stack trace to the errorLog
//  - Sends a generic 500 Internal Server Error response to the user
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	if app.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}
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

// ----------------------------------------------------------------------------
// HELPERS

// Random byte string
// Generate a random string n bytes long using these two functions
func (app *application) GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
func (app *application) GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := app.GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
