package db

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

func CreateUser(username, email, password string) error {
	// Normalize email and username by trimming spaces and converting to lowercase
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	// Log normalized values for debugging
	log.Printf("Normalized username: %s", username)
	log.Printf("Normalized email: %s", email)

	// Check if username or email already existsy
	var count int
	var existingEmail sql.NullString

	// Adjust query to handle cases where the email might be empty
	query := "SELECT COUNT(*), email FROM users WHERE LOWER(username) = LOWER(?) OR email = ?"
	err := DB.QueryRow(query, username, email).Scan(&count, &existingEmail)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	// Debug output to understand what's being retrieved
	log.Printf("User count: %d", count)
	log.Printf("Existing email: %v", existingEmail.String)

	if count > 0 && (existingEmail.Valid && existingEmail.String == email) {
		return errors.New("username or email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Insert the user into the database
	query = "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)"
	_, err = DB.Exec(query, username, email, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func AuthenticateUser(identifier, password string) (bool, error) {
	var storedHash string

	log.Println("Attempting to authenticate:", identifier)

	query := "SELECT password_hash FROM users WHERE username = ? OR email = ?"
	err := DB.QueryRow(query, identifier, identifier).Scan(&storedHash)
	if err == sql.ErrNoRows {
		log.Println("User not found:", identifier)
		return false, nil
	} else if err != nil {
		log.Println("Database error:", err)
		return false, err
	}

	log.Println("Password hash found, comparing...")

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Println("Password mismatch")
		return false, nil
	}

	log.Println("Authentication successful for:", identifier)
	return true, nil
}
