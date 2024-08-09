package main

import (
	"forum/cmd/utils"
	"forum/internal"
	"log"
	"net/http"
)

func handlerHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		utils.ErrorPage(w, http.StatusNotFound, "Page not found")
		return
	}

	if r.Method != "GET" {
		utils.ErrorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	allPosts, err := internal.GetAllPosts()
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal server error")
		log.Println(err)
		return
	}

	data := struct {
		Posts []internal.Post
	}{
		Posts: allPosts,
	}

	err = utils.RenderTemplate(w, "home.html", data, http.StatusOK)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return
	}

	if post.ID == 0 {
		utils.ErrorPage(w, http.StatusNotFound, "Post not found")
		return
	}

	err = utils.RenderTemplate(w, "post.html", post, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		err := utils.RenderTemplate(w, "create.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
	case r.Method == http.MethodPost:
		topic := r.FormValue("topic")
		body := r.FormValue("body")
		err := internal.CreatePost(topic, body)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handlerDeletePost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err := internal.DeletePost(id)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handlerEditPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		post, err := internal.GetPost(id)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		if post.ID == 0 {
			utils.ErrorPage(w, http.StatusNotFound, "Post not found")
			return
		}

		err = utils.RenderTemplate(w, "edit.html", post, http.StatusOK)
		if err != nil {
			log.Println(err)
		}

	case http.MethodPost:
		topic := r.FormValue("topic")
		body := r.FormValue("body")

		err := internal.EditPost(id, topic, body)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/post?id="+id, http.StatusSeeOther)
	}
}

func handlerSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "signup.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := internal.CreateUser(username, email, password)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Unable to create user")
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "login.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		identifier := r.FormValue("username") // Can be username or email
		password := r.FormValue("password")

		isAuthenticated, err := internal.AuthenticateUser(identifier, password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Println("Error in authentication:", err)
			return
		}

		if !isAuthenticated {
			utils.ErrorPage(w, http.StatusBadRequest, "Invalid username or password")
			log.Println("Failed login attempt for:", identifier)
			return
		}

		// Redirect to home page on successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
