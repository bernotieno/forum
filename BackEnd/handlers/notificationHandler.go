package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func GetNotificationsHandler(nc *controllers.NotificationController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(nc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to Get Notifications - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to Get Notifications",
			})
			return
		}

		notifications, unreadCount, err := nc.GetNotifications(userID)
		if err != nil {
			logger.Error("Failed to fetch notifications in GetNotificationsHandler %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"notifications": notifications,
			"unread":        unreadCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func MarkNotificationAsReadHandler(nc *controllers.NotificationController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(nc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to Read Notification - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to Read Notification",
			})
			return
		}
		notificationID := r.URL.Query().Get("notificationId")

		notificationIDInt, err := strconv.Atoi(notificationID)
		if err != nil {
			logger.Error("Invalid notification ID in MarkNotificationAsReadHandler %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		unreadCount, err := nc.MarkNotificationAsRead(notificationIDInt, userID)
		if err != nil {
			logger.Error("Failed to mark notification as read in MarkNotificationAsReadHandler %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"unread_count": unreadCount,
		})
	}
}

func ClearNotificationsHandler(nc *controllers.NotificationController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn, userID := isLoggedIn(nc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to Read Notification - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to Read Notification",
			})
			return
		}

		err := nc.ClearNotifications(userID)
		if err != nil {
			logger.Error("Failed to clear notifications in ClearNotificationsHandler %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"success": "All notifications cleared successfully",
		})
	}
}
