package controllers

import (
	"database/sql"
	"fmt"
)

type CommentVotesController struct {
	DB *sql.DB
}

func NewCommentVotesController(db *sql.DB) *CommentVotesController {
	return &CommentVotesController{DB: db}
}

func (cc *CommentVotesController) UpdateCommentVotes(commentID int) error {
	var commentExists bool
	err := cc.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)`, commentID).Scan(&commentExists)
	if err != nil {
		return fmt.Errorf("failed to check if comment exists: %w", err)
	}
	if !commentExists {
		return fmt.Errorf("comment with ID %d does not exist", commentID)
	}
	query := `
        SELECT 
            COALESCE(SUM(CASE WHEN vote_type = 'like' THEN 1 ELSE 0 END), 0) AS likes_count,
            COALESCE(SUM(CASE WHEN vote_type = 'dislike' THEN 1 ELSE 0 END), 0) AS dislikes_count
        FROM comment_votes
        WHERE comment_id = ?;
    `

	var likesCount, dislikesCount int
	err = cc.DB.QueryRow(query, commentID).Scan(&likesCount, &dislikesCount)
	if err != nil {
		return fmt.Errorf("failed to retrieve vote counts for comment %d: %w", commentID, err)
	}

	updateQuery := `
        UPDATE comments
        SET likes = ?, dislikes = ?
        WHERE id = ?;
    `
	_, err = cc.DB.Exec(updateQuery, likesCount, dislikesCount, commentID)
	if err != nil {
		return fmt.Errorf("failed to update comment votes: %w", err)
	}

	return nil
}

func (cc *CommentVotesController) GetCommentVotes(commentID int) (int, int, error) {
	query := `SELECT likes, dislikes FROM comments WHERE id = ?;`
	var likes, dislikes sql.NullInt64
	err := cc.DB.QueryRow(query, commentID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, err
	}
	return int(likes.Int64), int(dislikes.Int64), nil
}

func (cc *CommentVotesController) HandleCommentVote(commentID, userID int, voteType string) error {
	// Check if user has already voted
	existingVote, err := cc.CheckUserVote(commentID, userID)
	if err != nil {
		return err
	}

	if existingVote == voteType {
		// Remove vote if same type
		err = cc.RemoveUserVote(commentID, userID)
	} else {
		// Remove existing vote if any
		if existingVote != "" {
			err = cc.RemoveUserVote(commentID, userID)
			if err != nil {
				return err
			}
		}
		// Add new vote
		err = cc.AddUserVote(commentID, userID, voteType)
	}

	if err != nil {
		return err
	}

	// Update vote counts
	return cc.UpdateCommentVotes(commentID)
}

func (cc *CommentVotesController) CheckUserVote(commentID, userID int) (string, error) {
	var voteType string
	query := `SELECT vote_type FROM comment_votes WHERE comment_id = ? AND user_id = ?`
	err := cc.DB.QueryRow(query, commentID, userID).Scan(&voteType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return voteType, err
}

func (cc *CommentVotesController) RemoveUserVote(commentID, userID int) error {
	query := `DELETE FROM comment_votes WHERE comment_id = ? AND user_id = ?`
	_, err := cc.DB.Exec(query, commentID, userID)
	return err
}

func (cc *CommentVotesController) AddUserVote(commentID, userID int, voteType string) error {
	query := `INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (?, ?, ?)`
	_, err := cc.DB.Exec(query, commentID, userID, voteType)
	return err
}
