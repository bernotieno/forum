package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type NotificationController struct {
	DB *sql.DB
}

func NewNotificationController(db *sql.DB) *NotificationController {
	return &NotificationController{DB: db}
}

func (nc *NotificationController) CreateNotification(notification models.Notification) (int, error) {
	query := `
        INSERT INTO notifications (recipient_id, actor_id, type, entity_type, message)
        VALUES (?, ?, ?, ?, ?)
        RETURNING id, created_at
    `

	err := nc.DB.QueryRow(
		query,
		notification.RecipientID,
		notification.ActorID,
		notification.Type,
		notification.EntityType,
		notification.Message,
	).Scan(&notification.ID, &notification.CreatedAt)

	return notification.ID, err
}

func (nc *NotificationController) GetNotifications(userID int) ([]models.Notification, int, error) {
	query := `
        SELECT n.id, n.recipient_id, n.actor_id, n.type, n.entity_type, 
              n.message, n.is_read, n.created_at,
               u.username as actor_username, u.email as actor_email
        FROM notifications n
        LEFT JOIN users u ON n.actor_id = u.id
        WHERE n.recipient_id = ?
        ORDER BY n.created_at DESC
    `

	rows, err := nc.DB.Query(query, userID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var actor models.User
		err := rows.Scan(
			&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.EntityType,
			&n.Message, &n.IsRead, &n.CreatedAt,
			&actor.Username, &actor.Email,
		)
		if err != nil {
			continue
		}
		n.Actor = &actor
		notifications = append(notifications, n)
	}

	// Get unread count
	var unreadCount int
	err = nc.DB.QueryRow(
		"SELECT COUNT(*) FROM notifications WHERE recipient_id = ? AND is_read = FALSE",
		userID,
	).Scan(&unreadCount)
	if err != nil {
		return notifications, 0, err
	}

	return notifications, unreadCount, nil
}

func (nc *NotificationController) MarkNotificationAsRead(notificationID, userID int) (int, error) {
	result, err := nc.DB.Exec(
		"UPDATE notifications SET is_read = TRUE WHERE id = ? AND recipient_id = ?",
		notificationID, userID,
	)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return 0, fmt.Errorf("notification not found")
	}

	// Get updated unread count
	var unreadCount int
	err = nc.DB.QueryRow(
		"SELECT COUNT(*) FROM notifications WHERE recipient_id = ? AND is_read = FALSE",
		userID,
	).Scan(&unreadCount)
	if err != nil {
		return 0, err
	}

	return unreadCount, nil
}

func (nc *NotificationController) ClearNotifications(userID int) error {
	_, err := nc.DB.Exec("DELETE FROM notifications WHERE recipient_id = ?", userID)
	return err
}
