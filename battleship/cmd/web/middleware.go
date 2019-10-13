package main

import (
	"fmt"
	"net/http"
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
