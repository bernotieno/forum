package Test

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func TestPostController_InsertPost(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	tests := []struct {
		name    string
		post    models.Post
		wantErr bool
		errMsg  string
	}{
		{
			name: "Insert Valid Post",
			post: models.Post{
				Title:     "Test Post",
				Author:    "user1",
				UserID:    1,
				Category:  "general",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "", Valid: true},
				Content:   "Test content",
				ImageUrl:  sql.NullString{String: "http://example.com/image.jpg", Valid: true},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Insert Post with Missing Title",
			post: models.Post{
				Author:    "user1",
				UserID:    1,
				Category:  "general",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "", Valid: true},
				Content:   "Test content",
				ImageUrl:  sql.NullString{String: "http://example.com/image.jpg", Valid: true},
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "post title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID, err := pc.InsertPost(tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.InsertPost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PostController.InsertPost() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the post was inserted
			if !tt.wantErr {
				var insertedPost models.Post
				err := db.QueryRow(`
					SELECT id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp 
					FROM posts 
					WHERE id = ?`, postID).
					Scan(&insertedPost.ID, &insertedPost.Title, &insertedPost.Author, &insertedPost.UserID, &insertedPost.Category, &insertedPost.Likes, &insertedPost.Dislikes, &insertedPost.UserVote, &insertedPost.Content, &insertedPost.ImageUrl, &insertedPost.Timestamp)
				if err != nil {
					t.Errorf("Failed to verify post insertion: %v", err)
				}
				if insertedPost.Title != tt.post.Title {
					t.Errorf("PostController.InsertPost() title = %v, want %v", insertedPost.Title, tt.post.Title)
				}
			}
		})
	}
}

func TestPostController_GetAllPosts(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 0, 0, '', 'Test content 1', 'http://example.com/image1.jpg', CURRENT_TIMESTAMP),
               (2, 'Test Post 2', 'user2', 2, 'general', 0, 0, '', 'Test content 2', 'http://example.com/image2.jpg', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test posts: %v", err)
	}

	tests := []struct {
		name    string
		want    []models.Post
		wantErr bool
	}{
		{
			name: "Get All Posts",
			want: []models.Post{
				{
					ID:        1,
					Title:     "Test Post 1",
					Author:    "user1",
					UserID:    1,
					Category:  "general",
					Likes:     0,
					Dislikes:  0,
					UserVote:  sql.NullString{String: "", Valid: true},
					Content:   "Test content 1",
					ImageUrl:  sql.NullString{String: "http://example.com/image1.jpg", Valid: true},
					Timestamp: time.Now(),
				},
				{
					ID:        2,
					Title:     "Test Post 2",
					Author:    "user2",
					UserID:    2,
					Category:  "general",
					Likes:     0,
					Dislikes:  0,
					UserVote:  sql.NullString{String: "", Valid: true},
					Content:   "Test content 2",
					ImageUrl:  sql.NullString{String: "http://example.com/image2.jpg", Valid: true},
					Timestamp: time.Now(),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := pc.GetAllPosts()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.GetAllPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(posts) != len(tt.want) {
					t.Errorf("PostController.GetAllPosts() returned %v posts, want %v", len(posts), len(tt.want))
				}
				for i, post := range posts {
					if post.Title != tt.want[i].Title {
						t.Errorf("PostController.GetAllPosts() post[%d].Title = %v, want %v", i, post.Title, tt.want[i].Title)
					}
				}
			}
		})
	}
}

func TestPostController_GetPostByID(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 0, 0, '', 'Test content 1', 'http://example.com/image1.jpg', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		postID  string
		want    models.Post
		wantErr bool
	}{
		{
			name:   "Get Existing Post",
			postID: "1",
			want: models.Post{
				ID:        1,
				Title:     "Test Post 1",
				Author:    "user1",
				UserID:    1,
				Category:  "general",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "", Valid: true},
				Content:   "Test content 1",
				ImageUrl:  sql.NullString{String: "http://example.com/image1.jpg", Valid: true},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "Get Non-Existent Post",
			postID:  "999",
			want:    models.Post{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := pc.GetPostByID(tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.GetPostByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if post.Title != tt.want.Title {
					t.Errorf("PostController.GetPostByID() post.Title = %v, want %v", post.Title, tt.want.Title)
				}
			}
		})
	}
}

func TestPostController_UpdatePost(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 0, 0, '', 'Test content 1', 'http://example.com/image1.jpg', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		post    models.Post
		wantErr bool
	}{
		{
			name: "Update Existing Post",
			post: models.Post{
				ID:        1,
				Title:     "Updated Post",
				Author:    "user1",
				UserID:    1,
				Category:  "general",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "", Valid: true},
				Content:   "Updated content",
				ImageUrl:  sql.NullString{String: "http://example.com/updated_image.jpg", Valid: true},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Update Non-Existent Post",
			post: models.Post{
				ID:        999,
				Title:     "Non-Existent Post",
				Author:    "user1",
				UserID:    1,
				Category:  "general",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "", Valid: true},
				Content:   "Non-existent content",
				ImageUrl:  sql.NullString{String: "http://example.com/non_existent_image.jpg", Valid: true},
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.UpdatePost(tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.UpdatePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var updatedPost models.Post
				err := db.QueryRow(`
					SELECT id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp 
					FROM posts 
					WHERE id = ?`, tt.post.ID).
					Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Author, &updatedPost.UserID, &updatedPost.Category, &updatedPost.Likes, &updatedPost.Dislikes, &updatedPost.UserVote, &updatedPost.Content, &updatedPost.ImageUrl, &updatedPost.Timestamp)
				if err != nil {
					t.Errorf("Failed to verify post update: %v", err)
				}
				if updatedPost.Title != tt.post.Title {
					t.Errorf("PostController.UpdatePost() title = %v, want %v", updatedPost.Title, tt.post.Title)
				}
			}
		})
	}
}

func TestPostController_DeletePost(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 0, 0, '', 'Test content 1', 'http://example.com/image1.jpg', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		userID  int
		wantErr bool
	}{
		{
			name:    "Delete Existing Post",
			postID:  1,
			userID:  1,
			wantErr: false,
		},
		{
			name:    "Delete Non-Existent Post",
			postID:  999,
			userID:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.DeletePost(tt.postID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.DeletePost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", tt.postID).Scan(&count)
				if err != nil {
					t.Errorf("Failed to verify post deletion: %v", err)
				}
				if count != 0 {
					t.Errorf("PostController.DeletePost() post was not deleted")
				}
			}
		})
	}
}

func TestPostController_IsPostAuthor(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a PostController instance
	pc := controllers.NewPostController(db)

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, likes, dislikes, user_vote, content, image_url, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 0, 0, '', 'Test content 1', 'http://example.com/image1.jpg', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		userID  int
		want    bool
		wantErr bool
	}{
		{
			name:    "Check Author for Existing Post",
			postID:  1,
			userID:  1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Check Author for Non-Existent Post",
			postID:  999,
			userID:  1,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Check Author for Wrong User",
			postID:  1,
			userID:  2,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAuthor, err := pc.IsPostAuthor(tt.postID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostController.IsPostAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if isAuthor != tt.want {
				t.Errorf("PostController.IsPostAuthor() = %v, want %v", isAuthor, tt.want)
			}
		})
	}
}
