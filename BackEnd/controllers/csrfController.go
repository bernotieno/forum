package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/Raymond9734/forum.git/BackEnd/auth"
)

// Store CSRF tokens with a map
var csrfTokens = struct {
	sync.RWMutex
	tokens map[int]string
}{tokens: make(map[int]string)}

func GenerateCSRFToken(userID int) string {
	// Generate random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	// Convert to base64
	token := base64.URLEncoding.EncodeToString(b)

	// Store the token
	csrfTokens.Lock()
	csrfTokens.tokens[userID] = token
	csrfTokens.Unlock()

	return token
}

func VerifyCSRFToken(r *http.Request) bool {
	// First check the CSRF token from the header
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		// If not found in header, get from form
		token = r.FormValue("csrf_token")
		if token == "" {
			return false
		}
	}

	// Get userID from session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	userID, exists := auth.Sessions[cookie.Value]
	if !exists {
		return false
	}

	// Verify token
	csrfTokens.RLock()
	storedToken, exists := csrfTokens.tokens[userID]
	csrfTokens.RUnlock()

	return exists && storedToken == token
}
