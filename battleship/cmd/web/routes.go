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

	// Create a new middleware chain to accommodate our session middleware
	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	// More specific routes at the top, less specific routes follow...
	// BATTLES
	/*
		mux.HandleFunc("/battle", app.displayBattle)
		mux.HandleFunc("/battle/create", app.createBattle)
		mux.HandleFunc("/battle/list", app.listBattle)
		mux.HandleFunc("/battle/update", app.updateBattle)
	*/
	// Update routes to use the new dynamic middleware chain for our session middleware
	
	// BOARDS
	mux.Post("/board/create", dynamicMiddleware.ThenFunc(app.createBoard))			// save board info
	mux.Get("/board/create", dynamicMiddleware.ThenFunc(app.createBoardForm))		// display board if GET
	mux.Get("/board/list", dynamicMiddleware.ThenFunc(app.listBoards))
	mux.Get("/board/update/:id", dynamicMiddleware.ThenFunc(app.updateBoard))
	mux.Get("/board/:id", dynamicMiddleware.ThenFunc(app.displayBoard))
	// PLAYERS
	mux.Get("/player/list", dynamicMiddleware.ThenFunc(app.listPlayers))
	mux.Get("/player/update/:id", dynamicMiddleware.ThenFunc(app.updatePlayer))
	mux.Get("/player/:id", dynamicMiddleware.ThenFunc(app.displayPlayer))
	// POSITIONS
	mux.Get("/position/update/:id", dynamicMiddleware.ThenFunc(app.updatePosition))
	// SHIPS
	/*
		mux.HandleFunc("/ship", app.selectShip)
		mux.HandleFunc("/ship/list", app.listShip)
	*/
	// AUTH
	mux.Get("/signup", dynamicMiddleware.ThenFunc(app.getSignupForm))		// display form if GET
	mux.Post("/signup", dynamicMiddleware.ThenFunc(app.postSignup))			// save player info
	mux.Get("/login", dynamicMiddleware.ThenFunc(app.loginForm))
	mux.Post("/login", dynamicMiddleware.ThenFunc(app.postLogin))
	mux.Post("/logout", http.HandlerFunc(app.postLogout))
	mux.Post("/updatePlayer", dynamicMiddleware.ThenFunc(app.updatePlayer))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// remove a specific prefix from the request's URL path
	// before passing the request on to the file server
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// without using "alice"
	//return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	// using "alice"
	return standardMiddleware.Then(mux)
}
