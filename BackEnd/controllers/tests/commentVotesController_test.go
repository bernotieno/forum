package Test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func init() {
	// Initialize the logger for tests
	logger.Init()
}

func TestCommentVotesController_UpdateCommentVotes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 0, 0)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	// Insert test data into comment_votes table
	_, err = db.Exec(`
        INSERT INTO comment_votes (comment_id, user_id, vote_type) 
        VALUES (1, 1, 'like'), (1, 2, 'dislike')
    `)
	if err != nil {
		t.Fatalf("Failed to insert test votes: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "Valid Comment ID",
			commentID: 1,
			wantErr:   false,
		},
		{
			name:      "Invalid Comment ID",
			commentID: 999, // Non-existent comment ID
			wantErr:   true,
			errMsg:    "comment with ID 999 does not exist",
			setup: func(db *sql.DB) {
			},
		},
		{
			name:      "Database Error",
			commentID: 1,
			wantErr:   true,
			errMsg:    "failed to check if comment exists: sql: database is closed",
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

			err := cc.UpdateCommentVotes(tt.commentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.UpdateCommentVotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.UpdateCommentVotes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the updated vote counts
			var likes, dislikes int
			err = db.QueryRow(`SELECT likes, dislikes FROM comments WHERE id = ?`, tt.commentID).Scan(&likes, &dislikes)
			if err != nil {
				t.Errorf("Failed to verify updated vote counts: %v", err)
			}
			if likes != 1 || dislikes != 1 {
				t.Errorf("CommentVotesController.UpdateCommentVotes() likes = %v, dislikes = %v, want likes = 1, dislikes = 1", likes, dislikes)
			}
		})
	}
}

func TestCommentVotesController_GetCommentVotes(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 5, 3)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name         string
		commentID    int
		wantLikes    int
		wantDislikes int
		wantErr      bool
		errMsg       string
		setup        func(db *sql.DB)
	}{
		{
			name:         "Valid Comment ID",
			commentID:    1,
			wantLikes:    5,
			wantDislikes: 3,
			wantErr:      false,
		},
		{
			name:         "Invalid Comment ID",
			commentID:    999, // Non-existent comment ID
			wantLikes:    0,
			wantDislikes: 0,
			wantErr:      true,
			errMsg:       "sql: no rows in result set",
		},
		{
			name:         "Database Error",
			commentID:    1,
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

			likes, dislikes, err := cc.GetCommentVotes(tt.commentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.GetCommentVotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.GetCommentVotes() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if likes != tt.wantLikes || dislikes != tt.wantDislikes {
				t.Errorf("CommentVotesController.GetCommentVotes() likes = %v, dislikes = %v, want likes = %v, dislikes = %v", likes, dislikes, tt.wantLikes, tt.wantDislikes)
			}
		})
	}
}

func TestCommentVotesController_HandleCommentVote(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 0, 0)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		userID    int
		voteType  string
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "New Like Vote",
			commentID: 1,
			userID:    1,
			voteType:  "like",
			wantErr:   false,
		},
		{
			name:      "Change Vote to Dislike",
			commentID: 1,
			userID:    1,
			voteType:  "dislike",
			wantErr:   false,
			setup: func(db *sql.DB) {
				// Insert an existing vote
				_, _ = db.Exec(`INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (1, 1, 'like')`)
			},
		},
		{
			name:      "Remove Existing Vote",
			commentID: 1,
			userID:    1,
			voteType:  "like",
			wantErr:   false,
			setup: func(db *sql.DB) {
				// Insert an existing vote
				_, _ = db.Exec(`INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (1, 1, 'like')`)
			},
		},
		{
			name:      "Invalid Vote Type",
			commentID: 1,
			userID:    1,
			voteType:  "invalid",
			wantErr:   true,
			errMsg:    "CHECK constraint failed: vote_type IN ('like', 'dislike')",
		},
		{
			name:      "Database Error",
			commentID: 1,
			userID:    1,
			voteType:  "like",
			wantErr:   true,
			errMsg:    "sql: database is closed",
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

			err := cc.HandleCommentVote(tt.commentID, tt.userID, tt.voteType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.HandleCommentVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.HandleCommentVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
		})
	}
}

