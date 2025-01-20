package controllers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// InsertTestComment inserts a test comment into the database.
func InsertTestComment(db *sql.DB, comment models.Comment) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, author, content, likes, dislikes, user_vote, timestamp, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`, comment.PostID, comment.UserID, comment.Author, comment.Content, comment.Likes, comment.Dislikes,
		comment.UserVote, comment.Timestamp, comment.ParentID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func TestCommentController_InsertComment(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Define a base comment for testing
	baseComment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true}, // Updated to sql.NullString
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}

	tests := []struct {
		name          string
		comment       models.Comment
		setup         func(db *sql.DB) // Optional setup function for the test case
		wantCommentID bool
		wantErr       bool
		errMsg        string
	}{
		// Valid Comment
		{
			name:          "Valid Comment",
			comment:       baseComment,
			wantCommentID: true,
			wantErr:       false,
		},
		// Empty Comment Content
		{
			name: "Empty Comment Content",
			comment: models.Comment{
				PostID:    1,
				UserID:    1,
				Author:    "testuser",
				Content:   "", // Empty content
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "upvote", Valid: true}, // Updated to sql.NullString
				Timestamp: time.Now(),
				ParentID:  sql.NullInt64{Valid: false},
			},
			wantCommentID: false,
			wantErr:       true,
			errMsg:        "comment content cannot be empty",
		},
		// Comment Content Too Long
		{
			name: "Comment Content Too Long",
			comment: models.Comment{
				PostID:    1,
				UserID:    1,
				Author:    "testuser",
				Content:   string(make([]byte, 3001)), // Content exceeds 3000 characters
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "upvote", Valid: true}, // Updated to sql.NullString
				Timestamp: time.Now(),
				ParentID:  sql.NullInt64{Valid: false},
			},
			wantCommentID: false,
			wantErr:       true,
			errMsg:        "comment content too long",
		},
		// Nested Comment with Valid Depth
		{
			name: "Nested Comment with Valid Depth",
			comment: func() models.Comment {
				// Insert a parent comment
				parentComment := baseComment
				parentCommentID, err := InsertTestComment(db, parentComment)
				if err != nil {
					t.Fatalf("Failed to insert parent comment: %v", err)
				}

				// Create a nested comment
				nestedComment := baseComment
				nestedComment.ParentID = sql.NullInt64{Int64: parentCommentID, Valid: true}
				return nestedComment
			}(),
			wantCommentID: true,
			wantErr:       false,
		},
		// Nested Comment with Maximum Depth
		{
			name: "Nested Comment with Maximum Depth",
			comment: func() models.Comment {
				// Create a chain of nested comments with depth 4
				var parentCommentID int64
				for i := 0; i < 4; i++ {
					comment := baseComment
					comment.ParentID = sql.NullInt64{Int64: parentCommentID, Valid: true}
					parentCommentID, err = InsertTestComment(db, comment)
					if err != nil {
						t.Fatalf("Failed to insert nested comment: %v", err)
					}
				}

				// Create a nested comment that exceeds the maximum depth
				nestedComment := baseComment
				nestedComment.ParentID = sql.NullInt64{Int64: parentCommentID, Valid: true}
				return nestedComment
			}(),
			wantCommentID: true,
			wantErr:       false,
		},
		// Invalid Parent ID
		{
			name: "Invalid Parent ID",
			comment: models.Comment{
				PostID:    1,
				UserID:    1,
				Author:    "testuser",
				Content:   "This is a test comment.",
				Likes:     0,
				Dislikes:  0,
				UserVote:  sql.NullString{String: "upvote", Valid: true}, // Updated to sql.NullString
				Timestamp: time.Now(),
				ParentID:  sql.NullInt64{Int64: 999, Valid: true}, // Non-existent parent ID
			},
			wantCommentID: false,
			wantErr:       true,
			errMsg:        "failed to check comment depth: sql: Scan error on column index 0, name \"MAX(depth)\": converting NULL to int is unsupported",
		},
		// Database Error (Simulate by closing the database connection)
		{
			name:    "Database Error",
			comment: baseComment,
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			wantCommentID: false,
			wantErr:       true,
			errMsg:        "failed to insert comment: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			commentID, err := cCtrl.InsertComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.InsertComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.InsertComment() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if tt.wantCommentID && commentID <= 0 {
				t.Errorf("CommentController.InsertComment() returned invalid comment ID: %v", commentID)
			}
		})
	}
}

func TestCommentController_GetCommentsByPostID(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Define a base comment for testing
	baseComment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true},
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}

	// Insert a top-level comment
	topLevelCommentID, err := InsertTestComment(db, baseComment)
	if err != nil {
		t.Fatalf("Failed to insert top-level comment: %v", err)
	}

	// Insert a nested comment (reply to the top-level comment)
	nestedComment := baseComment
	nestedComment.ParentID = sql.NullInt64{Int64: topLevelCommentID, Valid: true}
	_, err = InsertTestComment(db, nestedComment)
	if err != nil {
		t.Fatalf("Failed to insert nested comment: %v", err)
	}

	tests := []struct {
		name    string
		postID  string
		setup   func(db *sql.DB) // Optional setup function for the test case
		want    []models.Comment
		wantErr bool
		errMsg  string
	}{
		{
			name:   "Valid Post ID",
			postID: "1",
			want: []models.Comment{
				{
					ID:        int(topLevelCommentID),
					PostID:    1,
					UserID:    1,
					Author:    "testuser",
					Content:   "This is a test comment.",
					Likes:     0,
					Dislikes:  0,
					UserVote:  sql.NullString{String: "upvote", Valid: true},
					Timestamp: baseComment.Timestamp,
					ParentID:  sql.NullInt64{Valid: false},
					Replies: []models.Comment{
						{
							PostID:    1,
							UserID:    1,
							Author:    "testuser",
							Content:   "This is a test comment.",
							Likes:     0,
							Dislikes:  0,
							UserVote:  sql.NullString{String: "upvote", Valid: true},
							Timestamp: nestedComment.Timestamp,
							ParentID:  sql.NullInt64{Int64: topLevelCommentID, Valid: true},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid Post ID - Non-numeric",
			postID:  "invalid",
			want:    nil,
			wantErr: true,
			errMsg:  "invalid post ID: strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name:    "Invalid Post ID - Negative",
			postID:  "-1",
			want:    nil,
			wantErr: true,
			errMsg:  "invalid post ID",
		},
		{
			name:    "No Comments for Post",
			postID:  "2", // Post ID with no comments
			want:    []models.Comment{},
			wantErr: false,
		},
		{
			name:   "Database Error",
			postID: "1",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to fetch comments: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			comments, err := cCtrl.GetCommentsByPostID(tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.GetCommentsByPostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.GetCommentsByPostID() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if len(comments) != len(tt.want) {
				t.Errorf("CommentController.GetCommentsByPostID() returned %d comments, want %d", len(comments), len(tt.want))
				return
			}
			// Compare each comment and its replies
			for i, comment := range comments {
				if comment.ID != tt.want[i].ID ||
					comment.PostID != tt.want[i].PostID ||
					comment.UserID != tt.want[i].UserID ||
					comment.Author != tt.want[i].Author ||
					comment.Content != tt.want[i].Content ||
					comment.Likes != tt.want[i].Likes ||
					comment.Dislikes != tt.want[i].Dislikes ||
					comment.UserVote != tt.want[i].UserVote ||
					!comment.Timestamp.Equal(tt.want[i].Timestamp) ||
					comment.ParentID != tt.want[i].ParentID {
					t.Errorf("CommentController.GetCommentsByPostID() comment mismatch: got %v, want %v", comment, tt.want[i])
				}
				// Compare replies
				if len(comment.Replies) != len(tt.want[i].Replies) {
					t.Errorf("CommentController.GetCommentsByPostID() reply count mismatch: got %d, want %d", len(comment.Replies), len(tt.want[i].Replies))
				}
			}
		})
	}
}

func TestCommentController_GetCommentCountByPostID(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Insert test comments
	comment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true},
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}
	_, err = InsertTestComment(db, comment)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name    string
		postID  int
		want    int
		wantErr bool
		errMsg  string
		setup   func(db *sql.DB)
	}{
		{
			name:    "Valid Post ID",
			postID:  1,
			want:    1,
			wantErr: false,
		},
		{
			name:    "No Comments for Post",
			postID:  2, // Post ID with no comments
			want:    0,
			wantErr: false,
		},
		{
			name:   "Database Error",
			postID: 1,
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			want:    0,
			wantErr: true,
			errMsg:  "failed to fetch comment count: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			count, err := cCtrl.GetCommentCountByPostID(tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.GetCommentCountByPostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.GetCommentCountByPostID() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if count != tt.want {
				t.Errorf("CommentController.GetCommentCountByPostID() = %v, want %v", count, tt.want)
			}
		})
	}
}

func TestCommentController_DeleteComment(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Insert a test comment
	comment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true},
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}
	commentID, err := InsertTestComment(db, comment)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		setup     func(db *sql.DB) // Optional setup function for the test case
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid Comment ID",
			commentID: int(commentID),
			wantErr:   false,
		},
		{
			name:      "Invalid Comment ID",
			commentID: 999, // Non-existent comment ID
			wantErr:   true,
			errMsg:    "no comment found with ID: 999",
		},
		{
			name:      "Database Error",
			commentID: int(commentID),
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			wantErr: true,
			errMsg:  "failed to delete comment: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := cCtrl.DeleteComment(tt.commentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.DeleteComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.DeleteComment() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCommentController_IsCommentAuthor(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Insert a test comment
	comment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true},
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}
	commentID, err := InsertTestComment(db, comment)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		userID    int
		want      bool
		wantErr   bool
		errMsg    string
		setup     func(db *sql.DB)
	}{
		{
			name:      "Valid Author",
			commentID: int(commentID),
			userID:    1,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "Invalid Author",
			commentID: int(commentID),
			userID:    2, // Different user ID
			want:      false,
			wantErr:   false,
		},
		{
			name:      "Invalid Comment ID",
			commentID: 999, // Non-existent comment ID
			userID:    1,
			want:      false,
			wantErr:   false,
		},
		{
			name:      "Database Error",
			commentID: int(commentID),
			userID:    1,
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			want:    false,
			wantErr: true,
			errMsg:  "failed to check comment author: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			isAuthor, err := cCtrl.IsCommentAuthor(tt.commentID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.IsCommentAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.IsCommentAuthor() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if isAuthor != tt.want {
				t.Errorf("CommentController.IsCommentAuthor() = %v, want %v", isAuthor, tt.want)
			}
		})
	}
}

func TestCommentController_UpdateComment(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a CommentController instance
	cCtrl := NewCommentController(db)

	// Insert a test comment
	comment := models.Comment{
		PostID:    1,
		UserID:    1,
		Author:    "testuser",
		Content:   "This is a test comment.",
		Likes:     0,
		Dislikes:  0,
		UserVote:  sql.NullString{String: "upvote", Valid: true},
		Timestamp: time.Now(),
		ParentID:  sql.NullInt64{Valid: false},
	}
	commentID, err := InsertTestComment(db, comment)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name      string
		commentID int
		content   string
		setup     func(db *sql.DB) // Optional setup function for the test case
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid Update",
			commentID: int(commentID),
			content:   "Updated comment content.",
			wantErr:   false,
		},
		{
			name:      "Invalid Comment ID",
			commentID: 999, // Non-existent comment ID
			content:   "Updated comment content.",
			wantErr:   true,
			errMsg:    "no comment found with ID: 999",
		},
		{
			name:      "Database Error",
			commentID: int(commentID),
			content:   "Updated comment content.",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
			wantErr: true,
			errMsg:  "failed to update comment: sql: database is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := cCtrl.UpdateComment(tt.commentID, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommentController.UpdateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errMsg {
					t.Errorf("CommentController.UpdateComment() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
			}
		})
	}
}
