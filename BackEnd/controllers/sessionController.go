package controllers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"
)

// Helper function to check if a session is valid
func IsValidSession(db *sql.DB, sessionToken string) (int, bool) {
	userID, expiresAt, err := GetSession(db, sessionToken)
	if err != nil {
		return userID, false
	}

	// Check if the session has expired
	if time.Now().After(expiresAt) {
		// Delete the expired session
		_ = DeleteSession(db, sessionToken)
		return userID, false
	}

	return userID, true
}

func GetSessionToken(r *http.Request) (string, error) {
	sessionToken, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	return sessionToken.Value, err
}

// AddSession adds a new session to the database
func AddSession(db *sql.DB, sessionToken string, userID int, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)",
		sessionToken, userID, expiresAt)
	return err
}

// GetSession retrieves session data from the database
func GetSession(db *sql.DB, sessionToken string) (int, time.Time, error) {
	var userID int
	var expiresAt time.Time
	err := db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE session_token = ?", sessionToken).
		Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}

// DeleteSession deletes a session from the database
func DeleteSession(db *sql.DB, sessionToken string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE session_token = ?", sessionToken)
	return err
}

// DeleteExpiredSessions deletes all expired sessions from the database
func DeleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

func DeleteUserSessions(db *sql.DB, userID int) error {
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// Cleanup expired sessions (replace with your actual cleanup logic)
func CleanupExpiredSessions(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping session cleanup task...")
			return
		case <-ticker.C:
			log.Println("Cleaning up expired sessions...")
			_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
			if err != nil {
				log.Printf("Failed to clean up expired sessions: %v\n", err)
			}
		}
	}
}
