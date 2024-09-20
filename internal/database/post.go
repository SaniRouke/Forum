package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type postDBMethods struct {
	DB *sql.DB
}
type PostDBInterface interface {
	CreatePost(topic, body, author string) error
	GetPost(id string) (Post, error)
	GetAll() ([]Post, error)
	AddComment(postID int, author string, commentBody string, date string) error
	GetComments(id string) ([]Comment, error)
}

type Post struct {
	ID       int
	Author   string
	Topic    string
	Body     string
	Date     string
	Comments []Comment
	// Likes
}

type Comment struct {
	ID     int
	PostID int
	Author string
	Body   string
	Date   string
	// Likes
}

func DataPostWorkerCreation(db *sql.DB) *postDBMethods {
	return &postDBMethods{DB: db}
}

func (p *postDBMethods) CreatePost(topic, body, author string) error {
	date := time.Now().Format("2006-01-02 15:04:05")
	query := "INSERT INTO posts (topic, body, author, date) VALUES (?, ?, ?, ?);"
	_, err := p.DB.Exec(query, topic, body, author, date)
	return err
}

func (p *postDBMethods) GetAll() ([]Post, error) {
	query := "SELECT id, topic, body FROM posts;"
	rows, err := p.DB.Query(query)
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

func (p *postDBMethods) GetPost(id string) (Post, error) {
	var post Post
	query := "SELECT id, topic, body, author, date FROM posts WHERE id = ?;"
	err := p.DB.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body, &post.Author, &post.Date)
	if err == sql.ErrNoRows {
		return Post{}, nil
	}
	return post, err
}

func (p *postDBMethods) GetComments(id string) ([]Comment, error) {
	query := "SELECT id, post_id, author, body, date FROM comments WHERE post_id = ?"
	rows, err := p.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment
		if err = rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Body, &comment.Date); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (p *postDBMethods) AddComment(postID int, author string, commentBody string, date string) error {
	query := "INSERT INTO comments (post_id, author, body, date) VALUES (?, ?, ?, ?)"
	_, err := p.DB.Exec(query, postID, author, commentBody, date)
	return err
}
