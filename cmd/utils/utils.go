package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	embed "forum/ui/html"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"unicode"
)

type User struct {
	Name   string
	IsAuth bool
}

var templates *template.Template
var logger *slog.Logger

func CachingTemplates() error {
	var err error
	templates, err = template.ParseFS(embed.HTMLFiles, "create.html", "error.html", "home.html", "login.html", "nav.html", "post.html", "signup.html", "user.html")
	if err != nil {
		return err
	}
	return nil
}

func RenderTemplate(w http.ResponseWriter, tmplName string, data any, statusCode int) error {
	if templates == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, tmplName, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logger.Error(err.Error())
		return err
	}

	w.WriteHeader(statusCode)
	_, err = buf.WriteTo(w)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func IsValidInput(input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return false
	}
	for _, r := range input {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false

	for _, c := range password {
		if unicode.IsLetter(c) {
			hasLetter = true
		} else if unicode.IsDigit(c) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func GenerateRandomPassword() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "defaultpassword123"
	}
	return hex.EncodeToString(bytes)
}
