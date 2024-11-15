package main

import (
	"fmt"
	"forum/cmd/utils"
	"net/http"
	"strings"
)

func (app *Application) authMW(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := app.GetUserSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		exist, err := app.Store.User.CheckToken(user.Token)
		if !exist {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

func (app *Application) ErrorPage(w http.ResponseWriter, statusCode int, statusMessage string) {

	type UserToDelete struct {
		Name   string
		IsAuth bool
	}

	data := struct {
		Code    int
		Message string
		User    UserToDelete
	}{
		Code:    statusCode,
		Message: statusMessage,
	}

	err := utils.RenderTemplate(w, "error.html", data, statusCode)
	if err != nil {
		app.Log.Error(err.Error())
	}
}

func (app *Application) Neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			app.ErrorPage(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) SaveUserSession(token string) error {

	user, err := app.Store.User.GetUserBySession(token)
	if err != nil {
		return err
	}

	userForHandler := User{
		ID:     user.ID,
		Name:   user.Username,
		IsAuth: true,
		Token:  token,
	}

	app.UserSessionCache[token] = userForHandler
	return nil
}

func (app *Application) GetUserSession(r *http.Request) (User, error) {
	tokenCookie, err := r.Cookie("auth_token")
	if err != nil || tokenCookie.Value == "" {
		return User{}, err
	}

	user, ok := app.UserSessionCache[tokenCookie.Value]
	if !ok {
		return User{}, fmt.Errorf("no registered user session")
	}
	return user, nil
}

func (app *Application) ServerErr(w http.ResponseWriter, err error) {
	app.ErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	app.Log.Error(err.Error())
}
