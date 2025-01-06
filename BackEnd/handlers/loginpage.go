package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login_page" {
		http.NotFound(w, r)
		return
	}
	if strings.ToUpper(r.Method) != "GET" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	tmp, err := template.ParseFiles("FrontEnd/Templates/login.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid template", http.StatusInternalServerError)
		return
	}

	err1 := tmp.Execute(w, nil)
	if err1 != nil {
		http.Error(w, "Failed execution", http.StatusInternalServerError)
		return
	}
}
