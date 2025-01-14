package controllers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

func GenerateCSRFToken(db *sql.DB, sessionToken string) (string, error) {
	// Generate random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Convert to base64
	token := base64.URLEncoding.EncodeToString(b)

	// Set expiration time (e.g., 1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	fmt.Println("==Token Generated==", token)

	// Store the token in the database
	err := AddCSRFToken(db, sessionToken, token, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func VerifyCSRFToken(db *sql.DB, r *http.Request) bool {
	// Get the CSRF token from the request
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		// If not found in header, get from form
		token = r.FormValue("csrf_token")
		if token == "" {
			return false
		}
	}
	fmt.Println("==Token Recieved==", token)

	// Get userID from session
	cookie, err := r.Cookie("session_token")
	fmt.Println("==Cookie Recieved==", cookie)
	if err != nil {
		return false
	}

	userID, exists := IsValidSession(db, cookie.Value)
	fmt.Println("==UserID Recieved==", userID)
	if !exists {
		return false
	}

	// Retrieve the stored token from the database
	storedToken, expiresAt, err := GetCSRFToken(db, cookie.Value)
	fmt.Println("==Stored Token Recieved==", storedToken)
	if err != nil {
		return false
	}

	// Check if the token matches and is not expired
	if storedToken != token || time.Now().After(expiresAt) {
		fmt.Println("==Token Mismatch==")
		// Delete the expired or invalid token
		_ = DeleteCSRFToken(db, cookie.Value)
		return false
	}

	return true
}

// AddCSRFToken stores a new CSRF token in the database
func AddCSRFToken(db *sql.DB, sessionToken string, csrfToken string, expiresAt time.Time) error {
	_, err := db.Exec("INSERT OR REPLACE INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		sessionToken, csrfToken, expiresAt)
	return err
}

// GetCSRFToken retrieves the CSRF token for a session from the database
func GetCSRFToken(db *sql.DB, sessionToken string) (string, time.Time, error) {
	var token string
	var expiresAt time.Time
	err := db.QueryRow("SELECT csrf_token, expires_at FROM csrf_tokens WHERE session_token = ?", sessionToken).
		Scan(&token, &expiresAt)
	return token, expiresAt, err
}

// DeleteCSRFToken deletes a CSRF token from the database
func DeleteCSRFToken(db *sql.DB, sessionToken string) error {
	_, err := db.Exec("DELETE FROM csrf_tokens WHERE session_token = ?", sessionToken)
	return err
}

// Cleanup expired CSRF tokens (replace with your actual cleanup logic)
func CleanupExpiredCSRFTokens(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping CSRF token cleanup task...")
			return
		case <-ticker.C:
			log.Println("Cleaning up expired CSRF tokens...")
			_, err := db.Exec("DELETE FROM csrf_tokens WHERE expires_at < ?", time.Now())
			if err != nil {
				log.Printf("Failed to clean up expired CSRF tokens: %v\n", err)
			}
		}
	}
}
