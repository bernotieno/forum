package Test

import (
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// Modify TestMain to handle setup and cleanup
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup after all tests are done
	cleanupTestResources()

	os.Exit(code)
}

func TestLikesController_InsertLikes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		like    models.Likes
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name: "Insert New Like",
			like: models.Likes{
				PostId:   1,
				UserId:   1,
				UserVote: "like",
			},
			wantErr: false,
		},
		{
			name: "Insert New Dislike",
			like: models.Likes{
				PostId:   1,
				UserId:   2,
				UserVote: "dislike",
			},
			wantErr: false,
		},
		{
			name: "Invalid User Vote",
			like: models.Likes{
				PostId:   1,
				UserId:   1,
				UserVote: "invalid", // Invalid vote type
			},
			wantErr: true,
			errMsg:  "CHECK constraint failed", // SQLite error for invalid user_vote
		},
		{
			name: "Database Error",
			like: models.Likes{
				PostId:   1,
				UserId:   1,
				UserVote: "like",
			},
			wantErr: true,
			errMsg:  "failed to check existing like: sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := lc.InsertLikes(tt.like)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.InsertLikes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.InsertLikes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the like was inserted or updated
			if !tt.wantErr {
				var userVote string
				err := db.QueryRow("SELECT user_vote FROM likes WHERE post_id = ? AND user_id = ?", tt.like.PostId, tt.like.UserId).
					Scan(&userVote)
				if err != nil {
					t.Errorf("Failed to verify like insertion: %v", err)
				}
				if userVote != tt.like.UserVote {
					t.Errorf("LikesController.InsertLikes() userVote = %v, want %v", userVote, tt.like.UserVote)
				}
			}
		})
	}
}

func TestLikesController_UpdatePostVotes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test data into the likes table
	_, err = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like'), (1, 2, 'dislike')")
	if err != nil {
		t.Fatalf("Failed to insert test likes: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name:    "Valid Post ID",
			postID:  1,
			wantErr: false,
		},
		{
			name:    "Invalid Post ID",
			postID:  999,   // Non-existent post ID
			wantErr: false, // No error, but no rows will be updated
		},
		{
			name:    "Database Error",
			postID:  1,
			wantErr: true,
			errMsg:  "failed to retrieve vote counts for post 1: sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := lc.UpdatePostVotes(tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.UpdatePostVotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.UpdatePostVotes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the post votes were updated
			if !tt.wantErr {
				var likes, dislikes int
				err := db.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", tt.postID).
					Scan(&likes, &dislikes)
				if err != nil {
					if err == sql.ErrNoRows {
						// Non-existent post ID, so likes and dislikes should be 0
						likes, dislikes = 0, 0
					} else {
						t.Errorf("Failed to verify post votes update: %v", err)
						return
					}
				}
				if tt.postID == 1 {
					// For valid post ID, expect likes = 1 and dislikes = 1
					if likes != 1 || dislikes != 1 {
						t.Errorf("LikesController.UpdatePostVotes() likes = %v, dislikes = %v, want likes = 1, dislikes = 1", likes, dislikes)
					}
				} else {
					// For invalid post ID, expect likes = 0 and dislikes = 0
					if likes != 0 || dislikes != 0 {
						t.Errorf("LikesController.UpdatePostVotes() likes = %v, dislikes = %v, want likes = 0, dislikes = 0", likes, dislikes)
					}
				}
			}
		})
	}
}

func TestLikesController_GetPostVotes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, likes, dislikes, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', 5, 3, CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name         string
		postID       int
		wantLikes    int
		wantDislikes int
		wantErr      bool
		errMsg       string
		setup        func(db *sql.DB)
	}{
		{
			name:         "Valid Post ID",
			postID:       1,
			wantLikes:    5,
			wantDislikes: 3,
			wantErr:      false,
		},
		{
			name:         "Invalid Post ID",
			postID:       999, // Non-existent post ID
			wantLikes:    0,
			wantDislikes: 0,
			wantErr:      true,
			errMsg:       "failed to retrieve vote counts",
		},
		{
			name:         "Database Error",
			postID:       1,
			wantLikes:    0,
			wantDislikes: 0,
			wantErr:      true,
			errMsg:       "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			likes, dislikes, err := lc.GetPostVotes(tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.GetPostVotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.GetPostVotes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if likes != tt.wantLikes || dislikes != tt.wantDislikes {
				t.Errorf("LikesController.GetPostVotes() likes = %v, dislikes = %v, want likes = %v, dislikes = %v", likes, dislikes, tt.wantLikes, tt.wantDislikes)
			}
		})
	}
}

