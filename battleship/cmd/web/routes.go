package main

import (
	"net/http"

	"github.com/bmizerany/pat"		// router
	"github.com/justinas/alice"		// middleware
)

// Update routes to make it return http.Handler rather than *http.ServeMux
func (app *application) routes() http.Handler {
	// a middleware chain using "alice"
	// every request will use this middleware chain
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain to accommodate:
	// a.) session middleware
	// b.) csrf protection (noSurf)
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	strikeMiddleware := alice.New(app.session.Enable, app.authenticate)

	mux := pat.New()
	// More specific routes at the top, less specific routes follow...

	// Basics - home and about
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/about", dynamicMiddleware.ThenFunc(app.about))

	// BATTLES
	// display list of battles
	mux.Get("/status/battles/list", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.listBattles))
	// see if there are any challenges out there
	mux.Get("/status/challenge", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.listBattles))
	// "accept" a challenge from another player
	mux.Post("/status/confirm/:battleID", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.confirmStatus))
	// get list of strikes to see if anything has changed
	mux.Get("/status/strikes/:battleID/:boardID", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.getStrikes))
	// accept a challenge and redirect to /battle/view
	mux.Post("/battle/accept", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.acceptBattle))
	// access a battlefield and continue battling - the battleID will be sent in form post
	mux.Post("/battle/get/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.getBattle))
	// view a non-playable version of a battle - the battleID will be sent in form post
	mux.Post("/battle/view/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.viewBattle))
	mux.Get("/battle/view/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.viewBattle))
	// Enter the battle (shows board with selections that the users can click on)
	mux.Post("/battle/enter/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.enterBattle))
	mux.Post("/battle/strike", strikeMiddleware.Append(app.requireAuthentication).ThenFunc(app.recordStrike))
	//mux.Post("/battle/strike", dynamicMiddleware.ThenFunc(app.recordStrike))

	/*
	mux.HandleFunc("/battle/create", app.createBattle)
	mux.HandleFunc("/battle/list", app.listBattle)
	mux.HandleFunc("/battle/update", app.updateBattle)
	*/
	// BOARDS
	mux.Post("/board/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createBoard))			// save board info
	mux.Get("/board/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createBoardForm))		// display board if GET
	mux.Get("/board/list", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.listBoards))
	mux.Post("/board/select", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.selectBoard))
	mux.Post("/board/update/:id", dynamicMiddleware.ThenFunc(app.updateBoard))
	mux.Get("/board/:id", dynamicMiddleware.ThenFunc(app.displayBoard))
	// PLAYERS
	mux.Post("/player/challenge", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.challengePlayer))
	mux.Get("/player/list", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.listPlayers))
	mux.Get("/player/update/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.updatePlayer))
	mux.Get("/player/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.displayPlayer))
	// POSITIONS
	mux.Get("/position/update/:id", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.updatePosition))
	// SHIPS
	/*
		mux.HandleFunc("/ship", app.selectShip)
		mux.HandleFunc("/ship/list", app.listShip)
	*/
	// AUTH
	mux.Get("/freshstart", alice.New(app.session.Enable).ThenFunc(app.getSignupForm))
	mux.Get("/signup", dynamicMiddleware.ThenFunc(app.getSignupForm))		// display form if GET
	mux.Post("/signup", alice.New(app.session.Enable).ThenFunc(app.postSignup))			// save player info
	mux.Get("/login", dynamicMiddleware.ThenFunc(app.loginForm))
	mux.Post("/login", dynamicMiddleware.ThenFunc(app.postLogin))
	mux.Post("/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.postLogout))
	mux.Post("/updatePlayer", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.updatePlayer))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// remove a specific prefix from the request's URL path
	// before passing the request on to the file server
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// without using "alice"
	//return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	// using "alice"
	return standardMiddleware.Then(mux)
}
