package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		//ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		//ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		RedirectURL: "https://localhost:8443/auth/google/callback",
		Scopes:      []string{"email", "profile"},
		Endpoint:    google.Endpoint,
	}

	githubOauthConfig = &oauth2.Config{
		//ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		//ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),

		RedirectURL: "https://localhost:8443/auth/github/callback",
		Scopes:      []string{"user:email"},
		Endpoint:    github.Endpoint,
	}
)
