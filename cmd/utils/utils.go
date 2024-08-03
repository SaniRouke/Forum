package utils

import (
	"bytes"
	embed "forum/ui/html"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func ErrorPage(w http.ResponseWriter, statusCode int, statusMessage string) {

	tmpl, err := template.ParseFS(embed.HTMLFiles, "error.html", "nav.html")

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	data := struct {
		Code    int
		Message string
	}{
		Code:    statusCode,
		Message: statusMessage,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "error.html", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	w.WriteHeader(statusCode)
	_, err = buf.WriteTo(w)
	if err != nil {
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
