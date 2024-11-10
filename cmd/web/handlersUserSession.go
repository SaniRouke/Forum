package main

import (
	"forum/cmd/utils"
	"log"
	"net/http"
	"regexp"
	"time"
)

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
			app.Log.Error.Println(err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (app *Application) handlerLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		user, err := app.GetUserSession(r)
		if err != nil {
			log.Println(err)
		}

		data := struct {
			User User
		}{
			User: user,
		}

		err = utils.RenderTemplate(w, "login.html", data, http.StatusOK)
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
			utils.ErrorPage(w, http.StatusUnauthorized, "User not found")
			log.Println(err)
			return
		}

		err = app.Store.User.DeletePreviousUserSession(user.ID)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		token, err := app.Store.User.CreateSessionInDB(user.ID)
		if err != nil {
			utils.ErrorPage(w, http.StatusUnauthorized, "You shall not pass!")
			log.Println(err)
			return
		}

		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60 * 60 * 24,
		}

		http.SetCookie(w, cookie)

		app.SaveUserSession(token)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *Application) handlerLogout(w http.ResponseWriter, r *http.Request) {

	user, err := app.GetUserSession(r)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	err = app.Store.User.DeleteUserSession(user.Token)
	if err != nil {
		log.Println(err)
		return
	}

	delete(app.UserSessionCache, user.Token)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
