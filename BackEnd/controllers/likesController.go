package controllers

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type LikesController struct {
	DB *sql.DB
}

func NewLikesController(db *sql.DB) *LikesController {
	return &LikesController{DB: db}
}

func (Lc *LikesController) InsertLikes(like models.Likes) error {
	// Check if a row with the same post_id and user_id exists
	existingRow := Lc.DB.QueryRow(`
        SELECT id FROM likes WHERE post_id = ? AND user_id = ?;
    `, like.PostId, like.UserId)

	var existingID int
	err := existingRow.Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing like: %w", err)
	}

	// If a matching row exists, delete it
	if err == nil {
		_, err := Lc.DB.Exec(`DELETE FROM likes WHERE id = ?;`, existingID)
		if err != nil {
			return fmt.Errorf("failed to delete existing like: %w", err)
		}
	}

	// Insert the new like
	_, err = Lc.DB.Exec(`
        INSERT INTO likes (post_id, user_id, user_vote)
        VALUES (?, ?, ?);
    `, like.PostId, like.UserId, like.UserVote)
	if err != nil {
		return fmt.Errorf("failed to insert like: %w", err)
	}

	return nil
}

func (lc *LikesController) UpdatePostVotes(postID int) error {
	// Query to calculate the likes and dislikes from the likes table
	query := `
        SELECT 
            COALESCE(SUM(CASE WHEN user_vote = 'like' THEN 1 ELSE 0 END), 0) AS likes_count,
            COALESCE(SUM(CASE WHEN user_vote = 'dislike' THEN 1 ELSE 0 END), 0) AS dislikes_count
        FROM likes
        WHERE post_id = ?;
    `

	var likesCount, dislikesCount int
	err := lc.DB.QueryRow(query, postID).Scan(&likesCount, &dislikesCount)
	if err != nil {
		return fmt.Errorf("failed to retrieve vote counts for post %d: %w", postID, err)
	}

	// Update the likes and dislikes in the posts table
	updateQuery := `
        UPDATE posts
        SET likes = ?, dislikes = ?
        WHERE id = ?;
    `
	_, err = lc.DB.Exec(updateQuery, likesCount, dislikesCount, postID)
	if err != nil {
		return fmt.Errorf("failed to update post votes for post %d: %w", postID, err)
	}

	return nil
}

func (lc *LikesController) GetPostVotes(postID int) (int, int, error) {
	// Query to calculate the likes and dislikes from the posts table
	query := `
		SELECT Likes, Dislikes FROM posts WHERE id = ?;
	`

	var likesCount, dislikesCount sql.NullInt64
	err := lc.DB.QueryRow(query, postID).Scan(&likesCount, &dislikesCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to retrieve vote counts for post %d: %w", postID, err)
	}

	// Convert sql.NullInt64 to int, using 0 if the value is NULL
	return int(likesCount.Int64), int(dislikesCount.Int64), nil
}

// CheckUserVote checks if the user has already voted on a post
func (lc *LikesController) CheckUserVote(postID, userID int) (string, error) {
	var userVote string
	query := `SELECT user_vote FROM likes WHERE post_id = ? AND user_id = ?`
	err := lc.DB.QueryRow(query, postID, userID).Scan(&userVote)
	if err != nil {
		if err == sql.ErrNoRows {
			// No vote found
			return "", nil
		}
		return "", err
	}
	return userVote, nil
}

// RemoveUserVote removes the user's vote from the database
func (lc *LikesController) RemoveUserVote(postID, userID int) error {
	query := `DELETE FROM likes WHERE post_id = ? AND user_id = ?`
	_, err := lc.DB.Exec(query, postID, userID)
	if err != nil {
		return err
	}
	return nil
}

// AddUserVote adds the user's vote to the database
func (lc *LikesController) AddUserVote(postID, userID int, vote string) error {
	query := `INSERT INTO likes (post_id, user_id, user_vote) VALUES (?, ?, ?)`
	_, err := lc.DB.Exec(query, postID, userID, vote)
	if err != nil {
		return err
	}
	return nil
}

// HandleVote handles the user's vote (like or dislike)
func (lc *LikesController) HandleVote(postID, userID int, vote string) error {
	// Check if the user has already voted
	existingVote, err := lc.CheckUserVote(postID, userID)
	if err != nil {
		return fmt.Errorf("failed to check user vote: %v", err)
	}

	// If the user has already voted in the same way, remove their vote
	if existingVote == vote {
		err = lc.RemoveUserVote(postID, userID)
		if err != nil {
			return fmt.Errorf("failed to remove user vote: %v", err)
		}
	} else if existingVote != "" {
		// If the user has voted in the opposite way, remove their previous vote and add the new one
		err = lc.RemoveUserVote(postID, userID)
		if err != nil {
			return fmt.Errorf("failed to remove user vote: %v", err)
		}
		err = lc.AddUserVote(postID, userID, vote)
		if err != nil {
			return fmt.Errorf("failed to add user vote: %v", err)
		}
	} else {
		// If the user has not voted, add their vote
		err = lc.AddUserVote(postID, userID, vote)
		if err != nil {
			return fmt.Errorf("failed to add user vote: %v", err)
		}
	}

	// Update the post's likes and dislikes count
	err = lc.UpdatePostVotes(postID)
	if err != nil {
		return fmt.Errorf("failed to update post votes: %v", err)
	}

	return nil
}

func (lc *LikesController) GetUserVotes(userID int) (map[string]string, error) {
	// Query the database to get the user's votes for all posts
	query := `
	SELECT post_id, user_vote
	FROM likes
	WHERE user_id = ?;
`
	rows, err := lc.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user votes: %v", err)
	}
	defer rows.Close()

	// Create a map to store the user's votes
	userVotes := make(map[string]string)
	for rows.Next() {
		var postID int
		var userVote string
		if err := rows.Scan(&postID, &userVote); err != nil {
			return nil, fmt.Errorf("failed to scan user votes: %v", err)
		}
		userVotes[strconv.Itoa(postID)] = userVote
	}

	return userVotes, nil
}
