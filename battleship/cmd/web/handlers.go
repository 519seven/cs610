package main

import (
	"github.com/519seven/cs610/battleship/pkg/forms"
	"github.com/519seven/cs610/battleship/pkg/models"
	"bytes"
	"golang.org/x/xerrors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// landing page
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
				fmt.Println("Getting the value at", "shipXY_"+rowStr+"_"+colStr)
				fmt.Println("That value is", shipXY)
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
		"| destroyer_status: ", destroyer_status)	// remove after debugging

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
func (app *application) listBoard(w http.ResponseWriter, r *http.Request) {
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

// -----------------------------------------------------------------------------
// Players

// create a new player
func (app *application) createPlayer(w http.ResponseWriter, r *http.Request) {
	// restrict this handler to HTTP POST methods only
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	screenName := r.URL.Query().Get("screenName")
	id, err := app.players.Insert(screenName)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/player/list/%d", id), http.StatusSeeOther)
}

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

	app.renderPlayers(w, r, "players.page.tmpl", &templateDataPlayer{
		Player: s,
	})
}

// list players
func (app *application) listPlayer(w http.ResponseWriter, r *http.Request) {
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
	app.renderPlayers(w, r, "players.page.tmpl", &templateDataPlayer{
		Players: s,
	})
}

// update player (POST or GET???)
func (app *application) updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	screenName := r.URL.Query().Get("boardName")
	id, err = app.players.Update(id, screenName)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/player/display/%d", id), http.StatusSeeOther)
}

// -----------------------------------------------------------------------------
// Auth

// Log out
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}
