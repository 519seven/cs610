package main

import (
	"github.com/519seven/cs610/battleship/pkg/models"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// landing page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

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
	// restrict this handler to HTTP POST methods only
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
	boardName := r.URL.Query().Get("boardName")
	userID, err := strconv.Atoi(r.URL.Query().Get("userID"))
	// Only I, the program, should be permitted to update this as a player enters a game
	//gameID := r.URL.Query().Get("gameID")
	id, err := app.boards.Insert(boardName, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/board?id=%d", id), http.StatusSeeOther)
}

// display board - the way it would appear in a 10x10 grid
func (app *application) displayBoard(w http.ResponseWriter, r *http.Request) {
	// Allow GET method only
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	boardID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || boardID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.boards.Get(boardID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.renderBoards(w, r, "boards.page.tmpl", &templateDataBoard{
		Board: s,
	})

}

// list boards
func (app *application) listBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // Allow GET method only
		w.Header().Set("Allow", http.MethodGet)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	// the userID should be in a session somewhere
	userID := 123
	if userID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.boards.List(userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.renderBoards(w, r, "boards.page.tmpl", &templateDataBoard{
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
	http.Redirect(w, r, fmt.Sprintf("/board/display?id=%d", id), http.StatusSeeOther)
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
	http.Redirect(w, r, fmt.Sprintf("/player?id=%d", id), http.StatusSeeOther)
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
	playerID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || playerID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.players.Get(playerID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
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
// Get a list of boards belonging to this user
func (app *application) listPlayer(w http.ResponseWriter, r *http.Request) {
	// Allow GET method only
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	// the userID should be in a session somewhere
	userID := 123
	if userID < 1 {
		app.notFound(w)
		return
	}
	s, err := app.players.List()
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
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

// update
func (app *application) updatePlayer(w http.ResponseWriter, r *http.Request) {
	// restrict this handler to HTTP POST methods only
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	screenName := r.URL.Query().Get("boardName")
	id, err = app.players.Update(id, screenName)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/player/display?id=%d", id), http.StatusSeeOther)
}

// -----------------------------------------------------------------------------
// Auth

// Log out
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}
