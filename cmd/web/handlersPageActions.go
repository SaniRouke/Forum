package main

import (
	"errors"
	"fmt"
	"forum/cmd/utils"
	"forum/internal/database"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var ErrorUserExist = errors.New("user already exist")

func (app *Application) handlerEditPost(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Error(err.Error())
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		app.ErrorPage(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		app.ErrorPage(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	switch {
	case r.Method == http.MethodGet:

		categoriesFromDB, err := app.Store.Post.GetCategories()
		if err != nil {
			app.ServerErr(w, err)
			return
		}

		post, err := app.Store.Post.GetPost(id, user.ID)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			app.Log.Error(err.Error())
			return
		}

		if post.ID == 0 {
			app.ErrorPage(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		type CategoryHTML struct {
			Name      string
			IsChecked bool
		}

		categories := func(list string) []CategoryHTML {
			categorySlice := strings.Split(list, ",")
			result := []CategoryHTML{}

			for _, name := range categoriesFromDB {
				result = append(result, CategoryHTML{
					Name:      name,
					IsChecked: slices.Contains(categorySlice, name),
				})
			}

			return result
		}(post.Category)

		type Data struct {
			User       User
			Post       database.Post
			Categories []CategoryHTML
		}

		data := &Data{
			User:       user,
			Categories: categories,
			Post:       post,
		}

		err = utils.RenderTemplate(w, "edit-post.html", data, http.StatusOK)
		if err != nil {
			app.Log.Error(err.Error())
		}

	case r.Method == http.MethodPost:
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			app.ErrorPage(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		topic := r.FormValue("topic")
		body := r.FormValue("body")

		if !utils.IsValidInput(topic) || !utils.IsValidInput(body) {
			app.ErrorPage(w, http.StatusBadRequest, "Write A Normal Post, Bro")
			return
		}

		category := strings.Join(r.PostForm["categories"], ",")

		validCategories, err := app.Store.Post.GetCategories()
		if err != nil {
			app.ServerErr(w, err)
			return
		}

		isValidCategory := true
		for _, c := range r.PostForm["categories"] {
			if !slices.Contains(validCategories, c) {
				isValidCategory = false
				break
			}
		}

		if !isValidCategory {
			app.ErrorPage(w, http.StatusBadRequest, "Don't Play With Us, Bro")
			fmt.Println("validCategories:", validCategories)
			fmt.Println("category:", category)
			return
		}

		imagePath, err := app.handleImageUpload(r, user)
		if err != nil {
			app.ErrorPage(w, http.StatusNotFound, err.Error())
			return
		}

		postForm := database.CreatePostForm{
			Topic:     topic,
			Body:      body,
			Category:  category,
			UserID:    user.ID,
			ImagePath: imagePath,
		}

		err = app.Store.Post.EditPost(idInt, postForm)

		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			app.Log.Error(err.Error())
			return
		}

		http.Redirect(w, r, "/post?id="+id, http.StatusSeeOther)
	}
}

func (app *Application) handlerRequestToModer(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if user.Role != "user" {
		app.ErrorPage(w, http.StatusForbidden, "You are already in power")
		return
	}

	err = app.Store.User.PromoteMe(user.ID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, "/user", http.StatusSeeOther)
}

func (app *Application) handlerPromoteToModer(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if user.Role != "admin" {
		app.ErrorPage(w, http.StatusForbidden, "You are not an admin")
		return
	}
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err = app.Store.User.PromoteToModer(userID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func (app *Application) handlerApprovePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.ErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if user.Role != "admin" && user.Role != "moderator" {
		app.ErrorPage(w, http.StatusForbidden, "You do not have permissions to access this page.")
		return
	}

	err = app.Store.Post.ApprovePost(postID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) handlerReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.ErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if user.Role != "moderator" {
		app.ErrorPage(w, http.StatusForbidden, "You do not have reporting rights.")
		return
	}

	err = app.Store.Post.ReportPost(user.ID, postID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}

func (app *Application) handlerCreatePost(w http.ResponseWriter, r *http.Request) {

	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Error(err.Error())
	}

	switch {
	case r.Method == http.MethodGet:

		categoriesFromDB, err := app.Store.Post.GetCategories()
		if err != nil {
			app.ServerErr(w, err)
			return
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

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			app.ErrorPage(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		topic := r.FormValue("topic")
		body := r.FormValue("body")

		if !utils.IsValidInput(topic) || !utils.IsValidInput(body) {
			app.ErrorPage(w, http.StatusBadRequest, "Write A Normal Post, Bro")
			return
		}

		category := strings.Join(r.PostForm["categories"], ",")

		validCategories, err := app.Store.Post.GetCategories()
		if err != nil {
			app.ServerErr(w, err)
			return
		}

		isValidCategory := true
		for _, c := range r.PostForm["categories"] {
			if !slices.Contains(validCategories, c) {
				isValidCategory = false
				break
			}
		}

		if !isValidCategory {
			app.ErrorPage(w, http.StatusBadRequest, "Don't Play With Us, Bro")
			return
		}

		imagePath, err := app.handleImageUpload(r, user)
		if err != nil {
			app.ErrorPage(w, http.StatusNotFound, err.Error())
			return
		}

		postForm := database.CreatePostForm{
			Topic:     topic,
			Body:      body,
			Category:  category,
			UserID:    user.ID,
			ImagePath: imagePath,
		}

		err = app.Store.Post.CreatePost(postForm)

		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			app.Log.Error(err.Error())
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *Application) handlerDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.ErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postIDStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := app.Store.Post.GetPost(postIDStr, user.ID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	if post.UserID != user.ID {
		app.ErrorPage(w, http.StatusForbidden, "You are not allowed to delete this post")
		return
	}

	err = app.Store.Post.DeletePost(postID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) handlerDeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.ErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	commentIDStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	comment, err := app.Store.Post.GetCommentByID(commentID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	if comment.UserID != user.ID {
		app.ErrorPage(w, http.StatusForbidden, "You are not allowed to delete this comment")
		return
	}

	err = app.Store.Post.DeleteComment(commentID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/post?id=%d", comment.PostID), http.StatusSeeOther)
}

func (app *Application) handleImageUpload(r *http.Request, user User) (string, error) {
	file, header, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", fmt.Errorf("error retrieving the file: %v", err)
	}
	defer file.Close()

	if header.Size > 20<<20 {
		return "", fmt.Errorf("file size exceeds the 20 MB limit")
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file header: %v", err)
	}
	fileType := http.DetectContentType(buffer)
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/gif" {
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %v", err)
	}

	fileExtension := filepath.Ext(header.Filename)
	if fileExtension == "" {
		return "", fmt.Errorf("file must have a valid extension")
	}

	fileName := fmt.Sprintf("%d_userID_%d%s", time.Now().UnixNano(), user.ID, fileExtension)
	filePath := filepath.Join("uploads", fileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return "/" + filePath, nil
}

func (app *Application) handlerEditComment(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Error(err.Error())
		app.ErrorPage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		app.ErrorPage(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Fetch the comment by ID
		comment, err := app.Store.Post.GetCommentByID(idInt)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, "Failed to fetch comment")
			app.Log.Error(err.Error())
			return
		}

		if comment.ID == 0 {
			app.ErrorPage(w, http.StatusNotFound, "Comment not found")
			return
		}

		// Check if the user is the author of the comment
		if comment.UserID != user.ID {
			app.ErrorPage(w, http.StatusForbidden, "You are not allowed to edit this comment")
			return
		}

		// Fetch the post associated with the comment (if needed)
		post, err := app.Store.Post.GetPost(id, user.ID)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, "Failed to fetch post")
			app.Log.Error(err.Error())
			return
		}

		data := struct {
			Post    database.Post
			Comment database.Comment
			User    User
		}{
			Post:    post,
			Comment: comment,
			User:    user,
		}

		err = utils.RenderTemplate(w, "edit-comment.html", data, http.StatusOK)
		if err != nil {
			app.Log.Error(err.Error())
		}

	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			app.ErrorPage(w, http.StatusBadRequest, "Invalid form data")
			return
		}
		body := r.FormValue("body")
		if body == "" {
			app.ErrorPage(w, http.StatusBadRequest, "Comment body cannot be empty")
		}

		comment, err := app.Store.Post.GetCommentByID(idInt)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, "Failed to fetch comment")
			app.Log.Error(err.Error())
			return
		}

		if comment.ID == 0 {
			app.ErrorPage(w, http.StatusNotFound, "Comment not found")
			return
		}

		// Check if the user is the author of the comment
		if comment.UserID != user.ID {
			app.ErrorPage(w, http.StatusForbidden, "You are not allowed to edit this comment")
			return
		}

		comment.Body = body

		err = app.Store.Post.EditComment(comment)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, "Failed to update comment")
			app.Log.Error(err.Error())
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/post?id=%d", comment.PostID), http.StatusSeeOther)
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

	if !utils.IsValidInput(commentBody) {
		app.ErrorPage(w, http.StatusBadRequest, "Write A Normal Comment, Bro")
		return
	}

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

	post, err := app.Store.Post.GetPost(postID, user.ID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	if post.UserID != user.ID {
		notification := database.Notification{
			UserID:          post.UserID,
			InitiatorUserID: user.ID,
			PostID:          id,
			CommentID:       nil,
			Type:            "comment",
			Date:            time.Now().Format("2006-01-02 15:04:05"),
		}
		err = app.Store.Notification.CreateNotification(notification)
		if err != nil {
			app.Log.Error("Error creating notification:", err)
		}
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
		return
	}

	postExists, err := app.Store.Post.DoesPostExist(intPostID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}
	if !postExists {
		app.ErrorPage(w, http.StatusForbidden, "Do Not Play With Us, Bro")
		return
	}

	currentReaction, err := app.Store.Post.CheckPostReaction(intPostID, userID)
	if err != nil {
		app.Log.Error(err.Error())
	}

	switch {
	case currentReaction == 0:
		err = app.Store.Post.SetPostReaction(intPostID, userID, reactionToDB)
	case currentReaction == reactionToDB:
		err = app.Store.Post.DeletePostReaction(intPostID, userID)
	default:
		err = app.Store.Post.UpdatePostReaction(intPostID, userID, reactionToDB)
	}

	if err != nil {
		app.Log.Error("error updating reaction:", err)
	}
	post, err := app.Store.Post.GetPost(postID, userID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}
	if post.UserID != userID {
		notification := database.Notification{
			UserID:          post.UserID,
			InitiatorUserID: userID,
			PostID:          intPostID,
			CommentID:       nil,
			Type:            "like", // or "dislike" based on the reaction
			Date:            time.Now().Format("2006-01-02 15:04:05"),
		}
		err = app.Store.Notification.CreateNotification(notification)
		if err != nil {
			app.Log.Error("Error creating notification:", err)
		}
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
	commentExists, err := app.Store.Post.DoesCommentExist(intCommentID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}
	if !commentExists {
		app.ErrorPage(w, http.StatusForbidden, "Do Not Play With Us, Bro")
		return
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
