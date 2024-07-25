package internal

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func Create(name string, email string) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "INSERT INTO users (name, email) VALUES (?, ?);"
	_, err = db.Exec(query, name, email)
	if err != nil {
		log.Fatal(err)
	}
}

func Read() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//result, err := db.Exec(query)

	var users []struct {
		id    int
		name  string
		email string
	}
	query := "select * from users;"
	rows, err := db.Query(query)

	for rows.Next() {
		var user struct {
			id    int
			name  string
			email string
		}
		err := rows.Scan(&user.id, &user.name, &user.email)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	fmt.Println(users)
}
func Update(name string, email string, id int) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	updateUserSQL := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err = db.Exec(updateUserSQL, name, email, id)
	if err != nil {
		log.Fatal(err)
	}
}
func Delete(id int) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	deleteUserSQL := `DELETE FROM users WHERE id = ?`
	_, err = db.Exec(deleteUserSQL, id)
	if err != nil {
		log.Fatal(err)
	}
}
