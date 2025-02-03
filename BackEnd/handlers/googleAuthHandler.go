package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/auth"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

const (
	clientID     = "1085346236158-jnh86upn77ed2hqvklm9nprasps7k7j5.apps.googleusercontent.com"
	clientSecret = "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"
	redirectURI  = "http://localhost:8080/auth/google/callback"
	authURL      = "https://accounts.google.com/o/oauth2/auth"
	tokenURL     = "https://oauth2.googleapis.com/token"
	userInfoURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// Handle Google login: Generate the Google OAuth URL
func GoogleHandler(w http.ResponseWriter, r *http.Request) {
	authLink := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile&access_type=offline", clientID,
		url.QueryEscape(redirectURI))

	http.Redirect(w, r, authLink, http.StatusFound)
}

// Handle the callback: Exchange the code for an access token and fetch user info
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		logger.Error("No code in the request: ")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input 1",
		})
		return
	}

	token, err := exchangeCodeForToken(code)
	if err != nil {
		logger.Error("Failed to exchange token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input 2",
		})
		return
	}

	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		logger.Error("Failed to fetch user info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input 3",
		})
		return
	}

	db := database.GloabalDB
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = db.QueryRow(
				"INSERT INTO users (email, username, password) VALUES (?, ?, ?) RETURNING id",
				userInfo.Email, userInfo.Name, "_",
			).Scan(&userID)
			if err != nil {
				logger.Error("failed to create user: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid input 4",
				})
				return
			}
		} else {
			logger.Error("Failed to retrieve user ID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input 5",
			})
			return
		}
	}

	err = auth.CreateSession(db, w, userID)
	if err != nil {
		logger.Error("failed to create session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid input 6",
		})
		return
	}

	// Use template for redirection
	tmpl, err := template.ParseFiles("FrontEnd/templates/oauth_callback.html")
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		logger.Error("Failed to execute template: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func exchangeCodeForToken(code string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Add("code", code)
	data.Add("client_id", clientID)
	data.Add("client_secret", clientSecret)
	data.Add("redirect_uri", redirectURI)
	data.Add("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getUserInfo fetches user info using the access token
func getUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// TokenResponse represents the response from the token endpoint
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
}