func TestCommentVotesController_CheckUserVote(t *testing.T) {
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

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 0, 0)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	// Insert test data into comment_votes table
	_, err = db.Exec(`
        INSERT INTO comment_votes (comment_id, user_id, vote_type) 
        VALUES (1, 1, 'like'), (1, 2, 'dislike')
    `)
	if err != nil {
		t.Fatalf("Failed to insert test votes: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		userID    int
		wantVote  string
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "Existing Vote",
			commentID: 1,
			userID:    1,
			wantVote:  "like",
			wantErr:   false,
		},
		{
			name:      "No Vote Exists",
			commentID: 1,
			userID:    3, // User with no vote
			wantVote:  "",
			wantErr:   false,
		},
		{
			name:      "Database Error",
			commentID: 1,
			userID:    1,
			wantVote:  "",
			wantErr:   true,
			errMsg:    "sql: database is closed",
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

			voteType, err := cc.CheckUserVote(tt.commentID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.CheckUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.CheckUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if voteType != tt.wantVote {
				t.Errorf("CommentVotesController.CheckUserVote() voteType = %v, wantVote %v", voteType, tt.wantVote)
			}
		})
	}
}

func TestCommentVotesController_RemoveUserVote(t *testing.T) {
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

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 0, 0)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	// Insert test data into comment_votes table
	_, err = db.Exec(`
        INSERT INTO comment_votes (comment_id, user_id, vote_type) 
        VALUES (1, 1, 'like'), (1, 2, 'dislike')
    `)
	if err != nil {
		t.Fatalf("Failed to insert test votes: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		userID    int
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "Existing Vote",
			commentID: 1,
			userID:    1,
			wantErr:   false,
		},
		{
			name:      "No Vote Exists",
			commentID: 1,
			userID:    3, // User with no vote
			wantErr:   false,
		},
		{
			name:      "Database Error",
			commentID: 1,
			userID:    1,
			wantErr:   true,
			errMsg:    "sql: database is closed",
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

			err := cc.RemoveUserVote(tt.commentID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.RemoveUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.RemoveUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the vote was removed
			if !tt.wantErr {
				voteType, err := cc.CheckUserVote(tt.commentID, tt.userID)
				if err != nil {
					t.Errorf("Failed to verify vote removal: %v", err)
				}
				if voteType != "" {
					t.Errorf("CommentVotesController.RemoveUserVote() voteType = %v, want empty string", voteType)
				}
			}
		})
	}
}

func TestCommentVotesController_AddUserVote(t *testing.T) {
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

	// Create a CommentVotesController instance
	cc := controllers.NewCommentVotesController(db)

	// Insert test data into comments table
	_, err = db.Exec(`
        INSERT INTO comments (id, post_id, user_id, author, content, likes, dislikes) 
        VALUES (1, 1, 1, 'testuser', 'Test comment', 0, 0)
    `)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		userID    int
		voteType  string
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "New Vote",
			commentID: 1,
			userID:    3,
			voteType:  "like",
			wantErr:   false,
		},
		{
			name:      "Invalid Vote Type",
			commentID: 1,
			userID:    3,
			voteType:  "invalid",
			wantErr:   true,
			errMsg:    "CHECK constraint failed: vote_type IN ('like', 'dislike')",
		},
		{
			name:      "Database Error",
			commentID: 1,
			userID:    3,
			voteType:  "like",
			wantErr:   true,
			errMsg:    "sql: database is closed",
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

			err := cc.AddUserVote(tt.commentID, tt.userID, tt.voteType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentVotesController.AddUserVote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CommentVotesController.AddUserVote() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the vote was added
			if !tt.wantErr {
				voteType, err := cc.CheckUserVote(tt.commentID, tt.userID)
				if err != nil {
					t.Errorf("Failed to verify vote addition: %v", err)
				}
				if voteType != tt.voteType {
					t.Errorf("CommentVotesController.AddUserVote() voteType = %v, want %v", voteType, tt.voteType)
				}
			}
		})
	}
}
