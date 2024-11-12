package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
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
	CreateSessionInDB(userID int) (string, error)
	CheckToken(token string) (bool, error)
	GetUserBySession(token string) (User, error)
	DeleteUserSession(token string) error
	DeletePreviousUserSession(user_id int) error
}

func DataUserWorkerCreation(db *sql.DB) *userDBMethods {
	return &userDBMethods{DB: db}
}

func (u *userDBMethods) CreateSessionInDB(userID int) (string, error) {

	token, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	query := "INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?);"

	_, err = u.DB.Exec(query, token, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func (u *userDBMethods) CheckToken(token string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM sessions WHERE token = ?"
	err := u.DB.QueryRow(query, token).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u *userDBMethods) GetUserBySession(token string) (User, error) {
	var user User
	query := "SELECT u.id, u.username, u.email FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token=?;"
	err := u.DB.QueryRow(query, token).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (u *userDBMethods) CreateUser(username, email, password, dateOfCreation string) error {

	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	log.Printf("Normalized username: %s", username)
	log.Printf("Normalized email: %s", email)

	var count int

	query := "SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)"
	err := u.DB.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	log.Printf("User count: %d", count)

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
	_, err = u.DB.Exec(query, username, email, hashedPassword, dateOfCreation)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (u *userDBMethods) AuthenticateUser(identifier, password string) (bool, error) {
	var storedHash string

	log.Println("Attempting to authenticate:", identifier)

	query := "SELECT password_hash FROM users WHERE username = ? OR email = ?"
	err := u.DB.QueryRow(query, identifier, identifier).Scan(&storedHash)
	if err == sql.ErrNoRows {

		log.Println("user not found:", identifier)
		return false, nil
	} else if err != nil {
		log.Println("database error:", err)
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

func (u *userDBMethods) GetUser(username string) (User, error) {
	var user User
	query := "SELECT id, username, email, password_hash FROM users WHERE username = ?;"
	err := u.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		return User{}, nil
	}
	return user, err
}

func (u *userDBMethods) DeleteUserSession(token string) error {
	query := "DElETE FROM sessions WHERE token = ?"
	_, err := u.DB.Exec(query, token)
	if err != nil {
		return err
	}
	return nil
}

func (u *userDBMethods) DeletePreviousUserSession(user_id int) error {
	query := "DElETE FROM sessions WHERE user_id = ?"
	_, err := u.DB.Exec(query, user_id)
	if err != nil {
		return err
	}
	return nil
}
