package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strings"
	"time"
)

type userDBMethods struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type User struct {
	ID       int
	Username string
	Email    string
	Role     string
	Password string
	Creation time.Time
}

type UsersForPromotion struct {
	ID   int
	Name string
}

type UserDBInterface interface {
	CreateUser(username, email, role, password, dateOfCreation string) error
	AuthenticateUser(identifier, password string) (bool, error)
	GetUser(email string) (User, error)
	GetUserByID(id int) (User, error)
	CreateSessionInDB(userID int) (string, error)
	CheckToken(token string) (bool, error)
	GetUserBySession(token string) (User, error)
	DeleteUserSession(token string) error
	DeletePreviousUserSession(user_id int) error
	UserExistsByEmail(email string) (bool, error)
	PromoteMe(userID int) error
	GetPromotionRequests() ([]UsersForPromotion, error)
	PromoteToModer(userID int) error
	DemoteToUser(userID int) error
}

func DataUserWorkerCreation(db *sql.DB, logger *slog.Logger) *userDBMethods {
	return &userDBMethods{
		DB:     db,
		Logger: logger,
	}
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
	query := "SELECT u.id, u.username, u.email, u.role FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token=?;"
	err := u.DB.QueryRow(query, token).Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (u *userDBMethods) CreateUser(username, email, password, role, dateOfCreation string) error {

	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	var count int

	query := "SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)"
	err := u.DB.QueryRow(query, username, email).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	if count > 0 {
		return errors.New("username or email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	query = "INSERT INTO users (username, email, role, password_hash, date_of_creation) VALUES (?, ?, ?, ?, ?)"
	_, err = u.DB.Exec(query, username, email, role, hashedPassword, dateOfCreation)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (u *userDBMethods) AuthenticateUser(identifier, password string) (bool, error) {

	var storedHash string

	query := "SELECT password_hash FROM users WHERE email = ?"
	err := u.DB.QueryRow(query, identifier).Scan(&storedHash)

	if err == sql.ErrNoRows {
		u.Logger.Warn("user not found:", identifier)
		return false, nil
	} else if err != nil {
		u.Logger.Error("database error:", err)
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		u.Logger.Debug("password mismatch")
		return false, nil
	}

	u.Logger.Info("authentication successful for:", identifier)
	return true, nil
}

func (u *userDBMethods) GetUser(email string) (User, error) {
	var user User
	query := "SELECT id, username, email, role, password_hash FROM users WHERE email = ?;"
	err := u.DB.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Password)
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

func (u *userDBMethods) UserExistsByEmail(email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = ?"
	err := u.DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u *userDBMethods) GetUserByID(id int) (User, error) {
	var user User
	query := "SELECT username, email FROM users WHERE id = ?;"
	err := u.DB.QueryRow(query, id).Scan(&user.Username, &user.Email)
	if err == sql.ErrNoRows {
		return User{}, nil
	}
	return user, err
}

func (u *userDBMethods) PromoteMe(userID int) error {
	query := `DELETE FROM moderator_requests WHERE user_id = ?`
	_, err := u.DB.Exec(query, userID)

	query = `	
	INSERT INTO moderator_requests (user_id) VALUES (?)
	`
	_, err = u.DB.Exec(query, userID)
	return err
}

func (u *userDBMethods) GetPromotionRequests() ([]UsersForPromotion, error) {
	query := `SELECT m.user_id, u.username
	FROM moderator_requests m
	JOIN users u ON u.id = m.user_id
	`
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UsersForPromotion
	for rows.Next() {
		var u UsersForPromotion
		err = rows.Scan(&u.ID, &u.Name)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userDBMethods) PromoteToModer(userID int) error {
	query := `UPDATE users SET role = 'moderator' WHERE id = ?;`
	_, err := u.DB.Exec(query, userID)

	query = `DELETE FROM moderator_requests WHERE user_id = ?`
	_, err = u.DB.Exec(query, userID)
	return err
}

func (u *userDBMethods) DemoteToUser(userID int) error {
	query := `UPDATE users SET role = 'user' WHERE id = ?;`
	_, err := u.DB.Exec(query, userID)
	return err
}
