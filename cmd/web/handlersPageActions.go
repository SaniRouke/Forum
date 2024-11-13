package main

import (
	"errors"
	"fmt"
	"forum/cmd/utils"
	"forum/internal/database"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrorUserExist = errors.New("user already exist")

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {

	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Error(err.Error())
	}

	switch {
	case r.Method == http.MethodGet:

		categoriesFromDB, err := app.Store.Post.GetCategories()
		if err != nil {
			app.Log.Error(err.Error())
		}
		data := struct {
			User       User
			Categories []string
		}{
			User:       user,
			Categories: categoriesFromDB,
		}

		err = utils.RenderTemplate(w, "create.html", data, http.StatusOK)
		if err != nil {
			app.Log.Error(err.Error())
		}

	case r.Method == http.MethodPost:

		topic := r.FormValue("topic")
		body := r.FormValue("body")

		if !utils.IsValidInput(topic) || !utils.IsValidInput(body) {
			app.ErrorPage(w, http.StatusBadRequest, "Write A Normal Post, Bro")
			return
		}

		category := strings.Join(r.PostForm["categories"], ",")
		postForm := database.CreatePostForm{
			topic, body, category, user.ID,
		}

		err := app.Store.Post.CreatePost(postForm)
		fmt.Println(postForm)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			app.Log.Error(err.Error())
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
	commentBody := r.FormValue("comment_body")
	date := time.Now().Format("2006-01-02 15:04:05")

	id, err := strconv.Atoi(postID)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	err = app.Store.Post.AddComment(id, user.ID, commentBody, date)
	if err != nil {
		app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		app.Log.Error(err.Error())
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
		app.Log.Error(err.Error())
	}

	currentReaction, err := app.Store.Post.CheckPostReaction(intPostID, userID)
	if err != nil {
		app.Log.Error(err.Error())
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
		app.Log.Error("error updating reaction:", err)
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
		app.Log.Error(err.Error())
	}

	currentReaction, err := app.Store.Post.CheckCommentReaction(intCommentID, userID)
	if err != nil {
		app.Log.Error(err.Error())
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
		app.Log.Error("Error updating reaction:", err)
	}

	http.Redirect(w, r, "/post?id="+string(postID), http.StatusSeeOther)

}
