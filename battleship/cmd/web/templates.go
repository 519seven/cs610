package main

import (
	"519seven/battleship/pkg/models"
	"html/template"
	"path/filepath"
	"time"
)

// Allow more data to be passed to the template
type templateDataBattle struct {
	CurrentYear int
	Battle      *models.Battle
	Battles     []*models.Battle
}
type templateDataBoard struct {
	CurrentYear int
	Board       *models.Board
	Boards      []*models.Board
}
type templateDataPlayer struct {
	CurrentYear int
	Player      *models.Player
	Players     []*models.Player
}
type templateDataPosition struct {
	CurrentYear int
	Position    *models.Position
	Positions   []*models.Position
}

// Give us human-friendly dates
func humanDate(t time.Time) string {
	return t.Format("Jan 02 2006 at 15:04")
}

// initialize template.FuncMap and store it in a global variable
var functions = template.FuncMap{
	"humanDate": humanDate,
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
