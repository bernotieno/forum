package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func SetupNotificationRoutes(db *sql.DB) {
	notificationController := controllers.NewNotificationController(db)

	http.Handle("/api/notifications", middleware.ApplyMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlers.GetNotificationsHandler(notificationController).ServeHTTP(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/api/notifications", http.MethodGet),
	))
	http.Handle("/api/notifications/clear", middleware.ApplyMiddleware(
		handlers.ClearNotificationsHandler(notificationController),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/api/notifications/clear", http.MethodDelete),
	))
	http.Handle("/api/notifications/read", middleware.ApplyMiddleware(
		handlers.MarkNotificationAsReadHandler(notificationController),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/api/notifications/read", http.MethodPut),
	))
}
