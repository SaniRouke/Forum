package database

import (
	"database/sql"
	"log/slog"
)

type notificationDBMethods struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type NotificationDBInterface interface {
	CreateNotification(notification Notification) error
	GetNotificationsByUserID(userID int) ([]Notification, error)
	MarkNotificationAsRead(notificationID int) error
	GetUnreadNotificationCount(userID int) (int, error)
	GetNotificationByID(notificationID int) (Notification, error)
}

type Notification struct {
	ID                int
	UserID            int
	InitiatorUserID   int
	PostID            int
	CommentID         *int
	Type              string
	Read              bool
	Date              string
	InitiatorUsername string
}

func DataNotificationWorkerCreation(db *sql.DB, logger *slog.Logger) *notificationDBMethods {
	return &notificationDBMethods{
		DB:     db,
		Logger: logger,
	}
}

func (n *notificationDBMethods) CreateNotification(notification Notification) error {
	query := `
    INSERT INTO notifications (user_id, initiator_user_id, post_id, comment_id, type, read, date)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := n.DB.Exec(query, notification.UserID, notification.InitiatorUserID, notification.PostID, notification.CommentID, notification.Type, 0, notification.Date)
	return err
}

func (n *notificationDBMethods) GetNotificationsByUserID(userID int) ([]Notification, error) {
	query := `
    SELECT n.id, n.user_id, n.initiator_user_id, n.post_id, n.comment_id, n.type, n.read, n.date, u.username
    FROM notifications n
    JOIN users u ON u.id = n.initiator_user_id
    WHERE n.user_id = ?
    ORDER BY n.date DESC
    `
	rows, err := n.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		var commentID sql.NullInt64
		var readInt int
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.InitiatorUserID, &notification.PostID, &commentID, &notification.Type, &readInt, &notification.Date, &notification.InitiatorUsername); err != nil {
			return nil, err
		}
		if commentID.Valid {
			cid := int(commentID.Int64)
			notification.CommentID = &cid
		}
		notification.Read = readInt == 1
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (n *notificationDBMethods) MarkNotificationAsRead(notificationID int) error {
	query := "UPDATE notifications SET read = 1 WHERE id = ?"
	_, err := n.DB.Exec(query, notificationID)
	return err
}

func (n *notificationDBMethods) GetUnreadNotificationCount(userID int) (int, error) {
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0"
	var count int
	err := n.DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (n *notificationDBMethods) GetNotificationByID(notificationID int) (Notification, error) {
	query := `
	SELECT n.id, n.user_id, n.initiator_user_id, n.post_id, n.comment_id, n.type, n.read, n.date, u.username
	FROM notifications n
	JOIN users u ON u.id = n.initiator_user_id
	WHERE n.id = ?
	`
	var notification Notification
	var commentID sql.NullInt64
	var readInt int

	err := n.DB.QueryRow(query, notificationID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.InitiatorUserID,
		&notification.PostID,
		&commentID,
		&notification.Type,
		&readInt,
		&notification.Date,
		&notification.InitiatorUsername,
	)
	if err != nil {
		return Notification{}, err
	}

	if commentID.Valid {
		cid := int(commentID.Int64)
		notification.CommentID = &cid
	}
	notification.Read = readInt == 1
	return notification, nil
}
