package main

import (
	"fmt"
	"forum/internal/database"
	"net/http"
)

func (app *Application) SaveUserSession(token string) error {

	user, err := app.Store.User.GetUserBySession(token)
	if err != nil {
		fmt.Println("УЗЕР")
		return err
	}

	userForHandler := User{
		ID:     user.ID,
		Name:   user.Username,
		IsAuth: true,
		Token:  token,
	}

	app.UserSessionCache[token] = userForHandler
	fmt.Println(app.UserSessionCache)
	fmt.Println("ФИНИШ ХУИНИШ")
	return nil
}

func (app *Application) GetUserSession(r *http.Request) (User, error) {
	fmt.Println(app.UserSessionCache)
	tokenCookie, err := r.Cookie("auth_token")
	if err != nil || tokenCookie.Value == "" {
		return User{}, err
	}

	user, ok := app.UserSessionCache[tokenCookie.Value]
	if !ok {
		return User{}, fmt.Errorf("Нету юзера, нету сессии, ну типа того")
	}
	return user, nil
}

func GetUserFromContext(r *http.Request) (User, error) {
	var userForTemplate User
	user, ok := r.Context().Value("user").(database.User)
	if !ok {
		return userForTemplate, fmt.Errorf("Юзер-хуюзер не найден Арара")
	} else {
		userForTemplate.Name = user.Username
		userForTemplate.IsAuth = true
	}
	return userForTemplate, nil
}
