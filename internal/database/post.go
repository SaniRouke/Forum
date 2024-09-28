package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

type postDBMethods struct {
	DB *sql.DB
}

type PostDBInterface interface {
	CreatePost(form CreatePostForm) error
	GetPost(id string) (Post, error)
	GetAll() ([]Post, error)
	AddComment(postID int, author string, commentBody string, date string) error
	GetComments(id string) ([]Comment, error)
	GetCategories() ([]string, error)
	GetPostsByCategory([]string) ([]Post, error)
}

type Post struct {
	ID       int
	Author   string
	Topic    string
	Body     string
	Date     string
	Comments []Comment
	Category string
	//likes / dislikes
}

type Comment struct {
	ID     int
	PostID int
	Author string
	Body   string
	Date   string
	//likes / dislikes
}

type CreatePostForm struct {
	Topic    string
	Body     string
	Category string
	Author   string
}

func DataPostWorkerCreation(db *sql.DB) *postDBMethods {
	return &postDBMethods{DB: db}
}

func (p *postDBMethods) CreatePost(form CreatePostForm) error {
	date := time.Now().Format("2006-01-02 15:04:05")
	query := "INSERT INTO posts (topic, body, category, author, date) VALUES (?, ?, ?, ?, ?);"
	_, err := p.DB.Exec(query, form.Topic, form.Body, form.Category, form.Author, date)
	return err
}

func (p *postDBMethods) GetAll() ([]Post, error) {
	query := "SELECT id, topic, body FROM posts ORDER BY date desc"
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

func (p *postDBMethods) GetPostsByCategory(categories []string) ([]Post, error) {

	if len(categories) == 0 {
		return nil, fmt.Errorf("Amount of categories should be greather than 0")
	}

	// Anime,Beer
	// sqlite> SELECT count(*) FROM posts WHERE category in (Beer, Altushki);
	// select count(*) from posts where category like "Beer" OR category like "Altushki";

	conditions := []string{}
	args := []any{}

	for _, category := range categories {
		conditions = append(conditions, "category LIKE ?")
		args = append(args, "%"+category+"%")
	}

	query := `
		SELECT id, topic FROM posts 
		WHERE ` + strings.Join(conditions, " OR ") + `
		ORDER BY date DESC;
	`

	fmt.Println(query)

	rows, err := p.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.ID, &post.Topic); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil

}

func (p *postDBMethods) GetPost(id string) (Post, error) {
	var post Post
	query := "SELECT id, topic, body, category, date, author FROM posts WHERE id = ?;"
	err := p.DB.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body, &post.Category, &post.Date, &post.Author)
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

func (p *postDBMethods) GetCategories() ([]string, error) {

	categoryList := []string{}

	query := "SELECT category FROM categories"
	rows, err := p.DB.Query(query)
	for rows.Next() {
		var oneCategory string
		if err = rows.Scan(&oneCategory); err != nil {
			return nil, err
		}
		categoryList = append(categoryList, oneCategory)
	}
	if err == sql.ErrNoRows {
		return nil, err
	}
	return categoryList, nil
}
