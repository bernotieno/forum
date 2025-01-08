package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type ReplyController struct {
	DB *sql.DB
}

func NewReplyController(db *sql.DB) *ReplyController {
	return &ReplyController{DB: db}
}

func (rc *ReplyController) InsertReply(reply models.Reply) (int, error) {
	result, err := rc.DB.Exec(`
		INSERT INTO replies (comment_id, user_id, author, content, likes, dislikes, user_vote, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`, reply.CommentID, reply.UserID, reply.Author, reply.Content, reply.Likes, reply.Dislikes, reply.UserVote, reply.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to insert reply: %w", err)
	}

	replyID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(replyID), nil
}
