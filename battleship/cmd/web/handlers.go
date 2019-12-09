package main

// The handlers are the functions that the application routes will use to field a request

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
// Log-In
// ----------------------------------------------------------------------------

// Display login using a new form
func (app *application) loginForm(w http.ResponseWriter, r *http.Request) {
	app.renderLogin(w, r, "login.page.tmpl", &templateDataLogin {
		Form: 			forms.New(nil),
	})
}


// Check login request to see if user can be authenticated
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
			form.Errors.Add("generic", "Supplied credentials are incorrect")
			app.renderLogin(w, r, "login.page.tmpl", &templateDataLogin{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}
	// Update loggedIn in database
	app.players.UpdateLogin(rowid, true)
	app.session.Put(r, "authenticatedPlayerID", rowid)
	app.session.Put(r, "screenName", form.Get("screenName"))
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}


// Handle logout request; remove authenticatedPlayerID from session object
func (app *application) postLogout(w http.ResponseWriter, r *http.Request) {
	// "log out" the user by removing their ID from the session
	rowid := app.session.PopInt(r, "authenticatedPlayerID")
	app.players.UpdateLogin(rowid, false)
	app.session.Put(r, "flash", "You've been logged out successfully")
	http.Redirect(w, r, "/login", 303)
}


// ----------------------------------------------------------------------------
// Sign-Up
// ----------------------------------------------------------------------------

// Display new player form
func (app *application) getSignupForm(w http.ResponseWriter, r *http.Request) {
	app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
		Form: forms.New(nil),
	})
}


