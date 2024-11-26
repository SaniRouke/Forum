package main

import (
	"context"
	"encoding/json"
	"forum/cmd/utils"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"time"
)

func (app *Application) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *Application) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		app.Log.Error("Google token exchange failed:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		app.Log.Error("Failed to get user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.Log.Error("Failed to read response:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(content, &userInfo); err != nil {
		app.Log.Error("Failed to parse user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	err = app.processOAuthUser(userInfo.Email, userInfo.Name)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	app.createSessionAndRedirect(w, r, userInfo.Email)
}

// GitHub Login Handler
func (app *Application) handleGithubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *Application) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		app.Log.Error("GitHub token exchange failed:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		app.Log.Error("Failed to get user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		app.Log.Error("Failed to parse user info:", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// GitHub may not provide email directly
	if userInfo.Email == "" {
		emailsResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			app.Log.Error("Failed to get user emails:", err)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		defer emailsResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}

		if err := json.NewDecoder(emailsResp.Body).Decode(&emails); err != nil {
			app.Log.Error("Failed to parse emails:", err)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		for _, email := range emails {
			if email.Primary && email.Verified {
				userInfo.Email = email.Email
				break
			}
		}
	}

	if userInfo.Email == "" {
		app.ErrorPage(w, http.StatusBadRequest, "GitHub email not found")
		return
	}

	username := userInfo.Name
	if username == "" {
		username = userInfo.Login
	}

	err = app.processOAuthUser(userInfo.Email, username)
	if err != nil {
		app.ServerErr(w, err)
		return
	}

	app.createSessionAndRedirect(w, r, userInfo.Email)
}

func (app *Application) processOAuthUser(email, username string) error {
	userExists, err := app.Store.User.UserExistsByEmail(email)
	if err != nil {
		return err
	}

	if !userExists {
		password := utils.GenerateRandomPassword()
		dateOfCreation := time.Now().Format("2006-01-02 15:04:05")
		err := app.Store.User.CreateUser(username, email, password, dateOfCreation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) createSessionAndRedirect(w http.ResponseWriter, r *http.Request, email string) {
	user, err := app.Store.User.GetUser(email)
	if err != nil {
		app.Log.Error("Failed to get user:", err)
		app.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = app.Store.User.DeletePreviousUserSession(user.ID)
	if err != nil {
		app.Log.Error("Failed to delete previous session:", err)
	}

	token, err := app.Store.User.CreateSessionInDB(user.ID)
	if err != nil {
		app.Log.Error("Failed to create session:", err)
		app.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	app.SaveUserSession(token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
