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
	GetPost(id string, userID int) (Post, error)
	GetAll() ([]Post, error)
	AddComment(postID int, userID int, commentBody string, date string) error
	GetComments(id string, userID int) ([]Comment, error)
	GetCategories() ([]string, error)
	GetPostsByCategory([]string) ([]Post, error)
	GetHotTopics() ([]Post, error)
	GetPostsByUser(userID int) ([]Post, error)
	GetPostsWithUserComments(userID int) ([]Post, error)
	GetPostsWithUserReactions(userID int) ([]Post, error)
	SetPostReaction(postID int, userID int, reaction int) error
	CheckPostReaction(postID int, userID int) (int, error)
	UpdatePostReaction(postID int, userID int, reaction int) error
	DeletePostReaction(postID int, userID int) error
	SetCommentReaction(commentID int, userID int, reaction int) error
	CheckCommentReaction(commentID int, userID int) (int, error)
	UpdateCommentReaction(commentID int, userID int, reaction int) error
	DeleteCommentReaction(commentID int, userID int) error
}

type Post struct {
	ID           int
	Author       string
	Topic        string
	Body         string
	Date         string
	Comments     []Comment
	Category     string
	Likes        int
	Dislikes     int
	UserLiked    bool
	UserDisliked bool
}

type Comment struct {
	ID           int
	PostID       int
	Author       string
	Body         string
	Date         string
	Likes        int
	Dislikes     int
	UserLiked    bool
	UserDisliked bool
}

type CreatePostForm struct {
	Topic    string
	Body     string
	Category string
	UserID   int
}

func DataPostWorkerCreation(db *sql.DB) *postDBMethods {
	return &postDBMethods{DB: db}
}

//if err == sql.ErrNoRows { // TODO: ?
//return Post{}, nil
//}

func (p *postDBMethods) CreatePost(form CreatePostForm) error {
	date := time.Now().Format("2006-01-02 15:04:05")
	query := "INSERT INTO posts (topic, body, category, user_id, date) VALUES (?, ?, ?, ?, ?);"
	_, err := p.DB.Exec(query, form.Topic, form.Body, form.Category, form.UserID, date)
	return err
}

func (p *postDBMethods) GetAll() ([]Post, error) {
	query := `
    SELECT p.id, p.topic, p.date, u.username, p.category,
           COALESCE(SUM(CASE WHEN rp.reaction = 1 THEN 1 ELSE 0 END), 0) AS Likes,
           COALESCE(SUM(CASE WHEN rp.reaction = -1 THEN 1 ELSE 0 END), 0) AS Dislikes
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN reactions_posts rp ON rp.post_id = p.id
    GROUP BY p.id
    ORDER BY p.date DESC;
`
	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post

		if err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category, &post.Likes, &post.Dislikes); err != nil {
			return nil, err
		}
		postDate, err := time.Parse("2006-01-02 15:04:05", post.Date)
		if err == nil {
			post.Date = postDate.Format("02.01.2006, 15:04")
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (p *postDBMethods) GetPostsByCategory(categories []string) ([]Post, error) {

	if len(categories) == 0 {
		return nil, fmt.Errorf("Amount of categories should be greather than 0")
	}

	conditions := []string{}
	args := []any{}

	for _, category := range categories {
		conditions = append(conditions, "category LIKE ?")
		args = append(args, "%"+category+"%")
	}

	query := `
    SELECT p.id, p.topic, p.date, u.username, p.category,
           COALESCE(SUM(CASE WHEN rp.reaction = 1 THEN 1 ELSE 0 END), 0) AS Likes,
           COALESCE(SUM(CASE WHEN rp.reaction = -1 THEN 1 ELSE 0 END), 0) AS Dislikes
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN reactions_posts rp ON rp.post_id = p.id
    WHERE ` + strings.Join(conditions, " OR ") + `
    GROUP BY p.id
    ORDER BY p.date DESC;
`

	rows, err := p.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category, &post.Likes, &post.Dislikes); err != nil {
			return nil, err
		}
		postDate, err := time.Parse("2006-01-02 15:04:05", post.Date)
		if err == nil {
			post.Date = postDate.Format("02.01.2006, 15:04")
		}
		posts = append(posts, post)
	}
	return posts, nil

}

