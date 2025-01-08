package models

import (
	"database/sql"
	"time"
)

// Reply represents a reply to a comment
type Reply struct {
	ID        int
	CommentID int
	UserID    int
	Author    string
	Content   string
	Likes     int
	Dislikes  int
	UserVote  sql.NullString // Can be "like", "dislike", or null
	Timestamp time.Time
}
type ReplyRequest struct {
	Content string `json:"content"`
}
