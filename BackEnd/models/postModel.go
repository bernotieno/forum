package models

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Post represents a forum post
type Post struct {
	ID        int
	IsAuthor  bool
	Title     string
	Author    string
	UserID    int
	Category  string
	Likes     int
	Dislikes  int
	UserVote  sql.NullString // Can be "like", "dislike", or null
	Content   string
	ImageUrl  sql.NullString
	Timestamp time.Time
}

type PostRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories string `json:"category"`
}
