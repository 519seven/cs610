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

// Custom application struct
// - Makes objects available to our handlers
type application struct {
	errorLog      	*log.Logger
	infoLog       	*log.Logger

	battles       	*sqlite3.BattleModel
	boards        	*sqlite3.BoardModel
	players       	*sqlite3.PlayerModel
	positions     	*sqlite3.PositionModel
	ships         	*sqlite3.ShipModel

	session			*sessions.Session
	templateCache 	map[string]*template.Template
}

// Why models?
// 1. Database logic isn't tied to our handlers which means that
//    handler responsibilities are limited to HTTP stuff
// 2. By creating the <item>Model type and implementing methods on it we've 
//    been able to make our model a single, neatly encapsulated object
// 3. We have total control over which database is used at runtime, just by 
//    using the command line flag
// 4. The directory structure scales nicely if your project has multiple back-ends

// Our entry point
//  - Parse runtime configuration settings for the application
//  - Establish the dependencies for the handlers
//  - Run the HTTP server

// -----------------------------------------------------------------------------
// main
func main() {
	port := flag.String("port", ":5033", "HTTPS port on which to listen")
	dsn := flag.String("dsn", "./battleship.db", "SQLite data source name")
	initdb := flag.Bool("initialize", false, "Start with a fresh database")
	// 32 bytes long
	secret := flag.String("secret", "nquR81XagSrAEHYXJSFw8y2PLbyWlF1Z", "Secret key")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := initializeDB(*dsn, *initdb)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Initialize new template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Sessions
	// - Initialize new session manager passing in the secret key
	// - Sessions expire after 12 hours
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	// New instance of application containing dependencies
	// - Some are sql.<models> instances
	app := &application{
		errorLog:      	errorLog,
		infoLog:       	infoLog,

		battles:       	&sqlite3.BattleModel{DB: db},
		boards:        	&sqlite3.BoardModel{DB: db},
		players:       	&sqlite3.PlayerModel{DB: db},
		positions:     	&sqlite3.PositionModel{DB: db},
		ships:         	&sqlite3.ShipModel{DB: db},

		session:		session,
		templateCache: 	templateCache,
	}

	// Struct to hold non-default TLS settings
	tlsConfig := &tls.Config {
		PreferServerCipherSuites: 	true,	// this serves many purposes ---> 	// a.) ignored if TLS 1.3 is negotiated
		CurvePreferences:			[]tls.CurveID{tls.X25519, tls.CurveP256},	// b.) prefer the cipher suites that are first in the slice
		CipherSuites: []uint16 {												// c.) also meant to prioritize what is best for my server's hardware
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
