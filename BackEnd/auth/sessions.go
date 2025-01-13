package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/google/uuid"
)

var Sessions = map[string]int{} // Map session token -> user ID

// Session store using standard Go map with mutex for thread safety
var (
	SessionStore = struct {
		sync.RWMutex
		Sessions map[string]sessionData
	}{
		Sessions: make(map[string]sessionData),
	}
)

type sessionData struct {
	UserID    int
	ExpiresAt time.Time
}

// Helper function to check if a session is valid
func IsValidSession(token string) bool {
	SessionStore.RLock()
	defer SessionStore.RUnlock()

	session, exists := SessionStore.Sessions[token]
	if !exists || time.Now().After(session.ExpiresAt) {
		return false
	}
	return true
}

// CreateSession creates a new session for a user
func CreateSession(w http.ResponseWriter, userID int) {
	sessionToken := uuid.New().String()

	SessionStore.Lock()
	SessionStore.Sessions[sessionToken] = sessionData{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	Sessions[sessionToken] = userID
	SessionStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})
}

func DeleteSession(w http.ResponseWriter, cookie *http.Cookie) {
	// Retrieve the session token value
	sessionToken := cookie.Value
	if userID, exists := Sessions[sessionToken]; exists {
		logger.Info("Deleting session for user ID: %d", userID)
	}

	delete(Sessions, sessionToken)

	// Invalidate the cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		MaxAge: -1,
	})
}
