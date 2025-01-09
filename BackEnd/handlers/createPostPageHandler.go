package handlers

import (
	"net/http"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
)

func CreatePostPageHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	loggedIn, userID := isLoggedIn(r)
	if !loggedIn {
		http.Redirect(w, r, "/login_Page", http.StatusSeeOther)
		return
	}

	// Generate CSRF token
	csrfToken := controllers.GenerateCSRFToken(userID)

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
