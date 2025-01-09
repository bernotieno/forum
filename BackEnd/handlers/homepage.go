package handlers

import (
	"net/http"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Check if user is logged in
	loggedIn, userID := isLoggedIn(r)

	// Generate CSRF token for the session
	csrfToken := controllers.GenerateCSRFToken(userID)

	data := struct {
		IsAuthenticated bool
		CSRFToken       string
		// Add other necessary data
		User *models.User // For user-specific data
	}{
		IsAuthenticated: loggedIn,
		CSRFToken:       csrfToken,
	}

	// Set security headers
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://cdnjs.cloudflare.com; font-src 'self' https://cdnjs.cloudflare.com")

	tmpl, err := template.ParseFiles(
		"./FrontEnd/templates/layout.html",
		"./FrontEnd/templates/homepage.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