func (p *postDBMethods) GetPost(id string, userID int) (Post, error) {

	var post Post

	query := `SELECT p.id, p.topic, p.body, u.username, p.date, p.category,
       COALESCE(SUM(CASE WHEN r.reaction = 1 THEN 1 ELSE 0 END), 0) AS Likes,
       COALESCE(SUM(CASE WHEN r.reaction = -1 THEN 1 ELSE 0 END), 0) AS Dislikes
       FROM posts AS p 
       JOIN users AS u ON u.id = p.user_id 
       LEFT JOIN reactions_posts AS r ON r.post_id = p.id
       WHERE p.id = ?
	   GROUP BY p.id;`

	err := p.DB.QueryRow(query, id).Scan(&post.ID, &post.Topic, &post.Body, &post.Author, &post.Date, &post.Category, &post.Likes, &post.Dislikes)
	if err == sql.ErrNoRows {
		return Post{}, nil
	}
	postDate, err := time.Parse("2006-01-02 15:04:05", post.Date)
	if err == nil {
		post.Date = postDate.Format("02.01.2006, 15:04")
	}

	if userID != 0 {
		userLiked, userDisliked, err := p.GetUserPostReaction(post.ID, userID)
		if err != nil {
			return Post{}, err
		}
		post.UserLiked = userLiked
		post.UserDisliked = userDisliked
	}

	return post, err
}

