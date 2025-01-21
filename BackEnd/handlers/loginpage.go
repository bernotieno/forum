package handlers

import (
	"net/http"
	"text/template"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn, _ := isLoggedIn(database.GloabalDB, r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmp, err := template.ParseFiles("FrontEnd/templates/login.html")
	if err != nil {
		logger.Error("Failed to parse login template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debug("Executing login template")
	if err := tmp.Execute(w, nil); err != nil {
		logger.Error("Failed to execute login template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
