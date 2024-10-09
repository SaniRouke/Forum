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
	SetReaction(postID int, userID int, reaction int) error
	CheckReaction(postID int, userID int) (int, error)
	UpdateReaction(postID int, userID int, reaction int) error
	DeleteReaction(postID int, userID int) error
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

//if err == sql.ErrNoRows { // TODO: ?
//return Post{}, nil
//}

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

	query := "SELECT p.id, p.topic, p.body, u.username, p.date, p.category FROM posts AS p JOIN users AS u ON u.id = p.user_id WHERE p.id = ?;"
	err := p.DB.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body, &post.Author, &post.Date, &post.Category)
	if err == sql.ErrNoRows {
		return Post{}, nil
	}
	return post, err
}

func (p *postDBMethods) GetComments(id string) ([]Comment, error) {
	query := "SELECT c.id, c.post_id, c.body, u.username, c.date FROM comments AS c JOIN users AS u ON u.id = c.user_id WHERE c.post_id = ?"
	rows, err := p.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment
		if err = rows.Scan(&comment.ID, &comment.PostID, &comment.Body, &comment.Author, &comment.Date); err != nil {
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

func (p *postDBMethods) SetReaction(postID int, userID int, reaction int) error {
	query := "INSERT INTO reactions_posts (post_id, user_id, reaction) VALUES (?, ?, ?)"
	_, err := p.DB.Exec(query, postID, userID, reaction)
	if err != nil {
		return err
	}
	return nil
}

func (p *postDBMethods) CheckReaction(postID int, userID int) (int, error) {
	var reaction int
	query := "SELECT reaction FROM reactions_posts WHERE post_id = ? AND user_id = ?"
	err := p.DB.QueryRow(query, postID, userID).Scan(&reaction)
	if err != nil {
		return 0, err
	}
	return reaction, nil
}

func (p *postDBMethods) UpdateReaction(postID int, userID int, reaction int) error {
	query := "UPDATE reactions_posts SET reaction = ? WHERE post_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, reaction, postID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postDBMethods) DeleteReaction(postID int, userID int) error {
	query := "DELETE FROM reactions_posts WHERE post_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, postID, userID)
	if err != nil {
		return err
	}

	return nil
}
