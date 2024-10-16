package main

import (
	"context"
	"fmt"
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//func (app *Application) handlerHomeTest(w http.ResponseWriter, r *http.Request) {
//
//	fmt.Fprintln(w, "Home handler activity")
//}
//
//func (app *Application) middlewareTest(w http.ResponseWriter, r *http.Request) func(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintln(w, "GET COOKIE IN MIDDLEWARE")
//	func
//}

func (app *Application) authMW(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("user_name")
		if err != nil || userCookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := app.Store.User.GetUser(userCookie.Value)
		if err != nil || user.ID == 0 {
			// If user is not found, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Attach the user data to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", user)
		r = r.WithContext(ctx)

		// Call the next handler with the updated request
		next(w, r)
	}
}

func (app *Application) handlerHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		utils.ErrorPage(w, http.StatusNotFound, "Page not found")
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
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal server error")
		log.Println(err)
		return
	}

	allCategories, err := app.Store.Post.GetCategories()
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal server error")
		log.Println(err)
		return
	}

	userNameCookie, err := r.Cookie("user_name")

	var user User
	if userNameCookie != nil {
		user.Name = userNameCookie.Value
		user.IsAuth = true
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
		log.Println(err)
	}
}

func (app *Application) handlerPostView(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userCookie, err := r.Cookie("user_name")
	if err != nil || userCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println(err)
		return
	}

	user, err := app.Store.User.GetUser(userCookie.Value)
	if err != nil || user.ID == 0 {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
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
		User: User{
			ID:     user.ID,
			Name:   user.Username,
			IsAuth: true,
		},
	}

	err = utils.RenderTemplate(w, "post.html", data, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func (app *Application) handlerUserPage(w http.ResponseWriter, r *http.Request) {
	userCookie, err := r.Cookie("user_name")
	if err != nil || userCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve user info from the database
	user, err := app.Store.User.GetUser(userCookie.Value)
	if err != nil || user.ID == 0 {
		utils.ErrorPage(w, http.StatusInternalServerError, "Failed to retrieve user information.")
		return
	}

	// Add IsAuth field to the user struct for template use
	userData := struct {
		database.User
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

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	var userForTemplate User
	user, ok := r.Context().Value("user").(database.User)
	if !ok {
		log.Println("Юзер-хуюзер не найден")
	} else {
		userForTemplate.Name = user.Username
		userForTemplate.IsAuth = true
	}
	switch {
	case r.Method == http.MethodGet:

		categoriesFromDB, err := app.Store.Post.GetCategories()
		if err != nil {
			log.Println(err)
		}
		data := struct {
			//Post internal.Post
			User       User
			Categories []string
		}{
			//Post: post,
			User:       userForTemplate,
			Categories: categoriesFromDB,
		}

		err = utils.RenderTemplate(w, "create.html", data, http.StatusOK)
		if err != nil {
			log.Println(err)
		}

	case r.Method == http.MethodPost:

		topic := r.FormValue("topic")
		body := r.FormValue("body")

		if !utils.IsValidInput(topic) || !utils.IsValidInput(body) {
			utils.ErrorPage(w, http.StatusBadRequest, "Write a normal post, bro.")
			return
		}
		//if len(r.PostForm["categories"]) == 0 {
		//	utils.ErrorPage(w, http.StatusBadRequest, "Please choose at least one category.")
		//	return
		//}
		category := strings.Join(r.PostForm["categories"], ",")
		postForm := database.CreatePostForm{
			topic, body, category, user.ID,
		}

		err := app.Store.Post.CreatePost(postForm)
		fmt.Println(postForm)
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
	user, err := app.Store.User.GetUser(userCookie.Value)
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

	err = app.Store.Post.AddComment(id, user.ID, commentBody, date)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Unable to add comment")
		log.Println(err)
		return
	}
	fmt.Println(id, user.Username, commentBody, date)
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

	var asciiRegex = regexp.MustCompile(`^[!-}]+$`)

	if r.Method == http.MethodPost {

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		dateOfCreation := time.Now().Format("2006-01-02 15:04:05")

		if !asciiRegex.MatchString(username) || !asciiRegex.MatchString(password) {
			utils.ErrorPage(w, http.StatusBadRequest, "My fellow skuf, your username and password can only contain ASCII characters between 33 and 125. If you're unfamiliar with the ASCII table, now is the time to check it out.")
			log.Println("Invalid username or password format.")
			return
		}

		if !utils.IsValidPassword(password) {
			utils.ErrorPage(w, http.StatusBadRequest, "My fellow skuf, your password must be at least 8 characters long and consist of letters and numbers.")
			return
		}

		err := app.Store.User.CreateUser(username, email, password, dateOfCreation)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "My fellow skuf, you are trying to use an existing email or username")
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
		password := r.FormValue("password")

		isAuthenticated, err := app.Store.User.AuthenticateUser(username, password)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		if !isAuthenticated {
			utils.ErrorPage(w, http.StatusUnauthorized, "Invalid username or password. Have you forgotten your login information?")
			return
		}

		user, err := app.Store.User.GetUser(username)
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

func (app *Application) handlerReactToPost(w http.ResponseWriter, r *http.Request) {
	userCookie, err := r.Cookie("user_name")
	if err != nil || userCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := app.Store.User.GetUser(userCookie.Value)
	if err != nil {
		log.Println("Invalid user ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID := user.ID
	postID := r.FormValue("post_id")
	reaction := r.FormValue("reaction")
	var reactionToDB int

	if reaction == "like" {
		reactionToDB = 1
	} else {
		reactionToDB = -1
	}
	intPostID, err := strconv.Atoi(postID)
	if err != nil {
		log.Println(err)
	}

	currentReaction, err := app.Store.Post.CheckPostReaction(intPostID, userID)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(currentReaction)

	switch {
	case currentReaction == 0:
		err = app.Store.Post.SetPostReaction(intPostID, userID, reactionToDB)
	case currentReaction == reactionToDB:
		app.Store.Post.DeletePostReaction(intPostID, userID)
	default:
		err = app.Store.Post.UpdatePostReaction(intPostID, userID, reactionToDB)
	}

	if err != nil {
		log.Println("Error updating reaction:", err)
	}

	http.Redirect(w, r, "/post?id="+string(postID), http.StatusSeeOther)

}

func (app *Application) handlerReactToComment(w http.ResponseWriter, r *http.Request) {
	userCookie, err := r.Cookie("user_name")
	if err != nil || userCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := app.Store.User.GetUser(userCookie.Value)
	if err != nil {
		log.Println("Invalid user ID")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID := user.ID
	postID := r.FormValue("post_id")
	commentID := r.FormValue("comment_id")
	reaction := r.FormValue("reaction")
	var reactionToDB int

	if reaction == "like" {
		reactionToDB = 1
	} else {
		reactionToDB = -1
	}
	intCommentID, err := strconv.Atoi(commentID)
	if err != nil {
		log.Println(err)
	}

	currentReaction, err := app.Store.Post.CheckCommentReaction(intCommentID, userID)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(currentReaction)

	switch {
	case currentReaction == 0:
		err = app.Store.Post.SetCommentReaction(intCommentID, userID, reactionToDB)
	case currentReaction == reactionToDB:
		err = app.Store.Post.DeleteCommentReaction(intCommentID, userID)
	default:
		err = app.Store.Post.UpdateCommentReaction(intCommentID, userID, reactionToDB)
	}

	if err != nil {
		log.Println("Error updating reaction:", err)
	}

	http.Redirect(w, r, "/post?id="+string(postID), http.StatusSeeOther)

}
