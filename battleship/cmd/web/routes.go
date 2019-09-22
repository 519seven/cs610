package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// Update routes to make it return http.Handler rather than *http.ServeMux
func (app *application) routes() http.Handler {
	// a middleware chain using "alice"
	// every request will use this middleware chain
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	// BATTLES
	/*
		mux.HandleFunc("/battle", app.displayBattle)
		mux.HandleFunc("/battle/create", app.createBattle)
		mux.HandleFunc("/battle/list", app.listBattle)
		mux.HandleFunc("/battle/update", app.updateBattle)
	*/
	// BOARDS
	mux.HandleFunc("/board", app.displayBoard)
	mux.HandleFunc("/board/create", app.createBoard)
	mux.HandleFunc("/board/list", app.listBoard)
	mux.HandleFunc("/board/update", app.updateBoard)
	// PLAYERS
	mux.HandleFunc("/player", app.displayPlayer)
	mux.HandleFunc("/player/create", app.createPlayer)
	mux.HandleFunc("/player/list", app.listPlayer)
	mux.HandleFunc("/player/update", app.updatePlayer)
	// POSITIONS
	/*
		mux.HandleFunc("/position", app.selectPosition)
		mux.HandleFunc("/position/create", app.createPosition)
		mux.HandleFunc("/position/list", app.listPosition)
		// no update needed
	*/
	// SHIPS
	/*
		mux.HandleFunc("/ship", app.selectShip)
		mux.HandleFunc("/ship/list", app.listShip)
	*/
	// AUTH
	mux.HandleFunc("/logout", app.logout)
	/*
		mux.HandleFunc("/user/create", app.createUser)
		mux.HandleFunc("/user/list", app.listUser)
		mux.HandleFunc("/user/update", app.updateUser)
	*/
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// remove a specific prefix from the request's URL path
	// before passing the request on to the file server
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// without using "alice"
	//return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	// using "alice"
	return standardMiddleware.Then(mux)
}
