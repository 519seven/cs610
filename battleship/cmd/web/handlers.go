package main

import (
	"github.com/519seven/cs610/battleship/pkg/forms"
	"github.com/519seven/cs610/battleship/pkg/models"
	"bytes"
	"errors"
	"golang.org/x/xerrors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// Auth

// Display new player form
func (app *application) getSignupForm(w http.ResponseWriter, r *http.Request) {
	app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
		Form: forms.New(nil),
	})
}
// End getSignupForm

// Create a new player - submit signup form (POST)
// Begin postSignup
func (app *application) postSignup(w http.ResponseWriter, r *http.Request) {
	// Create a new forms.Form struct containing the POSTed data from the
	//  form, then use the validation methods to check the content.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("screenName")
	form.MaxLength("screenName", 16)
	form.Required("emailAddress")
	form.MaxLength("emailAddress", 55)
	form.MatchesPattern("emailAddress", forms.EmailRX)
	form.Required("password")
	form.MaxLength("password", 55)
	form.MinLength("password", 8)
	form.Required("passwordConf")

	// If our validation has failed anywhere along the way, redisplay signup form
	if !form.Valid() {
		app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
			Form: form,
		})
		return
	}

	_, err = app.players.Insert(form.Get("screenName"), form.Get("emailAddress"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("emailAddress", "Address is already in use")
			app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
// End postSignup

// Begin loginForm
func (app *application) loginForm(w http.ResponseWriter, r *http.Request) {
	app.renderLogin(w, r, "login.page.tmpl", &templateDataLogin {
		Form: forms.New(nil),
	})
}
// End loginForm

// Begin postLogin
func (app *application) postLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("screenName")
	form.MaxLength("screenName", 16)
	//form.Required("emailAddress")
	//form.MaxLength("emailAddress", 55)
	//form.MatchesPattern("emailAddress", forms.EmailRX)

	// If our validation has failed anywhere along the way, redisplay signup form
	if !form.Valid() {
		app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
			Form: form,
		})
		return
	}

	_, err = app.players.Authenticate(form.Get("screenName"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.renderLogin(w, r, "login.page.tmpl", &templateDataLogin{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}
// End postLogin

// Begin postLogout
func (app *application) postLogout(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUserID")
	// Add a flash message to the session to confirm to the user that they've been // logged out.
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", 303)
}
// End postLogout

// End Auth
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// Home

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// write board data
	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}
	// render template
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// to catch template errors, write to buffer first
	buf := new(bytes.Buffer)
	err = ts.Execute(buf, nil)
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

// End Home
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// Boards

// create a new board
func (app *application) createBoard(w http.ResponseWriter, r *http.Request) {
	// our router, pat, takes care of this for us now...
	/*
	// One example: allow POST method only
	if r.Method != http.MethodPost {
		// Change response header map before WriteHeader or Write
		w.Header().Set("Allow", http.MethodPost)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// WriteHeader must be called explicity before any call to Write
		//w.WriteHeader(405)
		//w.Write([]byte("Method Not Allowed\r\n"))
		// http.Error handles both WriteHeader and Write
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	// Another example: allow GET method only
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	*/
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	
	// Create a new forms.Form struct containing the POSTed data from the
	//  form, then use the validation methods to check the content.
	form := forms.New(r.PostForm)
	form.Required("boardName")
	form.MaxLength("boardName", 35)

	var carrier []string
	cInd := 0
	var battleship []string
	bInd := 0
	var cruiser []string
	rInd := 0
	var submarine []string
	sInd := 0
	var destroyer []string
	dInd := 0
	// loop through the board's text fields, checking for their values
	// add coordinates to a given ship's array
    for row := 1; row < 11; row++ {
		rowStr := strconv.Itoa(row)
 		for _, col := range "ABCDEFGHIJ" {
			colStr := string(col)
			shipXY := r.PostForm.Get("shipXY"+rowStr+colStr)
			if shipXY != "" {
				// Only I, the program, should be permitted to update this as a player enters a game
				//gameID := r.URL.Query().Get("gameID")
				// userID should be gotten from somewhere else
				//userID = r.PostForm("userID")
				//fmt.Println("Getting the value at", "shipXY_"+rowStr+"_"+colStr)
				//fmt.Println("That value is", shipXY)
				switch strings.ToUpper(shipXY) {
				case "C":
					carrier = append(carrier, rowStr+","+colStr)
					cInd += 1
				case "B":
					battleship = append(battleship, rowStr+","+colStr)
					bInd += 1
				case "R":
					cruiser = append(cruiser, rowStr+","+colStr)
					rInd += 1
				case "S":
					submarine = append(submarine, rowStr+","+colStr)
					sInd += 1
				case "D":
					destroyer = append(destroyer, rowStr+","+colStr)
					dInd += 1
				default:
					fmt.Println("Unsupported character:", shipXY)
				}
			}
		}
	}
	carrier_status := 0
	battleship_status := 0
	cruiser_status := 0
	submarine_status := 0
	destroyer_status := 0

	// Test our numbers
	form.RequiredNumberOfItems("carrier", 5, cInd)
	form.RequiredNumberOfItems("battleship", 4, bInd)
	form.RequiredNumberOfItems("cruiser", 3, rInd)
	form.RequiredNumberOfItems("submarine", 3, sInd)
	form.RequiredNumberOfItems("destroyer", 2, dInd)

	// If our validation has failed anywhere along the way, bail
	if !form.Valid() {
		app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard{Form: form})
		return
	}

	// Create a new board, return boardID
	boardID, _ := app.boards.Create(form.Get("boardName"))

	// Carrier
	carrier_status, err = app.boards.Insert(boardID, "carrier", carrier)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Battleship
	battleship_status, err = app.boards.Insert(boardID, "battleship", battleship)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Cruiser
	cruiser_status, err = app.boards.Insert(boardID, "cruiser", cruiser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Submarine
	submarine_status, err = app.boards.Insert(boardID, "submarine", submarine)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Destroyer
	destroyer_status, err = app.boards.Insert(boardID, "destroyer", destroyer)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Println("carrier_status: ", carrier_status, "| battleship_status: ", battleship_status, 
		"| cruiser_status: ", cruiser_status, "| submarine_status: ", submarine_status, 
		"| destroyer_status: ", destroyer_status)	// debugging

	app.session.Put(r, "flash", "Board successfully created!")
	// Send user back to list of boards
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}

// form handler
func (app *application) createBoardForm(w http.ResponseWriter, r *http.Request) {
	app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard {
		Form: forms.New(nil),
	})
}

// display board - the way it would appear in a 10x10 grid
func (app *application) displayBoard(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || boardID < 1 {
		app.notFound(w)
		return
	}
	fmt.Println("boardID: ", boardID)
	s, err := app.boards.Get(boardID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard{
		PositionsOnBoard: s,
	})

}

// list boards
func (app *application) listBoards(w http.ResponseWriter, r *http.Request) {
	// the userID should be in a session somewhere
	userID := 1
	if userID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.boards.List(userID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.renderBoards(w, r, "list.boards.page.tmpl", &templateDataBoards{
		Boards: s,
	})
}

// update
func (app *application) updateBoard(w http.ResponseWriter, r *http.Request) {
	// restrict this handler to HTTP POST methods only
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	boardName := r.URL.Query().Get("boardName")
	userID := 123
	gameID := 1
	id, err = app.boards.Update(id, boardName, userID, gameID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board/display/%d", id), http.StatusSeeOther)
}

// End Boards
// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// Players

// display player
func (app *application) displayPlayer(w http.ResponseWriter, r *http.Request) {
	// Allow GET method only
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	playerID, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || playerID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.players.Get(playerID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.renderPlayer(w, r, "player.page.tmpl", &templateDataPlayer{
		Player: s,
	})
}

// list players
func (app *application) listPlayers(w http.ResponseWriter, r *http.Request) {
	// the userID should be in a session somewhere
	userID := 123
	if userID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.players.List()
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderPlayers(w, r, "players.page.tmpl", &templateDataPlayers{
		Players: s,
	})
}

// update player
func (app *application) updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	screenName := r.URL.Query().Get("boardName")
	id, err = app.players.Update(id, screenName)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/player/%d", id), http.StatusSeeOther)
}

// End Players
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// Position

// update a position's pinColor
func (app *application) updatePosition(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.URL.Query().Get("boardID"))	// get from session?
	if err != nil {
		app.serverError(w, err)
	}
	shipXY := r.URL.Query().Get("shipXY")						// get from board form
	playerID := 1												// get from session??
	coordX, err := strconv.Atoi(shipXY[len(shipXY)-2:])			// get the second-to-last character
	if err != nil {
		app.serverError(w, err)
	}
	coordY := shipXY[len(shipXY)-1:]							// get the last character
	pinColor := 1												// if it's 1 then 0; if it's 0 then 1
	rowid, err := app.positions.Update(boardID, playerID, coordX, coordY, pinColor)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintln(w, "Position (id #", rowid, ") has been updated...")
	//http.Redirect(w, r, fmt.Sprintf("/player/list/%d", id), http.StatusSeeOther)
}

// End Position
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------