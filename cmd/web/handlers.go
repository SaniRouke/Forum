package main

import (
	"fmt"
	"forum/cmd/utils"
	"forum/internal"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (app *Application) handlerHome(w http.ResponseWriter, r *http.Request) {

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

	userNameCookie, err := r.Cookie("user_name")
	if err != nil {
		log.Println(err)
	}

	var user User

	if userNameCookie != nil {
		user.Name = userNameCookie.Value
		user.IsAuth = true
	}

	data := struct {
		Posts []internal.Post
		User  User
	}{
		Posts: allPosts,
		User:  user,
	}

	err = utils.RenderTemplate(w, "home.html", data, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func (app *Application) handlerPostView(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := internal.GetPost(id)
	if err != nil {
		fmt.Println("add1")
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	if post.ID == 0 {
		utils.ErrorPage(w, http.StatusNotFound, "Post not found")
		return
	}

	comments, err := internal.GetComments(id)
	if err != nil {
		fmt.Println("add2")
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	post.Comments = comments

	userNameCookie, err := r.Cookie("user_name")
	if err != nil {
		log.Println(err)
	}

	var user User

	if userNameCookie != nil {
		user.Name = userNameCookie.Value
		user.IsAuth = true
	}

	data := struct {
		Post internal.Post
		User User
	}{
		Post: post,
		User: user,
	}

	err = utils.RenderTemplate(w, "post.html", data, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {

	switch {
	case r.Method == http.MethodGet:

		userCookie, err := r.Cookie("user_name")
		if err != nil {
			log.Println(err)
		}

		if userCookie == nil {
			http.Redirect(w, r, "login", http.StatusSeeOther)
			return
		}
		if userCookie.Value == "" {
			http.Redirect(w, r, "login", http.StatusSeeOther)
			return
		}

		userNameCookie, err := r.Cookie("user_name")
		if err != nil {
			log.Println(err)
		}

		var user User

		if userNameCookie != nil {
			user.Name = userNameCookie.Value
			user.IsAuth = true
		}

		data := struct {
			//Post internal.Post
			User User
		}{
			//Post: post,
			User: user,
		}

		err = utils.RenderTemplate(w, "create.html", data, http.StatusOK)
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

func (app *Application) handlerComment(w http.ResponseWriter, r *http.Request) {

	userCookie, err := r.Cookie("user_name")
	if err != nil || userCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := internal.GetUser(userCookie.Value)
	if err != nil || user.ID == 0 {
		http.Error(w, "Unauthorized: Invalid user", http.StatusUnauthorized)
		return
	}
	postID := r.FormValue("post_id")
	commentBody := r.FormValue("comment_body") // TODO: make constant
	date := time.Now().Format("2006-01-02 15:04:05")

	id, err := strconv.Atoi(postID)
	if err != nil {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err = internal.AddComment(id, user.ID, commentBody, date)
	if err != nil {

		utils.ErrorPage(w, http.StatusInternalServerError, "Unable to add comment")
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)

}

func (app *Application) handlerSignup(w http.ResponseWriter, r *http.Request) {
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

func (app *Application) handlerLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		data := struct {
			User User
		}{
			User: app.User,
		}

		err := utils.RenderTemplate(w, "login.html", data, http.StatusOK)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")

		user, err := internal.GetUser(username)
		if err != nil {
			log.Println(err)
			return
		}

		cookie := &http.Cookie{
			Name:     "user_name",
			Value:    user.Username,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60 * 60 * 24,
		}

		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *Application) handlerLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "user_name",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
