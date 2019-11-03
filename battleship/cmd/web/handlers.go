package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/xerrors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/519seven/cs610/battleship/pkg/forms"
	"github.com/519seven/cs610/battleship/pkg/models"
)

// BEGIN AUTH
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// Log-In
// ----------------------------------------------------------------------------

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
	form.Required("password")

	rowid, err := app.players.Authenticate(form.Get("screenName"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or password is incorrect")
			app.renderLogin(w, r, "login.page.tmpl", &templateDataLogin{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}
	// Update loggedIn in database
	app.players.UpdateLogin(rowid, true)
	app.session.Put(r, "authenticatedUserID", rowid)
	app.session.Put(r, "screenName", form.Get("screenName"))
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}
// End postLogin

// Begin postLogout
func (app *application) postLogout(w http.ResponseWriter, r *http.Request) {
	// "log out" the user by removing their ID from the session
	rowid := app.session.PopInt(r, "authenticatedUserID")
	app.players.UpdateLogin(rowid, false)
	app.session.Put(r, "flash", "You've been logged out successfully")
	http.Redirect(w, r, "/login", 303)
}
// End postLogout

// ----------------------------------------------------------------------------
// Sign-Up
// ----------------------------------------------------------------------------

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
	form.FieldsMatch("password", "passwordConf", true)

	// If our validation has failed anywhere along the way, redisplay signup form
	if !form.Valid() {
		app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
			Form: form,
		})
		return
	}

	_, err = app.players.Insert(r.PostForm.Get("screenName"), r.PostForm.Get("emailAddress"), r.PostForm.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
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

// END AUTH
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// BEGIN HOME

// Home
// - A method against *application
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

// END HOME
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// BEGIN ABOUT

// ABOUT
// - A method against *application
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	// write board data
	files := []string{
		"./ui/html/about.page.tmpl",
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

// END ABOUT
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// BEGIN BOARDS

// Create a new board
func (app *application) createBoard(w http.ResponseWriter, r *http.Request) {
	// POST /create/board
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	
	// Create a new forms.Form struct containing the POSTed data
	// - Use the validation methods to check the content
	form := forms.New(r.PostForm)
	form.Required("boardName")
	form.MaxLength("boardName", 35)

	// Before returning to the caller, let's check the validity of the ship coordinates
	// - If anything is amiss, we can send those errors back as well
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
	// Loop through the POSTed data, checking for their values
	// - Add coordinates to a given ship's array
    for row := 1; row < 11; row++ {
		rowStr := strconv.Itoa(row)
 		for _, col := range "ABCDEFGHIJ" {
			colStr := string(col)
			shipXY := form.Get("shipXY"+rowStr+colStr)
			if shipXY != "" {
				// Only I, the program, should be permitted to update this as a player enters a game
				//gameID := r.URL.Query().Get("gameID")
				// userID should be gotten from somewhere else
				//userID = r.PostForm("userID")

				// Upper the values to simplify testing
				// - Build the slices containing the submitted coordinates
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
					// Add this to Form's error object?
					// - I don't think it helps to tell the user this info
					//   unless they're struggling to build the board
					fmt.Println("Unsupported character:", shipXY)
				}
			}
		}
	}

	// Test our numbers, update .Valid property of our Form object
	form.RequiredNumberOfItems("carrier", 5, cInd)
	form.RequiredNumberOfItems("battleship", 4, bInd)
	form.RequiredNumberOfItems("cruiser", 3, rInd)
	form.RequiredNumberOfItems("submarine", 3, sInd)
	form.RequiredNumberOfItems("destroyer", 2, dInd)

	form.ValidNumberOfItems(carrier, "carrier")
	form.ValidNumberOfItems(battleship, "battleship")
	form.ValidNumberOfItems(cruiser, "cruiser")
	form.ValidNumberOfItems(submarine, "submarine")
	form.ValidNumberOfItems(destroyer, "destroyer")

	// If our validation has failed anywhere along the way, bail
	if !form.Valid() {
		// helper
		app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard{Form: form})
		return
	}

	// If we've made it to here, then we have a good set of coordinates for a ship
	// - We have a boardID, userID, shipName, and a bunch of coordinates

	// Create a new board, return boardID
	boardID, _ := app.boards.Create(form.Get("boardName"))

	// Carrier
	_, err = app.boards.Insert(boardID, "carrier", carrier)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Battleship
	_, err = app.boards.Insert(boardID, "battleship", battleship)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Cruiser
	_, err = app.boards.Insert(boardID, "cruiser", cruiser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Submarine
	_, err = app.boards.Insert(boardID, "submarine", submarine)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Destroyer
	_, err = app.boards.Insert(boardID, "destroyer", destroyer)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add status message to session data; create new if one doesn't exist
	app.session.Put(r, "flash", "Board successfully created!")
	// Send user back to list of boards
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}

// Form handler
func (app *application) createBoardForm(w http.ResponseWriter, r *http.Request) {
	// GET /create/board
	app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard {
		Form: 				forms.New(nil),
	})
}

