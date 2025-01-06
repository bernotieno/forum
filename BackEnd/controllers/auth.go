package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	DB *sql.DB
}

var Sessions = map[string]int{} // Map session token -> user ID

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{DB: db}
}

func (ac *AuthController) RegisterUser(email, username, password string) (int64, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	result, err := ac.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("email or username already taken")
	}

	// Get the auto-generated user ID
	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve user ID: %w", err)
	}

	// Return the user ID
	return userID, nil
}

func (ac *AuthController) AuthenticateUser(username, password string) (*models.User, error) {
	user := &models.User{}
	err := ac.DB.QueryRow("SELECT id, email, username, password FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid username")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// isValidEmail checks if the email is in a valid format
func (ac *AuthController) IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// isValidUsername checks if the username meets the requirements
func (ac *AuthController) IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	// Only allow letters, numbers, and underscores
	regex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return regex.MatchString(username)
}

// isValidPassword checks if the password meets the requirements
func (ac *AuthController) IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	// Check for at least one uppercase, one lowercase, one number, and one special character
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+{}|:"<>?~\-=[\]\\;',./]`).MatchString(password)
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// sanitizeInput removes potentially dangerous characters to prevent XSS
func (ac *AuthController) SanitizeInput(input string) string {
	// Replace HTML tags and special characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	return input
}

func (ac *AuthController) CreateSession(w http.ResponseWriter, userID int) {
	// Generate a new session token
	sessionToken := uuid.NewString()

	// Store the session token and user ID in the Sessions map
	Sessions[sessionToken] = userID

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		HttpOnly: true,                           // Prevent JavaScript access to the cookie
		Secure:   true,                           // Ensure the cookie is only sent over HTTPS
		Expires:  time.Now().Add(24 * time.Hour), // Set cookie expiration
	})
}

func DeleteSession(w http.ResponseWriter, cookie *http.Cookie) {
	// Retrieve the session token value
	sessionToken := cookie.Value

	// Delete the session from the Sessions map
	delete(Sessions, sessionToken)

	// Invalidate the cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		MaxAge: -1, // Deletes the cookie
	})
}