func (p *postDBMethods) GetComments(id string, userID int) ([]Comment, error) {
	query := `
    SELECT c.id, c.post_id, c.body, u.username, c.date,
    COALESCE(SUM(CASE WHEN r.reaction = 1 THEN 1 ELSE 0 END), 0) AS Likes,
    COALESCE(SUM(CASE WHEN r.reaction = -1 THEN 1 ELSE 0 END), 0) AS Dislikes
    FROM comments AS c 
    JOIN users AS u ON u.id = c.user_id
    LEFT JOIN reactions_comments AS r ON r.comment_id = c.id
    WHERE c.post_id = ?
	GROUP BY c.id, c.post_id, c.body, u.username, c.date;
`
	rows, err := p.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment
		if err = rows.Scan(&comment.ID, &comment.PostID, &comment.Body, &comment.Author, &comment.Date, &comment.Likes, &comment.Dislikes); err != nil {
			return nil, err
		}
		commentDate, err := time.Parse("2006-01-02 15:04:05", comment.Date)
		if err == nil {
			comment.Date = commentDate.Format("02.01.2006, 15:04")
		}
		if userID != 0 {
			userLiked, userDisliked, err := p.GetUserCommentReaction(comment.ID, userID)
			if err != nil {
				return nil, err
			}
			comment.UserLiked = userLiked
			comment.UserDisliked = userDisliked
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (p *postDBMethods) AddComment(postID int, userID int, commentBody string, date string) error {
	query := "INSERT INTO comments (post_id, user_id, body, date) VALUES (?, ?, ?, ?)"
	_, err := p.DB.Exec(query, postID, userID, commentBody, date)
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

func (p *postDBMethods) GetUserPostReaction(postID int, userID int) (bool, bool, error) {
	var reaction int
	query := "SELECT reaction FROM reactions_posts WHERE post_id = ? AND user_id = ?"
	err := p.DB.QueryRow(query, postID, userID).Scan(&reaction)

	if err == sql.ErrNoRows {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	var userLiked bool
	if reaction == 1 {
		userLiked = true
	}
	userDisliked := reaction == -1

	return userLiked, userDisliked, nil
}

func (p *postDBMethods) SetPostReaction(postID int, userID int, reaction int) error {
	query := "INSERT INTO reactions_posts (post_id, user_id, reaction) VALUES (?, ?, ?)"
	_, err := p.DB.Exec(query, postID, userID, reaction)
	if err != nil {
		return err
	}
	return nil
}

func (p *postDBMethods) CheckPostReaction(postID int, userID int) (int, error) {
	var reaction int
	query := "SELECT reaction FROM reactions_posts WHERE post_id = ? AND user_id = ?"
	err := p.DB.QueryRow(query, postID, userID).Scan(&reaction)
	if err != nil {
		return 0, err
	}
	return reaction, nil
}

func (p *postDBMethods) UpdatePostReaction(postID int, userID int, reaction int) error {
	query := "UPDATE reactions_posts SET reaction = ? WHERE post_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, reaction, postID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postDBMethods) DeletePostReaction(postID int, userID int) error {
	query := "DELETE FROM reactions_posts WHERE post_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, postID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postDBMethods) GetUserCommentReaction(commentID int, userID int) (bool, bool, error) {
	var reaction int
	query := "SELECT reaction FROM reactions_comments WHERE comment_id = ? AND user_id = ?"
	err := p.DB.QueryRow(query, commentID, userID).Scan(&reaction)

	if err == sql.ErrNoRows {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	userLiked := reaction == 1
	userDisliked := reaction == -1

	return userLiked, userDisliked, nil
}

func (p *postDBMethods) SetCommentReaction(commentID int, userID int, reaction int) error {
	query := "INSERT INTO reactions_comments (comment_id, user_id, reaction) VALUES (?, ?, ?)"
	_, err := p.DB.Exec(query, commentID, userID, reaction)
	if err != nil {
		return err
	}
	return nil
}

func (p *postDBMethods) CheckCommentReaction(commentID int, userID int) (int, error) {
	var reaction int
	query := "SELECT reaction FROM reactions_comments WHERE comment_id = ? AND user_id = ?"
	err := p.DB.QueryRow(query, commentID, userID).Scan(&reaction)
	if err != nil {
		return 0, err
	}
	return reaction, nil
}

func (p *postDBMethods) UpdateCommentReaction(commentID int, userID int, reaction int) error {
	query := "UPDATE reactions_comments SET reaction = ? WHERE comment_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, reaction, commentID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postDBMethods) DeleteCommentReaction(commentID int, userID int) error {
	query := "DELETE FROM reactions_comments WHERE comment_id = ? AND user_id = ?"
	_, err := p.DB.Exec(query, commentID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postDBMethods) GetHotTopics() ([]Post, error) {
	query := `
	SELECT p.id, p.topic, p.date, u.username, p.category
	FROM posts p
	JOIN users u ON u.id = p.user_id
	LEFT JOIN comments c ON c.post_id = p.id
	GROUP BY p.id
	ORDER BY MAX(c.date) DESC
	LIMIT 5;
	`
	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotPosts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category)
		if err != nil {
			return nil, err
		}
		hotPosts = append(hotPosts, post)
	}
	return hotPosts, nil
}

func (p *postDBMethods) GetPostsByUser(userID int) ([]Post, error) {
	query := `
	SELECT p.id, p.topic, p.date, u.username, p.category
	FROM posts p 
	JOIN users u ON u.id = p.user_id
	WHERE p.user_id = ?
	ORDER BY p.date DESC`
	rows, err := p.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userPosts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category)
		if err != nil {
			return nil, err
		}
		userPosts = append(userPosts, post)
	}
	return userPosts, nil

}

func (p *postDBMethods) GetPostsWithUserComments(userID int) ([]Post, error) {
	query := `
	SELECT DISTINCT p.id, p.topic, p.date, u.username, p.category
	FROM posts p 
	JOIN users u ON u.id = p.user_id
	JOIN comments c ON c.post_id = p.id
	WHERE c.user_id = ?
	ORDER BY p.date DESC`

	rows, err := p.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var commentedPosts []Post

	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category)
		if err != nil {
			return nil, err
		}
		commentedPosts = append(commentedPosts, post)
	}
	return commentedPosts, nil
}

func (p *postDBMethods) GetPostsWithUserReactions(userID int) ([]Post, error) {
	query := `
	SELECT DISTINCT p.id, p.topic, p.date, u.username, p.category
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN reactions_posts rp ON rp.post_id = p.id
    LEFT JOIN reactions_comments rc ON rc.comment_id IN (SELECT id FROM comments WHERE post_id = p.id)
    WHERE rp.user_id = ? OR rc.user_id = ?
    ORDER BY p.date DESC;
    `

	rows, err := p.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likedPosts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Topic, &post.Date, &post.Author, &post.Category)
		if err != nil {
			return nil, err
		}
		likedPosts = append(likedPosts, post)
	}
	return likedPosts, nil
}
