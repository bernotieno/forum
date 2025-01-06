package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// RegisterHandler registers a new user
func RegisterHandler(ac *controllers.AuthController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Warning("Invalid method %s for registration attempt", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode registration request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}

		logger.Debug("Registration attempt for email: %s, username: %s", req.Email, req.Username)

		// Validate email
		if !ac.IsValidEmail(req.Email) {
			logger.Warning("Invalid email format attempted: %s", req.Email)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid email format",
			})
			return
		}

		// Validate username
		if !ac.IsValidUsername(req.Username) {
			logger.Warning("Invalid username format attempted: %s", req.Username)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Username must be between 3 and 20 characters and contain only letters, numbers, and underscores",
			})
			return
		}

		// Validate password
		if !ac.IsValidPassword(req.Password) {
			logger.Warning("Invalid password format for user: %s", req.Username)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Password must be at least 8 characters long and include uppercase, lowercase, numbers, and special characters",
			})
			return
		}

		// Sanitize inputs
		sanitizedEmail := ac.SanitizeInput(req.Email)
		sanitizedUsername := ac.SanitizeInput(req.Username)

		userID, err := ac.RegisterUser(sanitizedEmail, sanitizedUsername, req.Password)
		if err != nil {
			logger.Error("Registration failed for user %s: %v", sanitizedUsername, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		ac.CreateSession(w, int(userID))
		logger.Info("Successfully registered user: %s (ID: %d)", sanitizedUsername, userID)

		w.WriteHeader(302)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"redirect": "/",
		})
	}
}

// LoginHandler authenticates and creates a session
func LoginHandler(ac *controllers.AuthController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Warning("Invalid method %s for login attempt", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode login request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}

		logger.Debug("Login attempt for username: %s", req.Username)

		user, err := ac.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			logger.Warning("Failed login attempt for user %s: %v", req.Username, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		ac.CreateSession(w, user.ID)
		logger.Info("Successful login for user: %s (ID: %d)", user.Username, user.ID)

		w.WriteHeader(302)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"redirect": "/",
		})
	}
}

func isLoggedIn(r *http.Request) (bool, int) {
	// Get the session_token cookie from the request
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false, 0 // No cookie found
	}

	// Check if the session_token exists in the Sessions map
	userID, exists := controllers.Sessions[cookie.Value]
	if !exists {
		return false, 0 // Invalid session_token
	}

	// User is logged in, return true and the user's ID
	return true, userID
}

func CheckLoginHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn, userID := isLoggedIn(r)

	if !loggedIn {
		logger.Debug("Unauthorized access attempt detected")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unauthorized",
		})
		return
	}

	logger.Debug("Verified logged-in status for user ID: %d", userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User is logged in",
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		logger.Debug("Logout attempted with no active session")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "No active session",
		})
		return
	}

	userID := controllers.Sessions[cookie.Value]
	controllers.DeleteSession(w, cookie)
	logger.Info("User (ID: %d) successfully logged out", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully logged out",
	})
}
