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
	if post.Title == "" {
		return 0, errors.New("post title is required")
	}
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
	SET title = ?, author = ?, user_id = ?, category = ?, likes = ?, dislikes = ?, user_vote = ?, content = ?, image_url = ?, timestamp = ?
	WHERE id = ?;
	`

	// Execute the SQL statement with the post data
	result, err := pc.DB.Exec(query,
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
		return fmt.Errorf("failed to update post: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no post found with ID %d", post.ID)
	}

	return nil
}

// DeletePost deletes a post from the database by its ID, along with its comments and associated images
func (pc *PostController) DeletePost(postID, userID int) error {
	// Ensure the database connection is not nil
	if pc.DB == nil {
		return errors.New("database connection is nil")
	}

	// Begin a transaction to ensure atomicity
	tx, err := pc.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback in case of error

	// Step 1: Delete all comments associated with the post
	_, err = tx.Exec(`
		DELETE FROM comments 
		WHERE post_id = ?;
	`, postID)
	if err != nil {
		return fmt.Errorf("failed to delete comments: %w", err)
	}

	// Step 2: Fetch image paths associated with the post before deleting the post
	var imagePaths []string
	rows, err := tx.Query(`
		SELECT image_url FROM posts 
		WHERE id = ? AND user_id = ?;
	`, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch image paths: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var imagePath sql.NullString // Use sql.NullString to handle NULL values
		if err := rows.Scan(&imagePath); err != nil {
			return fmt.Errorf("failed to scan image path: %w", err)
		}
		if imagePath.Valid && imagePath.String != "" { // Only append non-empty paths
			imagePaths = append(imagePaths, imagePath.String)
		}
	}

	// Step 3: Delete the post
	result, err := tx.Exec(`
		DELETE FROM posts 
		WHERE id = ? AND user_id = ?;
	`, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Check if the post was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no post found with the given ID or user ID")
	}

	// Step 4: Delete the image files from the upload folder
	err = RemoveImages(imagePaths)
	if err != nil {
		return fmt.Errorf("failed to delete image files: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (pc *PostController) IsPostAuthor(postID, userID int) (bool, error) {
	var authorID int

	err := pc.DB.QueryRow(`
		SELECT user_id 
		FROM posts 
		WHERE id = ?
	`, postID).Scan(&authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("failed to fetch post author: %w", err)
	}

	// Compare the post's author ID with the provided userID
	return authorID == userID, nil
}
