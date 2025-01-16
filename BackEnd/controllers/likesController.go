package controllers

import (
	"database/sql"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type LikesController struct {
	DB *sql.DB
}

func NewLikesController(db *sql.DB) *LikesController {
	return &LikesController{DB: db}
}

func (Lc *LikesController) InsertLikes(like models.Likes) (int, error) {
	// Insert the like with the provided PostId, UserId, and UserVote
	result, err := Lc.DB.Exec(`
        INSERT INTO likes (post_id, user_id, user_vote)
        VALUES (?, ?, ?);
    `, like.PostId, like.UserId, like.UserVote)
	if err != nil {
		return 0, fmt.Errorf("failed to insert like: %w", err)
	}

	// Get the ID of the newly inserted like
	likeID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(likeID), nil
}

func (Lc *LikesController) UpdatePostVotes(postID int) error {
	// Query to calculate the likes and dislikes from the likes table
	query := `
        SELECT 
            SUM(CASE WHEN user_vote = 'like' THEN 1 ELSE 0 END) AS likes_count,
            SUM(CASE WHEN user_vote = 'dislike' THEN 1 ELSE 0 END) AS dislikes_count
        FROM likes
        WHERE post_id = ?;
    `

	var likesCount, dislikesCount int
	err := Lc.DB.QueryRow(query, postID).Scan(&likesCount, &dislikesCount)
	if err != nil {
		return fmt.Errorf("failed to retrieve vote counts for post %d: %w", postID, err)
	}

	// Update the likes and dislikes in the posts table
	updateQuery := `
        UPDATE posts
        SET likes = ?, dislikes = ?
        WHERE id = ?;
    `
	_, err = Lc.DB.Exec(updateQuery, likesCount, dislikesCount, postID)
	if err != nil {
		return fmt.Errorf("failed to update post votes for post %d: %w", postID, err)
	}

	return nil
}
