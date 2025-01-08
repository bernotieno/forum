package models

import (
	"database/sql"
	"time"
)

// Comment represents a comment on a post
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Author    string
	Content   string
	Likes     int
	Dislikes  int
	UserVote  sql.NullString // Can be "like", "dislike", or null
	Timestamp time.Time
}

type CommentRequest struct {
	Content string `json:"content"`
}
