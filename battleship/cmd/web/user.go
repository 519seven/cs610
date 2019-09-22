package main

import (
	"html/template"
	"net/http"
)

/*============================================================================*/

type User struct {
	Title       string
	Captain     string
	TimesPlayed int
}

type UserList struct {
	ListTitle string
	Things    []User
}

/*============================================================================*/

func displayUserList(w http.ResponseWriter, r *http.Request) {
	data := UserList{
		ListTitle: "Opposing Leaders",
		Things: []User{
			{Title: "Captain Naval Lint", Captain: "John", TimesPlayed: 2},
			{Title: "Admiral Wet Socks", Captain: "Brett", TimesPlayed: 3},
			{Title: "Admiral Sassyfrass", Captain: "Pete", TimesPlayed: 22},
		},
	}
	tmpl := template.Must(template.ParseFiles("listings.html"))
	tmpl.Execute(w, data)
}
