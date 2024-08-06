package utils

import (
	"bytes"
	embed "forum/ui/html"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func RenderTemplate(w http.ResponseWriter, tmplName string, data any, statusCode int) error {
	templates, err := template.ParseFS(embed.HTMLFiles, tmplName, "nav.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return err
	}

	var buf bytes.Buffer
	err = templates.ExecuteTemplate(&buf, tmplName, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Print(err)
		return err
	}

	w.WriteHeader(statusCode)
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func ErrorPage(w http.ResponseWriter, statusCode int, statusMessage string) {
	data := struct {
		Code    int
		Message string
	}{
		Code:    statusCode,
		Message: statusMessage,
	}

	err := RenderTemplate(w, "error.html", data, statusCode)
	if err != nil {
		log.Print(err)
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