// Create a new player - submit signup form (POST)
func (app *application) postSignup(w http.ResponseWriter, r *http.Request) {
	// Create a new forms.Form struct containing the POSTed data from the
	//  form, then use the validation methods to check the content.
	fmt.Println("here...")
	err := r.ParseForm()
	if err != nil {
		fmt.Println("Form is empty")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("screenName")
	form.MaxLength("screenName", 16)
	form.SpacesAbsent("screenName")
	form.Required("password")
	form.MaxLength("password", 55)
	form.MinLength("password", 8)
	form.Required("passwordConf")
	form.FieldsMatch("password", "passwordConf", true)
	form.PasswordComplexity()

	// If our validation has failed anywhere along the way, redisplay signup form
	if !form.Valid() {
		app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup {
			Form: form,
		})
		return
	}

	_, err = app.players.Insert(r.PostForm.Get("screenName"), "", r.PostForm.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateScreenName) {
			form.Errors.Add("screenName", "Screen name is already in use")
			app.renderSignup(w, r, "signup.page.tmpl", &templateDataSignup{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")
	app.session.Remove(r, "authenticatedPlayerID")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}


// END AUTH
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// BEGIN HOME

// Display / (home) page
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

// ----------------------------------------------------------------------------
// BEGIN ABOUT

// Display the "about" page
func (app *application) about(w http.ResponseWriter, r *http.Request) {

	app.renderAbout(w, r, "about.page.tmpl", &templateDataAbout{})
}

// END ABOUT
// ----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// BEGIN BATTLES

// Accept battle (challenge) and redirect to view the battlefield
func (app *application) acceptBattle(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	form := forms.New(r.PostForm)
	battleID, err := strconv.Atoi(form.Get("battleID"))
	if !form.Valid() || err != nil {
		app.session.Put(r, "flash", "Unexplained error!")
		app.serverError(w, errors.New("Invalid form structure"))
		return
	}
	_, err = app.battles.Accept(playerID, app.session.GetInt(r, "boardID"), battleID)
	if err != nil {
		app.session.Put(r, "flash", "The person who accepted this board was not the person challenged!")
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "You have accepted the battle!")
	http.Redirect(w, r, fmt.Sprintf("/battle/view/%d", battleID), http.StatusSeeOther)
}


// Enter battle
func (app *application) enterBattle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entering the battlefield...")
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	battleID, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || battleID < 1 {
		app.infoLog.Println("Invalid boardID")
		app.notFound(w)
		return
	}
	// Get general information about the battle
	app.infoLog.Println("Getting general information about the battle")
	b, err := app.battles.Get(playerID, battleID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
	}
	// Get positions for Player1's ships (by boardID)
	c, err := app.boards.GetPositions(int(b.Player1BoardID))
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	// Get positions for Player2's ships (by boardID)
	o, err := app.boards.GetPositions(int(b.Player2BoardID))
	if err != nil {
		if xerrors.Is(err, models.ErrMissingBoardID) {
			app.session.Put(r, "flash", "Missing BoardID - Challenge cannot continue")
			http.Redirect(w, r, "/status/battles/list", http.StatusSeeOther)		
		} else if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.renderBattle(w, r, "enter.battle.page.tmpl", &templateDataBattle{
		Battle:							b,
		ChallengerID:					int(b.Player1ID),
		ChallengerBoardID:				int(b.Player1BoardID),
		ChallengerPositions: 			c,
		OpponentID:						int(b.Player2ID),
		OpponentBoardID:				int(b.Player2BoardID),
		OpponentPositions:				o,
	})
}


// Get battle
func (app *application) getBattle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Still to be done - The battlefield, which will have two boards and an active battle..."))
}


// List battles
func (app *application) listBattles(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	b, err := app.battles.GetChallenges(playerID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	for _, battle := range b {
		battle.AuthenticatedPlayerID = playerID
	}
	app.renderBattles(w, r, "list.challenges.page.tmpl", &templateDataBattles{
		Battles: 			b,
	})
}


// View battle
func (app *application) viewBattle(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	battleID, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || battleID < 1 {
		app.infoLog.Println("Invalid boardID")
		app.notFound(w)
		return
	}
	// Get general information about the battle
	b, err := app.battles.Get(playerID, battleID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
	}
	// Get postitions for Player1's ships (by boardID)
	//fmt.Println("challengerBoardID:", int(b.Player1BoardID))
	c, err := app.boards.GetPositions(int(b.Player1BoardID))
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	// Get postitions for Player2's ships (by boardID)
	o, err := app.boards.GetPositions(int(b.Player2BoardID))
	if err != nil {
		if xerrors.Is(err, models.ErrMissingBoardID) {
			app.session.Put(r, "flash", "Missing BoardID - Challenge cannot continue")
			http.Redirect(w, r, "/status/battles/list", http.StatusSeeOther)		
		} else if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderBattle(w, r, "display.battle.page.tmpl", &templateDataBattle{
		Battle:							b,
		ChallengerID:					int(b.Player1ID),
		ChallengerBoardID:				int(b.Player1BoardID),
		ChallengerPositions: 			c,
		OpponentID:						int(b.Player2ID),
		OpponentBoardID:				int(b.Player2BoardID),
		OpponentPositions:				o,
	})
}

// END BATTLES
// ----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// BEGIN BOARDS

// Create a new game board
func (app *application) createBoard(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
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
				//battleID := r.URL.Query().Get("battleID")
				// playerID should be gotten from somewhere else
				//playerID = r.PostForm("playerID")

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

	// If our validation has failed anywhere along the way
	// - Take the user back to their board
	if !form.Valid() {
		// helper
		app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard{Form: form})
		return
	}

	// If we've made it to here, then we have a good set of coordinates for a ship
	// - We have a boardID, playerID, shipName, and a bunch of coordinates

	// Create a new board, return boardID
	boardID, _ := app.boards.Create(playerID, form.Get("boardName"))

	// Carrier
	_, err = app.boards.Insert(playerID, boardID, "carrier", carrier)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Battleship
	_, err = app.boards.Insert(playerID, boardID, "battleship", battleship)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Cruiser
	_, err = app.boards.Insert(playerID, boardID, "cruiser", cruiser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Submarine
	_, err = app.boards.Insert(playerID, boardID, "submarine", submarine)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Destroyer
	_, err = app.boards.Insert(playerID, boardID, "destroyer", destroyer)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add status message to session data; create new if one doesn't exist
	app.session.Put(r, "flash", "Board successfully created!")
	// Send user back to list of boards
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}


// Display new form for creating a game board
func (app *application) createBoardForm(w http.ResponseWriter, r *http.Request) {
	app.renderBoard(w, r, "create.board.page.tmpl", &templateDataBoard {
		Form: 				forms.New(nil),
	})
}


// Display board - the way it would appear in a 10x10 grid
func (app *application) displayBoard(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	boardID, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || boardID < 1 {
		app.infoLog.Println("Invalid boardID")
		app.notFound(w)
		return
	}
	//app.infoLog.Println("boardID: ", boardID)
	p, err := app.boards.GetPositions(boardID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	b, err := app.boards.GetInfo(playerID, boardID)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
	}
	app.renderBoard(w, r, "display.board.page.tmpl", &templateDataBoard{
		Positions: 			p,
		Board:				b,
	})
}


// List boards
func (app *application) listBoards(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	//boardID := app.session.GetInt(r, "boardID")							// debug
	//app.infoLog.Println("boardID immediately after setting is:", boardID)	// debug
	s, err := app.boards.List(playerID)
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


// Select a board; needed before challenging an opponent to a battle
func (app *application) selectBoard(w http.ResponseWriter, r*http.Request) {
	form := forms.New(r.PostForm)
	boardID, err := strconv.Atoi(form.Get("boardID"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Board selected!")

	app.session.Remove(r, "boardID")
	app.session.Put(r, "boardID", boardID)
	app.infoLog.Println(app.session.GetInt(r, "boardID"))
	http.Redirect(w, r, "/board/list", http.StatusSeeOther)
}


// Update board information
func (app *application) updateBoard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	boardName := r.URL.Query().Get("boardName")
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	app.errorLog.Println("Fail")
	id, err = app.boards.Update(id, boardName, playerID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Board successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/board/display/%d", id), http.StatusSeeOther)
}

// END BOARDS
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// BEGIN PLAYERS

// Challenge a Player
func (app *application) challengePlayer(w http.ResponseWriter, r *http.Request) {
	player1ID := 0
	player2ID := 0
	secretTurn, err := app.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Player1 is the challenger; Player2 is the opponent
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Player1 information is retrieved from session object
	player1ID = app.session.GetInt(r, "authenticatedPlayerID")
	player1BoardID := app.session.GetInt(r, "boardID")
	app.infoLog.Println("Player1 boardID is", player1BoardID)
	if player1BoardID < 1 {
		app.session.Put(r, "flash", "You must select your board first, then issue a challenge!")
		http.Redirect(w, r, "/board/list", http.StatusSeeOther)
	}

	// Player2 information is retrieved from form
	form := forms.New(r.PostForm)
	if form.Get("playerID") == "" {
		//app.serverError(w, )
		app.infoLog.Println("playerID (from players list) is empty")
		return
	} else {
		player2ID, err = strconv.Atoi(form.Get("playerID"))
		if err != nil {
			app.infoLog.Println("player2ID is empty")
			return
		}
	}
	_, err = app.battles.Create(player1ID, player1BoardID, player2ID, secretTurn)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// This "update" now happens in "Create" - not ideal!
	//app.battles.UpdateChallenge(player1ID, player2ID, false, battleID)

	// Return user to player list where the flash message is displayed
	app.session.Put(r, "flash", "Challenge created!  Awaiting player's response.")
	http.Redirect(w, r, "/player/list", http.StatusSeeOther)
}


// Find out if there are any challenges for the user
func (app *application) challengeStatus(w http.ResponseWriter, r *http.Request) {
	// If we have a challenge, return JSON to client
	playerID := app.session.GetInt(r, "authenticatedPlayerID")	// get from session??
	challengerID, err := app.battles.GetChallenger(playerID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if challengerID > 0 {
		type JsonResponse struct {
			Status 			string 		`json:"status"`
			NextPage		string		`json:"redirect"`
			Time			time.Time	`json:"timestamp"`
		}
		redirect := "/status/battles/list"
		var JR JsonResponse
		JR.Status = "challenge"
		JR.NextPage = redirect
		JR.Time = time.Now()

		out, err := json.Marshal(JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.renderJson(w, r, out)
	} else {
		// do nothing
		return
	}
}


// Accept a challenge from another player
func (app *application) confirmStatus(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	b, err := app.battles.GetOpen(playerID, 0)
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderConfirmStatus(w, r, "status.confirm.page.tmpl", &templateDataBattle{
		Battles: 			b,
	})
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


// Display a list of players
func (app *application) listPlayers(w http.ResponseWriter, r *http.Request) {
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	p, err := app.players.List(playerID, "loggedIn")
	if err != nil {
		if xerrors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.renderPlayers(w, r, "list.players.page.tmpl", &templateDataPlayers{
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

// ----------------------------------------------------------------------------
// BEGIN POSITIONS

// Update a position's pinColor
func (app *application) updatePosition(w http.ResponseWriter, r *http.Request) {
	newSecret, err := app.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)
	}
	battleID, err := strconv.Atoi(r.URL.Query().Get("battleID"))
	if err != nil {
		app.serverError(w, err)
	}
	boardID, err := strconv.Atoi(r.URL.Query().Get("boardID"))
	if err != nil {
		app.serverError(w, err)
	}
	playerID := app.session.GetInt(r, "authenticatedPlayerID")

	shipXY := r.URL.Query().Get("shipXY")
	coordX, err := strconv.Atoi(shipXY[len(shipXY)-2:])		// get the second-to-last character (!!!BUG!!!)
	if err != nil {
		app.serverError(w, err)
	}
	coordY := shipXY[len(shipXY)-1:]						// get the last character
	pinColor, _, winner, err := app.positions.Update(playerID, playerID, battleID, boardID, coordX, coordY, newSecret)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", fmt.Sprintf("Position (id #%s) has been updated...", pinColor))
	if winner { app.session.Put(r, "winner", "A winner has been declared!") }
	http.Redirect(w, r, fmt.Sprintf("/player/list/%s", pinColor), http.StatusSeeOther)
}

// END POSITIONS
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// BEGIN STRIKE

// Get a list of strikes for the authenticated user's board
// - This information is used to update the authenticated user's board
func (app *application) getStrikes(w http.ResponseWriter, r *http.Request) {
	battleID, err := strconv.Atoi(r.URL.Query().Get(":battleID"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Need to check battleID against the battleID stored in the session
	sessionBattleID := app.session.GetInt(r, "battleID")
	if battleID == sessionBattleID {
		//app.serverError(w, err)
		//return
		fmt.Println("battleID does not match session's battleID...FAIL!")
	}
	// This board ID ought to be the authenticated user's board
	boardID, err := strconv.Atoi(r.URL.Query().Get(":boardID"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	// Does this battle include this player and this board?
	// - This is a simple double-check
	if app.battles.CheckBoardOwner(playerID, battleID, boardID) {
		// If we are who we say we are...
		// - Return a list of strikes for the opponent's board
		type JsonResponse struct {
			Turn 			string					`json:"turn"`
			Positions		[]*models.Position		`json:"strikes"`
			Winner			string					`json:"winner"`
		}
		var JR JsonResponse

		_, secretTurn := app.battles.CheckTurn(battleID, playerID)
		JR.Turn = secretTurn
		if err != nil {
			app.serverError(w, err)
			app.errorLog.Println("Error", err.Error())
			return
		}
		positions, err := app.positions.List(boardID, playerID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		JR.Positions = positions
		winner, err := app.positions.CheckWinner(playerID, boardID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if winner { JR.Winner = "winner" } else { JR.Winner = "chicken_dinner" }

		out, err := json.Marshal(JR)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.renderJson(w, r, out)
	} else {
		// do nothing
		app.serverError(w, errors.New(fmt.Sprintf("The challenger (%d) doesn't match the owner of the board or the participant in this battle (%d)", playerID, boardID)))
		return
	}
}


// When a player launches a strike, see if it is a hit (make pinColor=1) and record strike
func (app *application) recordStrike(w http.ResponseWriter, r *http.Request) {
	type PlayerTurn struct {
		Valid 			bool			`json:"valid"`
		PinColor		string			`json:"pin_color"`
		ShipType		string			`json:"sunken_ship"`
		Winner			bool			`json:"winner"`
	}

	err := r.ParseForm()
	form := forms.New(r.PostForm)
	battleID, _ := strconv.Atoi(form.Get("battleID"))
	boardID, _ := strconv.Atoi(form.Get("boardID"))
	coordX, _ := strconv.Atoi(form.Get("coordX"))
	coordY := form.Get("coordY")
	//fmt.Println("coordX:", coordX, "|coordY:", coordY)
	//fmt.Println("battleID:", battleID)
	//fmt.Println("boardID:", boardID)
	playerID := app.session.GetInt(r, "authenticatedPlayerID")
	secretTurn := form.Get("secretTurn")
	var out []byte;
	var PT PlayerTurn

	playerTakingTheirTurn, validTurn := app.checkTurn(playerID, battleID, secretTurn)
	if validTurn {
		// Make sure this player belongs to this battle
		//checkBattle(playerID, battleID)
		// Update the database with the new strike, update Turn to be the other player
		newSecret, err := app.GenerateRandomString(32)
		if err != nil {
			app.serverError(w, err)
		}
		pinColor, shipType, winner, err := app.positions.Update(playerID, playerTakingTheirTurn, battleID, boardID, coordX, coordY, newSecret)
		if err != nil {
			app.infoLog.Println("Update failed for ", playerID, battleID, boardID, coordX, coordY)
			app.errorLog.Println("Error", err.Error())
		}
		//fmt.Println("pinColor is", pinColor)
		PT.Valid = true
		PT.PinColor = pinColor
		PT.ShipType = shipType
		PT.Winner = winner
		out, err = json.Marshal(PT)
		if err != nil {
			app.serverError(w, err)
		}
	} else {
		app.errorLog.Println("Looks like somebody is attempting to hack or the person submitting the strike is not in sync with the database...")
		PT.Valid = false
		PT.PinColor = ""
		PT.ShipType = "charlie"
		PT.Winner = false
		out, err = json.Marshal(PT)
		if err != nil {
			app.serverError(w, err)
		}
	}
	app.renderJson(w, r, out)
}

// END STRIKE
// ----------------------------------------------------------------------------
// -----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// BEGIN TURN

// Given a key and playerID, check if they match what's in the database
func (app *application) checkTurn(playerID, battleID int, secretTurn string) (int, bool) {
	// Make sure the player that submitted this strike is the one whose turn it is
	// Check the secret_turn against the player's ID
	playerWhoTookTheirTurn, databaseSecret := app.battles.CheckTurn(battleID, playerID)
	if string(databaseSecret) == secretTurn && secretTurn != "" {
		return playerWhoTookTheirTurn, true
	}
	return 0, false
}
