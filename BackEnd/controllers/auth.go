package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		return 0, errors.New("internal server error")
	}

	result, err := ac.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
	if err != nil {
		logger.Warning("Registration failed - duplicate email or username: %v", err)
		return 0, errors.New("email or username already taken")
	}

	// Get the auto-generated user ID
	userID, err := result.LastInsertId()
	if err != nil {
		logger.Error("Failed to retrieve user ID after registration: %v", err)
		return 0, errors.New("failed to complete registration")
	}

	// Return the user ID
	return userID, nil
}

func (ac *AuthController) AuthenticateUser(username, password string) (*models.User, error) {
	user := &models.User{}
	err := ac.DB.QueryRow("SELECT id, email, username, password FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		logger.Warning("Authentication failed - invalid username: %s", username)
		return nil, errors.New("invalid username")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Warning("Authentication failed - invalid password for user: %s", username)
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// isValidEmail checks if the email is in a valid format
func (ac *AuthController) IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	if err != nil {
		logger.Debug("Invalid email format: %s", email)
		return false
	}
	return true
}

// isValidUsername checks if the username meets the requirements
func (ac *AuthController) IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		logger.Debug("Invalid username length: %s", username)
		return false
	}
	// Only allow letters, numbers, and underscores
	regex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !regex.MatchString(username) {
		logger.Debug("Username contains invalid characters: %s", username)
		return false
	}
	return true
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
	
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		logger.Debug("Password does not meet complexity requirements")
		return false
	}
	return true
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
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
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
