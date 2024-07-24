package main

import (
	"html/template"
	"log"
	"net/http"
)

func handlerHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("./ui/html/home.html")
	if err != nil {
		log.Print(err)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		// 500
		return
	}
}
