package controllers

import (
	"database/sql"
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