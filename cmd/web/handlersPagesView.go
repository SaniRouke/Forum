package main

import (
	"database/sql"
	"forum/cmd/utils"
	"forum/internal/database"
	"net/http"
	"strconv"
	"strings"
)

func (app *Application) handlerHome(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		app.ErrorPage(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	if r.Method != "GET" {
		app.ErrorPage(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

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
		app.Log.Info("unregistered user action")
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

func (app *Application) handlerPostView(w http.ResponseWriter, r *http.Request) {

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

	user, err := app.GetUserSession(r)
	if err != nil {
		app.Log.Error(err.Error())
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

	// Загрузка комментариев
	comments, err := app.Store.Post.GetComments(id, user.ID)
	if err != nil {
		app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		app.Log.Error(err.Error())
		return
	}
	post.Comments = comments

	// Если текущий пользователь - модератор, проверяем, репортил ли он этот пост
	if user.Role == "moderator" {
		reported, err := app.Store.Post.IsReportedByModerator(user.ID, idInt)
		if err == nil && reported {
			post.ReportedByCurrentModerator = true
		}
	}

	// Приведение категорий к более читабельному виду
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
		app.Log.Error(err.Error())
	}
}

func (app *Application) handlerUserPage(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userData := struct {
		User
		IsAuth bool
	}{
		User:   user,
		IsAuth: true,
	}

	var posts []database.Post
	var pageTitle string

	action := r.URL.Query().Get("action")

	hasRequest, err := app.Store.User.HasActiveModeratorRequest(user.ID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	switch action {
	case "posts":
		posts, err = app.Store.Post.GetPostsByUser(user.ID)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		pageTitle = "My Posts"

	case "comments":
		posts, err = app.Store.Post.GetPostsWithUserComments(user.ID)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		pageTitle = "Posts With My Comments"

	case "reactions":
		posts, err = app.Store.Post.GetPostsWithUserReactions(user.ID)
		if err != nil {
			app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			app.Log.Error(err.Error())
			return
		}
		pageTitle = "My Reactions"

	default:
		pageTitle = "User Profile"
	}

	data := struct {
		User                interface{}
		Posts               []database.Post
		PageTitle           string
		HasModeratorRequest bool
	}{
		User:                userData,
		Posts:               posts,
		PageTitle:           pageTitle,
		HasModeratorRequest: hasRequest,
	}

	err = utils.RenderTemplate(w, "user.html", data, http.StatusOK)
	if err != nil {
		app.Log.Error("error rendering template:", err)
	}
}

func (app *Application) handlerNotifications(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	notifications, err := app.Store.Notification.GetNotificationsByUserID(user.ID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	promotionRequests := []database.UsersForPromotion{}

	if user.Role == "admin" {
		promotionRequests, err = app.Store.User.GetPromotionRequests()
	}

	data := struct {
		User          User
		Notifications []database.Notification
		Promotions    []database.UsersForPromotion
	}{
		User:          user,
		Notifications: notifications,
		Promotions:    promotionRequests,
	}

	err = utils.RenderTemplate(w, "notifications.html", data, http.StatusOK)
	if err != nil {
		app.Log.Error("error rendering template:", err)
	}
}

func (app *Application) handlerMarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		app.ErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	notificationIDStr := r.FormValue("notification_id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		app.ErrorPage(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Retrieve the notification from the database
	notification, err := app.Store.Notification.GetNotificationByID(notificationID)
	if err != nil {
		if err == sql.ErrNoRows {
			app.ErrorPage(w, http.StatusNotFound, "Notification not found")
		} else {
			app.ServerErr(w, err)
		}
		return
	}

	// Check if the notification belongs to the current user
	if notification.UserID != user.ID {
		app.ErrorPage(w, http.StatusForbidden, "You are not authorized to perform this action")
		return
	}

	// Mark the notification as read
	err = app.Store.Notification.MarkNotificationAsRead(notificationID)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func (app *Application) handlerModeratorDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil || (user.Role != "moderator" && user.Role != "admin") {
		app.ErrorPage(w, http.StatusForbidden, "You are not allowed here.")
		return
	}

	pendingPosts, err := app.Store.Post.GetAllPending()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	data := struct {
		User         User
		PendingPosts []database.Post
	}{
		User:         user,
		PendingPosts: pendingPosts,
	}

	err = utils.RenderTemplate(w, "moderator-dashboard.html", data, http.StatusOK)
	if err != nil {
		app.Log.Error(err.Error())
	}
}

func (app *Application) handlerAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := app.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != "admin" {
		app.ErrorPage(w, http.StatusForbidden, "Forbidden")
		return
	}

	categories, err := app.Store.Post.GetCategories()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	reports, err := app.Store.Notification.GetAllReports()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	promotionRequests, err := app.Store.User.GetPromotionRequests()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	// Получаем список действующих модераторов
	moderators, err := app.Store.User.GetAllModerators()
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	data := struct {
		User              User
		Categories        []string
		Reports           []database.Report
		PromotionRequests []database.UsersForPromotion
		Moderators        []database.UsersForPromotion
	}{
		User:              user,
		Categories:        categories,
		Reports:           reports,
		PromotionRequests: promotionRequests,
		Moderators:        moderators,
	}

	err = utils.RenderTemplate(w, "admin-dashboard.html", data, http.StatusOK)
	if err != nil {
		app.Log.Error("error rendering template:", err)
	}
}
