package main

import (
	"fmt"
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (app *Application) handlerHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		utils.ErrorPage(w, http.StatusNotFound, "Page not found")
		app.Log.Info("Page not found by (polzovatel dolboeb)")
		return
	}

	if r.Method != "GET" {
		utils.ErrorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// FILTER
	selectedCategories := r.URL.Query()["categories"]

	var allPosts []database.Post
	var err error

	if len(selectedCategories) > 0 {
		allPosts, err = app.Store.Post.GetPostsByCategory(selectedCategories)
	} else {
		allPosts, err = app.Store.Post.GetAll()
	}

	if err != nil {
		app.ServerErr(w, err)
		return
	}

	allCategories, err := app.Store.Post.GetCategories()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Info("Get User Session")
	}

	for i := range allPosts {
		allPosts[i].Category = strings.ReplaceAll(allPosts[i].Category, ",", ", ")
	}

	data := struct {
		Posts      []database.Post
		Categories []string
		User       User
	}{
		Posts:      allPosts,
		Categories: allCategories,
		User:       user,
	}

	err = utils.RenderTemplate(w, "home.html", data, http.StatusOK)
	if err != nil {
		app.ServerErr(w, err)
	}
}

//func (app *Application) handlerShowUserPost(w http.ResponseWriter, r *http.Request) {
//	app.Store.Post.GetPostsByUser()
//}

// TODO: добавить atoi проверку id - валидация
func (app *Application) handlerPostView(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID") //TODO: make constnts
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		log.Println(err)
	}

	post, err := app.Store.Post.GetPost(id, user.ID)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	if post.ID == 0 {
		utils.ErrorPage(w, http.StatusNotFound, "Post not found")
		return
	}

	comments, err := app.Store.Post.GetComments(id, user.ID)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	post.Comments = comments

	post.Category = strings.ReplaceAll(post.Category, ",", ", ")
	data := struct {
		Post database.Post
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

func (app *Application) handlerUserPage(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Add IsAuth field to the user struct for template use
	userData := struct {
		User
		IsAuth bool
	}{
		User:   user,
		IsAuth: true, // Set to true because user is authenticated
	}

	// Initialize variables for data
	var posts []database.Post
	var pageTitle string

	// Determine which section the user is trying to view
	action := r.URL.Query().Get("action")

	switch action {
	case "posts":
		posts, err = app.Store.Post.GetPostsByUser(user.ID)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Failed to retrieve your posts.")
			return
		}
		pageTitle = "My Posts"

	case "comments":
		posts, err = app.Store.Post.GetPostsWithUserComments(user.ID)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Failed to retrieve posts with your comments.")
			return
		}
		pageTitle = "Posts with My Comments"

	case "reactions":
		posts, err = app.Store.Post.GetPostsWithUserReactions(user.ID)
		if err != nil {
			fmt.Println(err)
			utils.ErrorPage(w, http.StatusInternalServerError, "Failed to retrieve posts with your reactions.")
			return
		}
		pageTitle = "My Reactions"

	default:
		pageTitle = "User Profile"
	}

	// Combine user data and other template data into a single struct
	data := struct {
		User      interface{}
		Posts     []database.Post
		PageTitle string
	}{
		User:      userData,
		Posts:     posts,
		PageTitle: pageTitle,
	}

	err = utils.RenderTemplate(w, "user.html", data, http.StatusOK)
	if err != nil {
		log.Println("Error rendering template:", err)
	}
}
