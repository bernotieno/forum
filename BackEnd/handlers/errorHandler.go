package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

// ErrorPageData holds the data for the error page template
type ErrorPageData struct {
	ErrorCode    int
	ErrorTitle   string
	ErrorMessage string
}

// serveErrorPage renders the error page with the provided error code, title, and message
func ServeErrorPage(w http.ResponseWriter, errorCode int, errorTitle, errorMessage string) {
	tmpl, err := template.ParseFiles("FrontEnd/templates/errorPage.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	data := ErrorPageData{
		ErrorCode:    errorCode,
		ErrorTitle:   errorTitle,
		ErrorMessage: errorMessage,
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Execute template without writing status
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Println(err)
		return
	}
}
