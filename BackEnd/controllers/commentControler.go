package controllers

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type CommentController struct {
	DB *sql.DB
}

func NewCommentController(db *sql.DB) *CommentController {
	return &CommentController{DB: db}
}

func (cCtrl *CommentController) InsertComment(comment models.Comment) (int, error) {
	// Add validation
	if len(comment.Content) == 0 {
		return 0, fmt.Errorf("comment content cannot be empty")
	}
	if len(comment.Content) > 3000 {
		return 0, fmt.Errorf("comment content too long")
	}

	result, err := cCtrl.DB.Exec(`
		INSERT INTO comments (post_id, user_id, author, content, likes, dislikes, user_vote, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`, comment.PostID, comment.UserID, comment.Author, comment.Content, comment.Likes, comment.Dislikes,
		comment.UserVote, comment.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(commentID), nil
}

func (cc *CommentController) GetCommentsByPostID(postID string) ([]models.Comment, error) {
	// Convert postID string to int
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %w", err)
	}

	// Get all comments for the post
	rows, err := cc.DB.Query(`
		SELECT id, post_id, user_id, author, content, likes, dislikes, 
			   user_vote, timestamp
		FROM comments 
		WHERE post_id = ?
		ORDER BY timestamp DESC
	`, postIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Author,
			&comment.Content, &comment.Likes, &comment.Dislikes,
			&comment.UserVote, &comment.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}
