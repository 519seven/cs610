package main

import (
	"github.com/519seven/cs610/battleship/pkg/models/sqlite3"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

// custom application struct
// this makes objects available to our handlers
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	battles       *sqlite3.BattleModel
	boards        *sqlite3.BoardModel
	positions     *sqlite3.PositionModel
	ships         *sqlite3.ShipModel
	players       *sqlite3.PlayerModel
	templateCache map[string]*template.Template
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
	port := flag.String("port", ":5033", "HTTP port on which to listen")
	dsn := flag.String("dsn", "./battleship.db", "SQLite data source name")
	initdb := flag.Bool("initialize", false, "Start with a fresh database")
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
	// new instance of application containing dependencies
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		battles:       &sqlite3.BattleModel{DB: db},
		boards:        &sqlite3.BoardModel{DB: db},
		players:       &sqlite3.PlayerModel{DB: db},
		positions:     &sqlite3.PositionModel{DB: db},
		ships:         &sqlite3.ShipModel{DB: db},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *port,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *port)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
