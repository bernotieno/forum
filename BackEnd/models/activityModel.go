package models

import (
	"time"
)

// UserCommentActivity represents a user's comment along with the associated post
type UserCommentActivity struct {
	// Comment information
	CommentID        int
	CommentContent   string
	CommentTimestamp time.Time
	CommentLikes     int
	CommentDislikes  int
	
	// Associated post information
	PostID      int
	PostTitle   string
	PostAuthor  string
	PostContent string
}

// UserActivity represents all user activity data for the activity page
type UserActivity struct {
	UserPosts    []Post
	VotedPosts   []Post
	UserComments []UserCommentActivity
} 