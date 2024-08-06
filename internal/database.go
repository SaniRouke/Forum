package internal

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

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

type Post struct {
	ID    int
	Topic string
	Body  string
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

func UpdatePost(id int, topic, body string) error {
	query := "UPDATE posts SET topic = ?, body = ? WHERE id = ?;"
	_, err := DB.Exec(query, topic, body, id)
	return err
}

func DeletePost(id int) error {
	query := "DELETE FROM posts WHERE id = ?;"
	_, err := DB.Exec(query, id)
	return err
}
