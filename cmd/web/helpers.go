package main

import (
	"fmt"
	"forum/internal/database"
	"net/http"
)

func GetUserFromContext(r *http.Request) (User, error) {
	var userForTemplate User
	user, ok := r.Context().Value("user").(database.User)
	if !ok {
		return userForTemplate, fmt.Errorf("Юзер-хуюзер не найден")
	} else {
		userForTemplate.Name = user.Username
		userForTemplate.IsAuth = true
	}
	return userForTemplate, nil
}
