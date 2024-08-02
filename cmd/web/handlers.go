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

	tmpl, err := template.ParseFiles("./ui/html/home.html", "./ui/html/nav.html")
	if err != nil {
		log.Print(err)
		return
	}
	allPosts := internal.Read()

	data := struct {
		Posts []internal.Post
	}{
		Posts: allPosts,
	}

	err = tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Print(err)
		return
	}
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := internal.GetPost(id)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Print(err)
		return
	}

	if post.ID == 0 {
		utils.ErrorPage(w, http.StatusNotFound, "Post not found")
		return
	}

	tmpl, err := template.ParseFiles("./ui/html/post.html", "./ui/html/nav.html")
	if err != nil {
		log.Print(err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "post.html", post)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Print(err)
		return
	}
}

func handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("./ui/html/create.html", "./ui/html/nav.html")
		if err != nil {
			log.Print(err)
			return
		}
		err = tmpl.ExecuteTemplate(w, "create.html", nil)
		if err != nil {
			log.Print(err)
			return
		}
	} else if r.Method == http.MethodPost {
		topic := r.FormValue("topic")
		body := r.FormValue("body")
		err := internal.Create(topic, body)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
