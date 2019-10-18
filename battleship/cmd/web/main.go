package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/519seven/cs610/battleship/pkg/models/sqlite3"
	"github.com/golangcollege/sessions"
)

// custom application struct
// this makes objects available to our handlers
type application struct {
	battles       	*sqlite3.BattleModel
	boards        	*sqlite3.BoardModel
	errorLog      	*log.Logger
	infoLog       	*log.Logger
	players       	*sqlite3.PlayerModel
	positions     	*sqlite3.PositionModel
	session			*sessions.Session
	ships         	*sqlite3.ShipModel
	templateCache 	map[string]*template.Template
}

// why models?
// 1. database logic isn't tied to our handlers which means that handler responsibilities are limited to HTTP stuff
// 2. by creating the <item>Model type and implementing methods on it we've been able to make our model a single, neatly encapsulated object
// 3. we have total control over which database is used at runtime, just by using the command line flag
// 4. the directory structure scales nicely if your project has multiple back-ends

// Our entry point
//  - parse runtime configuration settings for the application
//  - establish the dependencies for the handlers
//  - run the HTTP server

// -----------------------------------------------------------------------------
// main
func main() {
	port := flag.String("port", ":5033", "HTTPS port on which to listen")
	dsn := flag.String("dsn", "./battleship.db", "SQLite data source name")
	initdb := flag.Bool("initialize", false, "Start with a fresh database")
	secret := flag.String("secret", "nquR81XagSrAEHYXJSFw8y2PLbyWlF1Z", "Secret key")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := initializeDB(*dsn, *initdb)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// initialize new template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize new session manager passing in the secret key
	session := sessions.New([]byte(*secret))
	// sessions expire after 12 hours
	session.Lifetime = 12 * time.Hour

	// new instance of application containing dependencies
	// some are sql.<models> instances
	app := &application{
		battles:       	&sqlite3.BattleModel{DB: db},
		boards:        	&sqlite3.BoardModel{DB: db},
		errorLog:      	errorLog,
		infoLog:       	infoLog,
		players:       	&sqlite3.PlayerModel{DB: db},
		positions:     	&sqlite3.PositionModel{DB: db},
		session:		session,
		ships:         	&sqlite3.ShipModel{DB: db},
		templateCache: 	templateCache,
	}
	// Struct to hold non-default TLS settings
	tlsConfig := &tls.Config {
		PreferServerCipherSuites: 	true,				// ignored if TLS 1.3 is negotiated
		CurvePreferences:			[]tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16 {
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, 
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, 
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, 
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, 
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, 
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:     		*port,
		ErrorLog: 		errorLog,
		Handler:  		app.routes(),
		TLSConfig: 		tlsConfig,			// tslConfig defined above
		IdleTimeout:	time.Minute,		// Keep-alives on accepted connections prevent
		ReadTimeout: 	5 * time.Second,	//  having to repeat the handshake, but 
		WriteTimeout: 	10 * time.Second,	//  idle connections need to be closed
		MaxHeaderBytes:	520192,				// Short ReadTimeout helps prevent against 
											//  slow-client attacks - Slowloris
											// WriteTimeout prevents the data that the 
											//  handler returns from taking too long to write.
											//  It is not meant to prevent long-running handlers
	}

	infoLog.Printf("Starting HTTPS server on %s", *port)
	// Start the HTTPS server, pass in the paths to the TLS cert
	//  and corresponding private key
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
