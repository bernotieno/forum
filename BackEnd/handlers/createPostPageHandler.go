package handlers

import (
	"net/http"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func CreatePostPageHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	loggedIn, _ := isLoggedIn(database.GloabalDB, r)
	if !loggedIn {
		http.Redirect(w, r, "/login_Page", http.StatusSeeOther)
		return
	}

	sessioToken, err := controllers.GetSessionToken(r)
	if err != nil {
		logger.Error("Error getting session token: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate CSRF token
	csrfToken, err := controllers.GenerateCSRFToken(database.GloabalDB, sessioToken)
	if err != nil {
		logger.Error("Error generating CSRF token: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		IsAuthenticated bool
		CSRFToken       string
	}{
		IsAuthenticated: loggedIn,
		CSRFToken:       csrfToken,
	}

	tmpl, err := template.ParseFiles(
		"./FrontEnd/templates/layout.html",
		"./FrontEnd/templates/post.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
