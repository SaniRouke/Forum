package main

import (
	"errors"
	"fmt"
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
TODO: 2 pages: likes, my post
TODO: Добавить отображение пользователя
TODO: User page
*/

var ErrorUserExist = errors.New("user already exist")

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {

	user, err := app.GetUserSession(r)
	if err != nil {
		log.Println(err)
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
			User:       user,
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

	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)

}

func (app *Application) handlerReactToPost(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
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
	user, err := app.GetUserSession(r)
	if err != nil {
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
