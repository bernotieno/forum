package controllers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type PostController struct {
	DB *sql.DB
}

func NewPostController(db *sql.DB) *PostController {
	return &PostController{DB: db}
}

func (pc *PostController) InsertPost(post models.Post) (int, error) {
	// Insert the post with the UserID
	result, err := pc.DB.Exec(`
		INSERT INTO posts (title, user_id, author, category, likes, dislikes, user_vote, content, timestamp, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`, post.Title, post.UserID, post.Author, post.Category, post.Likes, post.Dislikes, post.UserVote, post.Content, post.Timestamp, post.ImageUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to insert post: %w", err)
	}

	// Get the ID of the newly inserted post
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(postID), nil
}

func (pc *PostController) GetAllPosts() ([]models.Post, error) {
	rows, err := pc.DB.Query(`
		SELECT id, title, user_id, author, category, likes, dislikes, 
			   user_vote, content, timestamp, image_url 
		FROM posts 
		ORDER BY timestamp DESC
	`)
	if err != nil {
		logger.Error("Database query failed in GetAllPosts: %v", err)
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
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
			logger.Error("Row scan failed in GetAllPosts: %v", err)
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (pc *PostController) GetPostByID(postID string) (models.Post, error) {
	var post models.Post
	err := pc.DB.QueryRow(`
        SELECT id, title, user_id, author, category, likes, dislikes, 
               user_vote, content, timestamp, image_url 
        FROM posts 
        WHERE id = ?
    `, postID).Scan(
		&post.ID, &post.Title, &post.UserID, &post.Author,
		&post.Category, &post.Likes, &post.Dislikes,
		&post.UserVote, &post.Content, &post.Timestamp, &post.ImageUrl,
	)
	if err != nil {
		return post, fmt.Errorf("failed to fetch post: %w", err)
	}
	return post, nil
}

func (pc *PostController) UpdatePost(post models.Post) error {
	// Prepare the SQL statement for updating the post
	query := `
	UPDATE posts
	SET Title = ?, Author = ?, UserID = ?, Category = ?, Likes = ?, Dislikes = ?, UserVote = ?, Content = ?, ImageUrl = ?, Timestamp = ?
	WHERE ID = ?;
	`

	// Execute the SQL statement with the post data
	_, err := pc.DB.Exec(query,
		post.Title,
		post.Author,
		post.UserID,
		post.Category,
		post.Likes,
		post.Dislikes,
		post.UserVote,
		post.Content,
		post.ImageUrl,
		post.Timestamp,
		post.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeletePost deletes a post from the database by its ID
func (pc *PostController) DeletePost(postID int, userID int) error {
	// Ensure the database connection is not nil
	if pc.DB == nil {
		return errors.New("database connection is nil")
	}

	// Prepare the SQL statement to delete the post
	query := `
	DELETE FROM posts
	WHERE ID = ? AND UserID = ?;
	`

	// Execute the SQL statement
	result, err := pc.DB.Exec(query, postID, userID)
	if err != nil {
		return err
	}

	// Check if the post was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no post found with the given ID or user ID")
	}

	return nil
}
