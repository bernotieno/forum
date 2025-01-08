package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type CommentController struct {
	DB *sql.DB
}

func NewCommentController(db *sql.DB) *CommentController {
	return &CommentController{DB: db}
}

func (cCtrl *CommentController) InsertComment(comment models.Comment) (int, error) {
	result, err := cCtrl.DB.Exec(`
		INSERT INTO comments (post_id, user_id, author, content, likes, dislikes, user_vote, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`, comment.PostID, comment.UserID, comment.Author, comment.Content, comment.Likes, comment.Dislikes, comment.UserVote, comment.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(commentID), nil
}