func TestLikesController_CheckUserVote(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test data into the likes table
	_, err = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like')")
	if err != nil {
		t.Fatalf("Failed to insert test like: %v", err)
	}

	tests := []struct {
		name     string
		postID   int
		userID   int
		wantVote string
		wantErr  bool
		errMsg   string
		setup    func(db *sql.DB)
	}{
		{
			name:     "Existing Vote",
			postID:   1,
			userID:   1,
			wantVote: "like",
			wantErr:  false,
		},
		{
			name:     "No Vote Exists",
			postID:   1,
			userID:   2, // User with no vote
			wantVote: "",
			wantErr:  false,
		},
		{
			name:     "Database Error",
			postID:   1,
			userID:   1,
			wantVote: "",
			wantErr:  true,
			errMsg:   "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			vote, err := lc.CheckUserVote(tt.postID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.CheckUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.CheckUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if vote != tt.wantVote {
				t.Errorf("LikesController.CheckUserVote() vote = %v, wantVote %v", vote, tt.wantVote)
			}
		})
	}
}

func TestLikesController_RemoveUserVote(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test data into the likes table
	_, err = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like')")
	if err != nil {
		t.Fatalf("Failed to insert test like: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		userID  int
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name:    "Valid Vote Removal",
			postID:  1,
			userID:  1,
			wantErr: false,
		},
		{
			name:    "Non-Existent Vote",
			postID:  1,
			userID:  2,     // User with no vote
			wantErr: false, // No error, but no rows will be deleted
		},
		{
			name:    "Database Error",
			postID:  1,
			userID:  1,
			wantErr: true,
			errMsg:  "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := lc.RemoveUserVote(tt.postID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.RemoveUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.RemoveUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the vote was removed
			if !tt.wantErr {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND user_id = ?", tt.postID, tt.userID).
					Scan(&count)
				if err != nil {
					t.Errorf("Failed to verify vote removal: %v", err)
				}
				if count != 0 {
					t.Errorf("LikesController.RemoveUserVote() count = %v, want 0", count)
				}
			}
		})
	}
}

func TestLikesController_AddUserVote(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		userID  int
		vote    string
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name:    "Valid Like Vote",
			postID:  1,
			userID:  1,
			vote:    "like",
			wantErr: false,
		},
		{
			name:    "Valid Dislike Vote",
			postID:  1,
			userID:  1,
			vote:    "dislike",
			wantErr: false,
			setup: func(db *sql.DB) {
				// Ensure the likes table is empty before the test
				_, _ = db.Exec("DELETE FROM likes WHERE post_id = 1 AND user_id = 1")
			},
		},
		{
			name:    "Invalid Vote Type",
			postID:  1,
			userID:  1,
			vote:    "invalid", // Invalid vote type
			wantErr: true,
			errMsg:  "CHECK constraint failed", // SQLite error for invalid user_vote
		},
		{
			name:    "Database Error",
			postID:  1,
			userID:  1,
			vote:    "like",
			wantErr: true,
			errMsg:  "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := lc.AddUserVote(tt.postID, tt.userID, tt.vote)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.AddUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.AddUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the vote was added
			if !tt.wantErr {
				var userVote string
				err := db.QueryRow("SELECT user_vote FROM likes WHERE post_id = ? AND user_id = ?", tt.postID, tt.userID).
					Scan(&userVote)
				if err != nil {
					t.Errorf("Failed to verify vote addition: %v", err)
				}
				if userVote != tt.vote {
					t.Errorf("LikesController.AddUserVote() userVote = %v, want %v", userVote, tt.vote)
				}
			}
		})
	}
}

func TestLikesController_HandleVote(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		userID  int
		vote    string
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name:    "New Like Vote",
			postID:  1,
			userID:  1,
			vote:    "like",
			wantErr: false,
		},
		{
			name:    "Change Like to Dislike",
			postID:  1,
			userID:  1,
			vote:    "dislike",
			wantErr: false,
			setup: func(db *sql.DB) {
				// Insert an existing like
				_, _ = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like')")
			},
		},
		{
			name:    "Remove Existing Vote",
			postID:  1,
			userID:  1,
			vote:    "like",
			wantErr: false,
			setup: func(db *sql.DB) {
				// Insert an existing like
				_, _ = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like')")
			},
		},
		{
			name:    "Database Error",
			postID:  1,
			userID:  1,
			vote:    "like",
			wantErr: true,
			errMsg:  "failed to check user vote: sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := lc.HandleVote(tt.postID, tt.userID, tt.vote)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.HandleVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.HandleVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the vote was handled correctly
			if !tt.wantErr {
				var userVote string
				err := db.QueryRow("SELECT user_vote FROM likes WHERE post_id = ? AND user_id = ?", tt.postID, tt.userID).
					Scan(&userVote)
				if err != nil && err != sql.ErrNoRows {
					t.Errorf("Failed to verify vote handling: %v", err)
				}
				if tt.vote == "like" && userVote != "like" {
					t.Errorf("LikesController.HandleVote() userVote = %v, want like", userVote)
				}
				if tt.vote == "dislike" && userVote != "dislike" {
					t.Errorf("LikesController.HandleVote() userVote = %v, want dislike", userVote)
				}
			}
		})
	}
}

