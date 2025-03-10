// Package controllers provides functionality for handling activity-related operations
package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// ActivityController handles user activity-related operations
type ActivityController struct {
	DB *sql.DB
}

// NewActivityController creates a new ActivityController instance
func NewActivityController(db *sql.DB) *ActivityController {
	return &ActivityController{DB: db}
}

// GetUserPosts retrieves all posts created by a specific user
// It returns a slice of Post models and any error encountered
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
// It returns both posts the user voted on directly and posts containing comments the user voted on
// The results include the vote type (like/dislike) and any voted comments
func (ac *ActivityController) GetUserVotedPosts(userID int) ([]models.Post, error) {
	query := `
		WITH UserVotes AS (
			SELECT 
				p.id, p.title, p.user_id, p.author, p.category, p.likes, p.dislikes,
				p.content, p.timestamp, p.image_url,
				l.user_vote as post_vote,
				c.id as comment_id,
				c.content as comment_content,
				cv.vote_type as comment_vote,
				CASE 
					WHEN l.user_id IS NOT NULL THEN 'post'
					WHEN cv.user_id IS NOT NULL THEN 'comment'
				END as vote_source
			FROM posts p
			LEFT JOIN likes l ON p.id = l.post_id AND l.user_id = ?
			LEFT JOIN comments c ON c.post_id = p.id
			LEFT JOIN comment_votes cv ON cv.comment_id = c.id AND cv.user_id = ?
			WHERE l.user_id IS NOT NULL OR cv.user_id IS NOT NULL
		)
		SELECT * FROM UserVotes
		ORDER BY timestamp DESC
	`
	rows, err := ac.DB.Query(query, userID, userID)
	if err != nil {
		logger.Error("Database query failed in GetUserVotedPosts: %v", err)
		return nil, fmt.Errorf("failed to fetch user voted posts: %w", err)
	}
	defer rows.Close()

	// Use a map to deduplicate posts while preserving voted comments
	postMap := make(map[int]*models.Post)
	var result []models.Post

	for rows.Next() {
		var (
			post           models.Post
			postVote       sql.NullString
			commentID      sql.NullInt64
			commentContent sql.NullString
			commentVote    sql.NullString
			voteSource     string
		)

		err := rows.Scan(
			&post.ID, &post.Title, &post.UserID, &post.Author,
			&post.Category, &post.Likes, &post.Dislikes,
			&post.Content, &post.Timestamp, &post.ImageUrl,
			&postVote, &commentID, &commentContent, &commentVote,
			&voteSource,
		)
		if err != nil {
			logger.Error("Row scan failed in GetUserVotedPosts: %v", err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		// If we haven't seen this post before, initialize it
		if _, exists := postMap[post.ID]; !exists {
			post.VoteSource = voteSource
			post.UserVote = postVote
			post.VotedComments = make([]models.VotedComment, 0)
			postMap[post.ID] = &post
		}

		// Add voted comment if it exists
		if commentID.Valid && commentVote.Valid {
			votedComment := models.VotedComment{
				CommentID:      commentID.Int64,
				CommentContent: commentContent.String,
				UserVote:       commentVote.String,
			}
			postMap[post.ID].VotedComments = append(postMap[post.ID].VotedComments, votedComment)
		}
	}

	// Convert map to slice for return
	for _, post := range postMap {
		result = append(result, *post)
	}

	return result, nil
}

// GetUserComments retrieves all comments made by a user along with the associated post information
// It returns a slice of UserCommentActivity models containing both comment and parent post details
func (ac *ActivityController) GetUserComments(userID int) ([]models.UserCommentActivity, error) {
	query := `		SELECT c.id, c.content, c.timestamp, c.likes, c.dislikes,
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
