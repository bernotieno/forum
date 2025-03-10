package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type ActivityController struct {
	DB *sql.DB
}

func NewActivityController(db *sql.DB) *ActivityController {
	return &ActivityController{DB: db}
}

// GetUserPosts retrieves all posts created by a specific user
func (ac *ActivityController) GetUserPosts(userID int) ([]models.Post, error) {
	query := `
		SELECT id, title, user_id, author, category, likes, dislikes, 
			   user_vote, content, timestamp, image_url 
		FROM posts 
		WHERE user_id = ?
		ORDER BY timestamp DESC
	`
	rows, err := ac.DB.Query(query, userID)
	if err != nil {
		logger.Error("Database query failed in GetUserPosts: %v", err)
		return nil, fmt.Errorf("failed to fetch user posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID, &post.Title, &post.UserID, &post.Author,
			&post.Category, &post.Likes, &post.Dislikes,
			&post.UserVote, &post.Content, &post.Timestamp, &post.ImageUrl,
		)
		if err != nil {
			logger.Error("Row scan failed in GetUserPosts: %v", err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetUserVotedPosts retrieves all posts that a user has voted on (liked or disliked)
func (ac *ActivityController) GetUserVotedPosts(userID int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, p.user_id, p.author, p.category, p.likes, p.dislikes, 
			   l.user_vote, p.content, p.timestamp, p.image_url 
		FROM posts p
		JOIN likes l ON p.id = l.post_id
		WHERE l.user_id = ?
		ORDER BY p.timestamp DESC
	`
	rows, err := ac.DB.Query(query, userID)
	if err != nil {
		logger.Error("Database query failed in GetUserVotedPosts: %v", err)
		return nil, fmt.Errorf("failed to fetch user voted posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID, &post.Title, &post.UserID, &post.Author,
			&post.Category, &post.Likes, &post.Dislikes,
			&post.UserVote, &post.Content, &post.Timestamp, &post.ImageUrl,
		)
		if err != nil {
			logger.Error("Row scan failed in GetUserVotedPosts: %v", err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetUserComments retrieves all comments made by a user along with the associated post information
func (ac *ActivityController) GetUserComments(userID int) ([]models.UserCommentActivity, error) {
	query := `
		SELECT c.id, c.content, c.timestamp, c.likes, c.dislikes,
			   p.id, p.title, p.author, p.content
		FROM comments c
		JOIN posts p ON c.post_id = p.id
		WHERE c.user_id = ?
		ORDER BY c.timestamp DESC
	`
	rows, err := ac.DB.Query(query, userID)
	if err != nil {
		logger.Error("Database query failed in GetUserComments: %v", err)
		return nil, fmt.Errorf("failed to fetch user comments: %w", err)
	}
	defer rows.Close()

	var comments []models.UserCommentActivity
	for rows.Next() {
		var comment models.UserCommentActivity
		err := rows.Scan(
			&comment.CommentID, &comment.CommentContent, &comment.CommentTimestamp, 
			&comment.CommentLikes, &comment.CommentDislikes,
			&comment.PostID, &comment.PostTitle, &comment.PostAuthor, &comment.PostContent,
		)
		if err != nil {
			logger.Error("Row scan failed in GetUserComments: %v", err)
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
} 