func TestLikesController_GetUserVotes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post', 'user1', 1, 'general', 'Test content', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test data into the likes table
	_, err = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like')")
	if err != nil {
		t.Fatalf("Failed to insert test like: %v", err)
	}

	tests := []struct {
		name      string
		userID    int
		wantVotes map[string]string
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:   "Valid User ID",
			userID: 1,
			wantVotes: map[string]string{
				"1": "like",
			},
			wantErr: false,
		},
		{
			name:      "No Votes",
			userID:    2, // User with no votes
			wantVotes: map[string]string{},
			wantErr:   false,
		},
		{
			name:      "Database Error",
			userID:    1,
			wantVotes: nil,
			wantErr:   true,
			errMsg:    "failed to query user votes",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			votes, err := lc.GetUserVotes(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.GetUserVotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.GetUserVotes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if !reflect.DeepEqual(votes, tt.wantVotes) {
				t.Errorf("LikesController.GetUserVotes() votes = %v, want %v", votes, tt.wantVotes)
			}
		})
	}
}

func TestLikesController_GetUserLikesPosts(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Clear database tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a LikesController instance
	lc := controllers.NewLikesController(db)

	// Insert test data into the users table
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Insert test data into the posts table
	_, err = db.Exec(`
        INSERT INTO posts (id, title, author, user_id, category, content, timestamp)
        VALUES (1, 'Test Post 1', 'user1', 1, 'general', 'Test content 1', CURRENT_TIMESTAMP),
               (2, 'Test Post 2', 'user1', 1, 'general', 'Test content 2', CURRENT_TIMESTAMP)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test posts: %v", err)
	}

	// Insert test data into the likes table
	_, err = db.Exec("INSERT INTO likes (post_id, user_id, user_vote) VALUES (1, 1, 'like'), (2, 1, 'like')")
	if err != nil {
		t.Fatalf("Failed to insert test likes: %v", err)
	}

	tests := []struct {
		name      string
		userID    int
		wantPosts []models.Post
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:   "Valid User ID",
			userID: 1,
			wantPosts: []models.Post{
				{
					ID:        1,
					Title:     "Test Post 1",
					Author:    "user1",
					UserID:    1,
					Category:  "general",
					Content:   "Test content 1",
					Timestamp: time.Now(), // Adjust based on actual timestamp
				},
				{
					ID:        2,
					Title:     "Test Post 2",
					Author:    "user1",
					UserID:    1,
					Category:  "general",
					Content:   "Test content 2",
					Timestamp: time.Now(), // Adjust based on actual timestamp
				},
			},
			wantErr: false,
		},
		{
			name:      "No Liked Posts",
			userID:    2, // User with no liked posts
			wantPosts: []models.Post{},
			wantErr:   false,
		},
		{
			name:      "Database Error",
			userID:    1,
			wantPosts: nil,
			wantErr:   true,
			errMsg:    "failed to query user likes",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			posts, err := lc.GetUserLikesPosts(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LikesController.GetUserLikesPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("LikesController.GetUserLikesPosts() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the posts
			if len(posts) != len(tt.wantPosts) {
				t.Errorf("LikesController.GetUserLikesPosts() posts count = %v, want %v", len(posts), len(tt.wantPosts))
				return
			}

			for i, post := range posts {
				if post.ID != tt.wantPosts[i].ID ||
					post.Title != tt.wantPosts[i].Title ||
					post.Author != tt.wantPosts[i].Author ||
					post.UserID != tt.wantPosts[i].UserID ||
					post.Category != tt.wantPosts[i].Category ||
					post.Content != tt.wantPosts[i].Content {
					t.Errorf("LikesController.GetUserLikesPosts() post = %+v, want %+v", post, tt.wantPosts[i])
				}
			}
		})
	}
}
