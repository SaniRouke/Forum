package database

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
)

type userDBMethods struct {
	DB *sql.DB
}

type User struct {
	ID       int
	Username string
	Email    string
	Password string
	Creation time.Time
}

type UserDBInterface interface {
	CreateUser(username, email, password, dateOfCreation string) error
	AuthenticateUser(identifier, password string) (bool, error)
	GetUser(username string) (User, error)
}

func DataUserWorkerCreation(db *sql.DB) *userDBMethods {
	return &userDBMethods{DB: db}
}

func (p *userDBMethods) CreateUser(username, email, password, dateOfCreation string) error {
	// Normalize email and username by trimming spaces and converting to lowercase
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	// Log normalized values for debugging
	log.Printf("Normalized username: %s", username)
	log.Printf("Normalized email: %s", email)

	// Check if username or email already exists
	var count int
	//var existingEmail sql.NullString
	//
	// Adjust query to handle cases where the email might be empty
	//query := "SELECT COUNT(*), email FROM users WHERE LOWER(username) = LOWER(?) OR email = ?"
	query := "SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)"
	err := p.DB.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	// Debug output to understand what's being retrieved
	log.Printf("User count: %d", count)
	//log.Printf("Existing email: %v", existingEmail.String)

	if count > 0 {
		return errors.New("username or email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Insert the user into the database
	query = "INSERT INTO users (username, email, password_hash, date_of_creation) VALUES (?, ?, ?, ?)"
	_, err = p.DB.Exec(query, username, email, hashedPassword, dateOfCreation)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (p *userDBMethods) AuthenticateUser(identifier, password string) (bool, error) {
	var storedHash string

	log.Println("Attempting to authenticate:", identifier)

	query := "SELECT password_hash FROM users WHERE username = ? OR email = ?"
	err := p.DB.QueryRow(query, identifier, identifier).Scan(&storedHash)
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

func (p *userDBMethods) GetUser(username string) (User, error) {
	var user User
	query := "SELECT id, username, email, password_hash FROM users WHERE username = ?;"
	err := p.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		return User{}, nil
	}
	return user, err
}
