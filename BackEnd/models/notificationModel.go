package models

type Notification struct {
	ID          int    `json:"id"`
	RecipientID int    `json:"recipient_id"`
	ActorID     int    `json:"actor_id"`
	Type        string `json:"type"`
	EntityType  string `json:"entity_type"`
	Message     string `json:"message"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
	Actor       *User  `json:"actor,omitempty"`
}
