package main

import (
	"html/template"
	"net/http"
)

/*============================================================================*/

type Board struct {
	Title      string
	Opposition string
	Result     string
}

type BoardList struct {
	ListTitle string
	Things    []Board
}

/*============================================================================*/

func displayBoardList(w http.ResponseWriter, r *http.Request) {
	data := BoardList{
		ListTitle: "Your Campaigns, Captain!",
		Things: []Board{
			{Title: "East India", Opposition: "John", Result: "Win"},
			{Title: "South Sea", Opposition: "Brett", Result: "Win"},
			{Title: "Gibraltar", Opposition: "Pete", Result: "Win"},
		},
	}
	tmpl := template.Must(template.ParseFiles("listings.html"))
	tmpl.Execute(w, data)
}
