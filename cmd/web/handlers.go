package main

import (
	"forum/cmd/utils"
	"forum/internal"
	"html/template"
	"log"
	"net/http"
)

func handlerHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		utils.ErrorPage(w, http.StatusNotFound, "page not found")
		return
	}

	tmpl, err := template.ParseFiles("./ui/html/home.html")
	if err != nil {
		log.Print(err)
		return
	}
	users := internal.Read()

	err = tmpl.Execute(w, users)
	if err != nil {
		utils.ErrorPage(w, http.StatusNotFound, "page not found")
		log.Print(err)
		return
	}
}

//func handlerPost(w http.ResponseWriter, r http.Request) {
//
//}
