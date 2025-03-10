package models

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type VotedComment struct {
	CommentID      int64
	CommentContent string
	UserVote       string
}

// Post represents a forum post
type Post struct {
	ID             int
	IsAuthor       bool
	Title          string
	Author         string
	UserID         int
	Category       string
	Likes          int
	Dislikes       int
	UserVote       sql.NullString
	Content        string
	ImageUrl       sql.NullString
	Timestamp      time.Time
	Comments       []Comment
	CommentCount   int
	VoteSource     string
	VotedCommentID int64
	VotedComments  []VotedComment
}

type PostRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories string `json:"category"`
}
