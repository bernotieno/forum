package handlers

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login_Page" {
		logger.Warning("Invalid path access attempt: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	if strings.ToUpper(r.Method) != "GET" {
		logger.Warning("Invalid method %s for login page access", r.Method)
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	loggedIn, _ := isLoggedIn(database.GloabalDB, r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmp, err := template.ParseFiles("FrontEnd/templates/login.html")
	if err != nil {
		logger.Error("Failed to parse login template: %v", err)
		http.Error(w, "Invalid template", http.StatusInternalServerError)
		return
	}

	logger.Debug("Executing login template")
	if err := tmp.Execute(w, nil); err != nil {
		logger.Error("Failed to execute login template: %v", err)
		http.Error(w, "Failed execution", http.StatusInternalServerError)
		return
	}
}
