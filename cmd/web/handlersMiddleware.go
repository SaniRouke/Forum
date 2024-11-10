package main

import "net/http"

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

		//ctx := r.Context()
		//ctx = context.WithValue(ctx, "user", user)
		//r = r.WithContext(ctx)

		// Call the next handler with the updated request
		next(w, r)
	}
}
