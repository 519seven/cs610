package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/519seven/cs610/battleship/pkg/forms"
	"github.com/519seven/cs610/battleship/pkg/models"
)

// Allow more data to be passed to the template
type templateDataBattle struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Battle      			*models.Battle
	Battles     			[]*models.Battle
}
// single board
type templateDataBoard struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	PositionsOnBoard       	*models.PositionsOnBoard
	PositionsOnBoards      	[]*models.PositionsOnBoard
}
// list of boards
type templateDataBoards struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Board					*models.Board
	Boards					[]*models.Board
}
type templateDataLogin struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Login      				*models.Login
}
type templateDataPlayer struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Player      			*models.Player
	Players     			[]*models.Player
}
type templateDataPlayers struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Player      			*models.Player
	Players     			[]*models.Player
}
type templateDataSignup struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Signup      			*models.Signup
}
type templateDataPosition struct {
	CurrentYear 			int
	Flash					string
	Form					*forms.Form
	IsAuthenticated			bool
	Position    			*models.Position
	Positions   			[]*models.Position
}

// Give us human-friendly dates
func humanDate(t time.Time) string {
	return t.Format("Jan 02 2006 at 15:04")
}

// Give us an iterate function to use in templates
func iterateRows(count uint, start uint) []uint {
	var i uint
	var items []uint
	for i = start; i < count+1; i++ {
		items = append(items, i)
	}
	return items
}

// Iterate over letters
func iterateColumns(count uint) []string {
	var i uint
	var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var items []string
	for i = 0; i < (count); i++ {
		items = append(items, string(alphabet[i]))
	}
	return items
}

// Initialize template.FuncMap and store it in a GLOBAL variable
// - FuncMap is essentially a string-keyed map
//   - acts as a go-between for custom template functions and the functions themselves
var functions = template.FuncMap{
	"humanDate": humanDate,
	"iterateColumns": iterateColumns,
	"iterateRows": iterateRows,
}

// Cache templates
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// the template.FuncMap must be registered with the template set
		// use .New() to create an empty template set, use Funcs to register
		// the FuncMap, and parse the files as normal
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
