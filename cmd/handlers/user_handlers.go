package hdl

import (
	"forum/cmd/utils"
	"forum/database"
	"log"
	"net/http"
)

func HandlerSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "signup.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := db.CreateUser(username, email, password)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Unable to create user")
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "login.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		identifier := r.FormValue("username") // Can be username or email
		password := r.FormValue("password")

		isAuthenticated, err := db.AuthenticateUser(identifier, password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Println("Error in authentication:", err)
			return
		}

		if !isAuthenticated {
			utils.ErrorPage(w, http.StatusBadRequest, "Invalid username or password")
			log.Println("Failed login attempt for:", identifier)
			return
		}

		// Redirect to home page on successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
