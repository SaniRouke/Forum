package internal

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
)

var DB *sql.DB

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

type Post struct {
	ID    int
	Topic string
	Body  string
}

func InitializeDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	return nil
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

func CreatePost(topic, body string) error {
	query := "INSERT INTO posts (topic, body) VALUES (?, ?);"
	_, err := DB.Exec(query, topic, body)
	return err
}

func GetAllPosts() ([]Post, error) {
	query := "SELECT id, topic, body FROM posts;"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.ID, &post.Topic, &post.Body); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPost(id string) (Post, error) {
	var post Post
	query := "SELECT id, topic, body FROM posts WHERE id = ?;"
	err := DB.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body)
	if err == sql.ErrNoRows {
		return Post{}, nil
	}
	return post, err
}

func EditPost(id string, topic, body string) error {
	query := "UPDATE posts SET topic = ?, body = ? WHERE id = ?;"
	_, err := DB.Exec(query, topic, body, id)
	return err
}

func DeletePost(id string) error {
	query := "DELETE FROM posts WHERE id = ?;"
	_, err := DB.Exec(query, id)
	return err
}
