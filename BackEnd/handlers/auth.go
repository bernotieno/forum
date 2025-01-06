package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// RegisterHandler registers a new user
func RegisterHandler(ac *controllers.AuthController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var req models.RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}
		// fmt.Println("Data Recieved Registration")
		// fmt.Println("Email: ", req.Email)
		// fmt.Println("Username: ", req.Username)
		// fmt.Println("Passwrd: ", req.Password)

		// Validate email
		if !ac.IsValidEmail(req.Email) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid email format",
			})
			return
		}

		// Validate username
		if !ac.IsValidUsername(req.Username) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Username must be between 3 and 20 characters and contain only letters, numbers, and underscores",
			})
			return
		}

		// Validate password
		if !ac.IsValidPassword(req.Password) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Password must be at least 8 characters long and include uppercase, lowercase, numbers, and special characters",
			})
			return
		}

		// Sanitize inputs to prevent XSS
		sanitizedEmail := ac.SanitizeInput(req.Email)
		sanitizedUsername := ac.SanitizeInput(req.Username)

		userID, err := ac.RegisterUser(sanitizedEmail, sanitizedUsername, req.Password)
		if err != nil {
			fmt.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		ac.CreateSession(w, int(userID))

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
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}
		fmt.Println("Data Recieved")
		fmt.Println("Username: ", req.Username)
		fmt.Println("Passwrd: ", req.Password)
		user, err := ac.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			fmt.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		// Create session cookie
		ac.CreateSession(w, user.ID)

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
	loggedIn, _ := isLoggedIn(r)

	if !loggedIn {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unauthorized",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User is logged in",
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session_token cookie from the request
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// No session token, user is already logged out or never logged in
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "No active session",
		})
		return
	}

	// Delete cookies session
	controllers.DeleteSession(w, cookie)

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully logged out",
	})
}
