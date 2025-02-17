package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/auth"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

const (
	githubClientID     = "Ov23liTW1Eb8CVoGBA7R"
	githubClientSecret = "ffc063cbbac3cc9f6e0d8babc9968f1fe8d6e1ff"
	githubRedirectURI  = "http://localhost:8080/auth/github/callback"
	githubAuthURL      = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserAPIURL   = "https://api.github.com/user"
)

// GitHubHandler redirects the user to GitHub OAuth
func GitHubHandler(w http.ResponseWriter, r *http.Request) {
	authLink := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=user:email",
		githubAuthURL, githubClientID, url.QueryEscape(githubRedirectURI))

	http.Redirect(w, r, authLink, http.StatusFound)
}

// GitHubCallbackHandler handles the callback from GitHub OAuth
func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		logger.Error("No code in the request")
		http.Error(w, `{"error": "Invalid input 1"}`, http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token
	token, err := exchangeGitHubCodeForToken(code)
	if err != nil {
		logger.Error("Failed to exchange token: %v", err)
		http.Error(w, `{"error": "Invalid input 2"}`, http.StatusInternalServerError)
		return
	}

	// Get GitHub user info
	userInfo, err := getGitHubUserInfo(token.AccessToken)
	if err != nil {
		logger.Error("Failed to fetch user info: %v", err)
		http.Error(w, `{"error": "Invalid input 3"}`, http.StatusInternalServerError)
		return
	}

	// Check if user exists in the database
	query := `SELECT id FROM users WHERE email = ?`
	db := database.GloabalDB
	var userID int
	err = db.QueryRow(query, userInfo.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist, create a new user
			query := `INSERT INTO users (email, username, password) VALUES (?, ?, ?) RETURNING id`
			err = db.QueryRow(query, userInfo.Email, userInfo.Login, "_").Scan(&userID)
			if err != nil {
				logger.Error("Failed to create user: %v", err)
				http.Error(w, `{"error": "Invalid input 4"}`, http.StatusInternalServerError)
				return
			}
		} else {
			logger.Error("Failed to retrieve user ID: %v", err)
			http.Error(w, `{"error": "Invalid input 5"}`, http.StatusInternalServerError)
			return
		}
	}

	// Create a session for the user
	err = auth.CreateSession(db, w, userID)
	if err != nil {
		logger.Error("Failed to create session: %v", err)
		http.Error(w, `{"error": "Invalid input 6"}`, http.StatusInternalServerError)
		return
	}

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

// exchangeGitHubCodeForToken exchanges the OAuth code for an access token
func exchangeGitHubCodeForToken(code string) (*GithubTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", githubClientID)
	data.Set("client_secret", githubClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", githubRedirectURI)

	req, err := http.NewRequest("POST", githubTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp GithubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getGitHubUserInfo fetches user information using the access token
func getGitHubUserInfo(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", githubUserAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// If email is missing, fetch emails separately
	if userInfo.Email == "" {
		emails, err := getGitHubEmails(accessToken)
		if err != nil {
			return nil, err
		}
		userInfo.Email = emails
	}

	return &userInfo, nil
}

// getGitHubEmails fetches the primary email from GitHub
func getGitHubEmails(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find the primary and verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// Fallback: return first email if no primary is found
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

// TokenResponse represents GitHub OAuth token response
type GithubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GitHubUser represents GitHub user information
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// GitHubEmail represents GitHub email response
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}
