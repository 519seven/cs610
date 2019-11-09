package main

import (
	"context"
	//"errors"
	"fmt"
	"net/http"

	//"github.com/519seven/cs610/battleship/pkg/models"
	"github.com/justinas/nosurf"	// csrf protection

)

// simple function to  add headers to help prevent XSS and Clickjacking attacks
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

// noSurf - prevent CSRF by adding a token to a hidden field in each form
//          and check that the token and cookie info match
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: 	true,
		Path: 		"/",
		Secure:		true,
	})
	return csrfHandler
}

// recoverPanic - unwind the stack when an error occurs
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// recover when the stack unwinds
		defer func() {
			if err := recover(); err != nil {
				// set a connection close header on the response
				w.Header().Set("Connection", "close")
				// call app.serverError to return a 500 response
				// normalize err into an error using Efforf to create
				// a new error object containing the default textual
				// representation of the interface{} value
				// then pass it to serverError helper method
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if authenticatedPlayerID is present; if not, call the next handler
		fmt.Println("checking for authenticatedPlayerID")
		exists := app.session.Exists(r, "authenticatedPlayerID")
		if !exists {
			fmt.Println("authenticatedPlayerID does not exist")
			next.ServeHTTP(w, r)
			return
		}
		fmt.Println("authenticatedPlayerID exists?:", exists)

		// fetch details of current user from database
		// if no matching record was found, remove their session info
		// and call the next handler in the chain as normal
		player, err := app.players.Get(app.session.GetInt(r, "authenticatedPlayerID"))
		if err != nil {
			if app.session != nil {
				app.session.Remove(r, "authenticatedPlayerID")
			}
		}
		fmt.Println("Player ID:", player.ID)

		if player.ID == 0 {					// or no rows in the result set
			fmt.Println("Session has not been established:", err.Error())
			if app.session != nil {
				app.session.Remove(r, "authenticatedPlayerID")
			}
			// if user is invalid, pass the original, unchanged 
			// *http.Request to the next handler in the chain
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}
		fmt.Println("Everything seems normal up to this point...")
		// if the user appears to be active and legit:
		// - create a new copy of the request with a true boolean value added 
		//   to the request context to indicate our satisfaction with their status
		// - call the next handler in the chain using the new copy of the request
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/login", 302)
			return
		}
		// For pages that require authentication, add header
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}