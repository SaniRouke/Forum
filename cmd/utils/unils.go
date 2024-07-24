package utils

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

func ErrorPage(w http.ResponseWriter, statusCode int, statusMessage string) {
	tmpl, err := template.ParseFiles("./ui/html/error.html")

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.WriteHeader(statusCode)

	data := struct {
		Code    int
		Message string
	}{
		Code:    statusCode,
		Message: statusMessage,
	}
	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return
	}
}

func Neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			ErrorPage(w, http.StatusNotFound, "page not found")
			return
		}
		next.ServeHTTP(w, r)
	})
}
