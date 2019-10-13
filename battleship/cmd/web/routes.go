package main

import (
	"net/http"

	"github.com/bmizerany/pat"	// router
	"github.com/justinas/alice"	// middleware
)

// Update routes to make it return http.Handler rather than *http.ServeMux
func (app *application) routes() http.Handler {
	// a middleware chain using "alice"
	// every request will use this middleware chain
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(app.home))
	// More specific routes at the top, less specific routes follow...
	// BATTLES
	/*
		mux.HandleFunc("/battle", app.displayBattle)
		mux.HandleFunc("/battle/create", app.createBattle)
		mux.HandleFunc("/battle/list", app.listBattle)
		mux.HandleFunc("/battle/update", app.updateBattle)
	*/
	// BOARDS
	mux.Post("/board/create", http.HandlerFunc(app.createBoard))		// save board info
	mux.Get("/board/create", http.HandlerFunc(app.createBoardForm))		// display board if GET
	mux.Get("/board/list", http.HandlerFunc(app.listBoard))
	mux.Get("/board/update/:id", http.HandlerFunc(app.updateBoard))
	mux.Get("/board/:id", http.HandlerFunc(app.displayBoard))
	// PLAYERS
	mux.Post("/player/create", http.HandlerFunc(app.createPlayer))		// save player info
	mux.Get("/player/create", http.HandlerFunc(app.createPlayer))		// display form if GET
	mux.Get("/player/list", http.HandlerFunc(app.listPlayer))
	mux.Get("/player/update/:id", http.HandlerFunc(app.updatePlayer))
	mux.Get("/player/:id", http.HandlerFunc(app.displayPlayer))
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
	mux.Post("/logout", http.HandlerFunc(app.logout))
	/*
		mux.HandleFunc("/user/create", app.createUser)
		mux.HandleFunc("/user/list", app.listUser)
		mux.HandleFunc("/user/update", app.updateUser)
	*/
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// remove a specific prefix from the request's URL path
	// before passing the request on to the file server
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// without using "alice"
	//return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	// using "alice"
	return standardMiddleware.Then(mux)
}
