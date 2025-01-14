package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/google/uuid"
)

// CreateSession creates a new session for a user
func CreateSession(db *sql.DB, w http.ResponseWriter, userID int) error {
	// Delete all existing sessions for the user
	err := controllers.DeleteUserSessions(db, userID)
	if err != nil {
		return err
	}

	// Create a new session
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	err = controllers.AddSession(db, sessionToken, userID, expiresAt)
	if err != nil {
		return err
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	return nil
}

func DeleteSession(db *sql.DB, w http.ResponseWriter, cookie *http.Cookie) {
	sessionToken := cookie.Value

	// Delete the session from the database
	err := controllers.DeleteSession(db, sessionToken)
	if err != nil {
		logger.Error("Failed to delete session: %v", err)
	}

	// Invalidate the cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		MaxAge: -1,
	})
}