// Display board - the way it would appear in a 10x10 grid
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
		PositionsOnBoard: 	s,
	})

}

// List boards
func (app *application) listBoards(w http.ResponseWriter, r *http.Request) {
	// the userID should be in a session somewhere
	userID := app.session.GetInt(r, "authenticatedUserID")
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
		Boards: 			s,
	})
}

// Select
func (app *application) selectBoard(w http.ResponseWriter, r*http.Request) {
	form := forms.New(r.PostForm)
	boardID := form.Get("boardID")
	fmt.Println("submitted boardID:", boardID)
	// make sure my boardID belong to this user
	app.session.Put(r, "boardID", boardID)
	app.session.Put(r, "flash", "Board selected!")
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}

// Update
func (app *application) updateBoard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	boardName := r.URL.Query().Get("boardName")
	userID := 123
	gameID := 1
	id, err = app.boards.Update(id, boardName, userID, gameID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Board successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/board/display/%d", id), http.StatusSeeOther)
}

// END BOARDS
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// BEGIN PLAYERS

// Challenge a Player
func (app *application) challengePlayer(w http.ResponseWriter, r *http.Request) {
	// - The user that challenges is the challenger
	// - The user that is being challenged is the challengee
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// The challengee is selected from a list of players
	challengee := 0
	form := forms.New(r.PostForm)
	userID := form.Get("userID")
	if userID == "" {
		//app.serverError(w, )
		fmt.Println("challengee ID is empty")
		return
	} else {
		challengee, err = strconv.Atoi(userID)
		if err != nil {
			fmt.Println("challengee ID is empty")
			return
		}
	}
	// See if there is a battleID (why???)
	//battleID := app.session.GetInt(r, "battleID")
	//boardID := app.session.GetString(r, "boardID")
	
	// The person who initiated the challenge is the one that has the primary session
	challenger := app.session.GetInt(r, "authenticatedUserID")
	// After creating a battle, return the battleID
	battleID, err := app.battles.Create(challenger, challengee)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.battles.UpdateChallenge(challenger, challengee, false, battleID)

	// If things are successful, return user to player list
	app.session.Put(r, "flash", "Challenge created!  Awaiting player's response.")
	http.Redirect(w, r, "/player/list", http.StatusSeeOther)
}

// Display player
func (app *application) displayPlayer(w http.ResponseWriter, r *http.Request) {
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
		Player: 			s,
	})
}

// List players
func (app *application) listPlayers(w http.ResponseWriter, r *http.Request) {
	p, err := app.players.List("loggedIn")
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderPlayers(w, r, "players.page.tmpl", &templateDataPlayers{
		Players: 			p,
	})
}

// Update player
func (app *application) updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	screenName := r.URL.Query().Get("boardName")
	id, err = app.players.Update(id, screenName)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Player successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/player/%d", id), http.StatusSeeOther)
}

// END PLAYERS
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// BEGIN POSITIONS

// Update a position's pinColor
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
	app.session.Put(r, "flash", fmt.Sprintf("Position (id #%d) has been updated...", rowid))
	http.Redirect(w, r, fmt.Sprintf("/player/list/%d", rowid), http.StatusSeeOther)
}

// END POSITIONS
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// BEGIN STATUS

// Update a position's pinColor
func (app *application) challengeStatus(w http.ResponseWriter, r *http.Request) {
	userID := app.session.GetInt(r, "authenticatedUserID")	// get from session??
	challengerID, err := app.battles.GetChallenger(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if challengerID != 0 {
		p, err := app.players.Get(challengerID)
		if err != nil {
			if xerrors.Is(err, models.ErrNoRecord) {
				app.notFound(w)
			} else {
				app.serverError(w, err)
			}
			return
		}
		// Create session key of "challenger" and redirect to list where you'd make a selection Yes/No
		app.session.Put(r, "challenger", p.ScreenName)
		//app.session.Put(r, "challenger", p.ScreenName)
		//app.session.Put(r, "challenger", p.ScreenName)
		type JsonResponse struct {
			ChallengerID 	int 		`json:"challengerID"`
			NextPage		string		`json:"redirect"`
			Time			time.Time	`json:"timestamp"`
		}
		j := fmt.Sprintf(`{"challengerID":%d, "redirect":"/status/confirm"}`, challengerID)
		bytes := []byte(j)

		var JR JsonResponse
		JR.Time = time.Now()
		err = json.Unmarshal(bytes, &JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		out, err := json.Marshal(JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.renderJson(w, r, out)
	} else {
		// do nothing, there is no challenger
		return
	}
}

func (app *application) confirmStatus(w http.ResponseWriter, r *http.Request) {
	battleID, err := strconv.Atoi(r.URL.Query().Get(":battleID"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	b, err := app.battles.Get(battleID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderConfirmStatus(w, r, "status.confirm.page.tmpl", &templateDataBattle{
		Battle: 			b,
	})
}

// END STATUS
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------