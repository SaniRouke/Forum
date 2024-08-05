package internal

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Post struct {
	ID    int
	Topic string
	Body  string
}

func CreatePost(topic string, body string) error {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	query := "INSERT INTO posts (topic, body) VALUES (?, ?);"
	_, err = db.Exec(query, topic, body)
	if err != nil {
		return err
	}

	return nil
}

func GetAllPosts() []Post {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var posts []Post

	query := "SELECT id, topic, body FROM posts;"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Topic, &post.Body)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	return posts
}

func GetPost(id string) (Post, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return Post{}, err
	}
	defer db.Close()

	var post Post

	query := "SELECT id, topic, body FROM posts WHERE id = ?;"
	err = db.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body)
	if err != nil {
		if err == sql.ErrNoRows {
			return Post{}, nil
		}
		return Post{}, err
	}

	return post, nil
}

func UpdatePost(id int, topic string, body string) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	updateUserSQL := `UPDATE posts SET topic = ?, body = ? WHERE id = ?`
	_, err = db.Exec(updateUserSQL, topic, body, id)
	if err != nil {
		log.Fatal(err)
	}
}

func DeletePost(id int) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	deleteUserSQL := `DELETE FROM posts WHERE id = ?`
	_, err = db.Exec(deleteUserSQL, id)
	if err != nil {
		log.Fatal(err)
	}
}
