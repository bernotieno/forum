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
	ParentID  sql.NullInt64
	Author    string
	Content   string
	Likes     int
	Dislikes  int
	UserVote  sql.NullString
	Timestamp time.Time
	Replies   []Comment `json:"replies,omitempty"`
	Depth     int       `json:"depth"`
}

type CommentRequest struct {
	Content  string `json:"content"`
	ParentID int    `json:"parentId,omitempty"`
}
