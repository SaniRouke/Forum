package main

import (
	"net/http"
)

func (app *Application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", app.Neuter(fileServer)))

	imageServer := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads", app.Neuter(imageServer)))

	mux.HandleFunc("/auth/google/login", app.rateLimiterMW(app.handleGoogleLogin))
	mux.HandleFunc("/auth/google/callback", app.rateLimiterMW(app.handleGoogleCallback))

	mux.HandleFunc("/auth/github/login", app.rateLimiterMW(app.handleGithubLogin))
	mux.HandleFunc("/auth/github/callback", app.rateLimiterMW(app.handleGithubCallback))

	mux.HandleFunc("/", app.rateLimiterMW(app.handlerHome))
	mux.HandleFunc("/post", app.rateLimiterMW(app.handlerPostView))
	mux.HandleFunc("GET /user", app.rateLimiterMW(app.authMW(app.handlerUserPage)))
	mux.HandleFunc("POST /react", app.rateLimiterMW(app.rateLimiterMW(app.handlerReactToPost)))

	mux.HandleFunc("POST /comment", app.rateLimiterMW(app.handlerComment))
	mux.HandleFunc("GET /edit-comment", app.rateLimiterMW(app.handlerEditComment))
	mux.HandleFunc("POST /edit-comment", app.rateLimiterMW(app.handlerEditComment))
	mux.HandleFunc("POST /delete-comment", app.rateLimiterMW(app.handlerDeleteComment))

	mux.HandleFunc("POST /comment-react", app.rateLimiterMW(app.handlerReactToComment))

	mux.HandleFunc("GET /create", app.rateLimiterMW(app.authMW(app.handlerCreatePost)))
	mux.HandleFunc("POST /create", app.rateLimiterMW(app.authMW(app.handlerCreatePost)))

	mux.HandleFunc("GET /edit-post", app.rateLimiterMW(app.handlerEditPost))
	mux.HandleFunc("POST /edit-post", app.rateLimiterMW(app.handlerEditPost))

	mux.HandleFunc("POST /delete-post", app.rateLimiterMW(app.handlerDeletePost))

	mux.HandleFunc("GET /signup", app.rateLimiterMW(app.handlerSignup))
	mux.HandleFunc("POST /signup", app.rateLimiterMW(app.handlerSignup))
	mux.HandleFunc("GET /login", app.rateLimiterMW(app.handlerLogin))
	mux.HandleFunc("POST /login", app.rateLimiterMW(app.handlerLogin))
	mux.HandleFunc("POST /logout", app.rateLimiterMW(app.handlerLogout))

	mux.HandleFunc("GET /notifications", app.rateLimiterMW(app.authMW(app.handlerNotifications)))
	mux.HandleFunc("POST /notifications/mark-as-read", app.rateLimiterMW(app.authMW(app.handlerMarkNotificationAsRead)))

	mux.HandleFunc("GET /admin", app.rateLimiterMW(app.authMW(app.handlerAdminDashboard)))
	mux.HandleFunc("GET /moderator", app.rateLimiterMW(app.authMW(app.handlerModeratorDashboard)))

	mux.HandleFunc("POST /approve-post", app.rateLimiterMW(app.authMW(app.handlerApprovePost)))

	mux.HandleFunc("POST /report-post", app.rateLimiterMW(app.authMW(app.handlerReportPost)))

	mux.HandleFunc("GET /promote", app.rateLimiterMW(app.authMW(app.handlerRequestToModer)))

	mux.HandleFunc("GET /promote-to-moder", app.rateLimiterMW(app.authMW(app.handlerPromoteToModer)))

	mux.HandleFunc("POST /admin/categories/add", app.rateLimiterMW(app.authMW(app.handlerAdminAddCategory)))
	mux.HandleFunc("POST /admin/categories/delete", app.rateLimiterMW(app.authMW(app.handlerAdminDeleteCategory)))

	mux.HandleFunc("POST /admin/promote-user", app.rateLimiterMW(app.authMW(app.handlerAdminPromoteUser)))
	mux.HandleFunc("POST /admin/decline-user", app.rateLimiterMW(app.authMW(app.handlerAdminDeclineUser)))
	mux.HandleFunc("POST /admin/demote-user", app.rateLimiterMW(app.authMW(app.handlerAdminDemoteUser)))

	return mux
}
