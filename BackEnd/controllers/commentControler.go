package controllers

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type CommentController struct {
	DB *sql.DB
}

func NewCommentController(db *sql.DB) *CommentController {
	return &CommentController{DB: db}
}

func (cCtrl *CommentController) InsertComment(comment models.Comment) (int, error) {
	// Add validation
	if len(comment.Content) == 0 {
		return 0, fmt.Errorf("comment content cannot be empty")
	}
	if len(comment.Content) > 3000 {
		return 0, fmt.Errorf("comment content too long")
	}

	// Check depth limit for replies
	if comment.ParentID.Valid {
		var depth int
		err := cCtrl.DB.QueryRow(`
			WITH RECURSIVE CommentDepth AS (
				SELECT id, parent_id, 0 as depth
				FROM comments
				WHERE id = ?
				
				UNION ALL
				
				SELECT c.id, c.parent_id, cd.depth + 1
				FROM comments c
				INNER JOIN CommentDepth cd ON c.id = cd.parent_id
			)
			SELECT MAX(depth) FROM CommentDepth;
		`, comment.ParentID.Int64).Scan(&depth)

		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to check comment depth: %w", err)
		}

		if depth >= 4 {
			return 0, fmt.Errorf("maximum nesting depth (4) reached")
		}
	}

	result, err := cCtrl.DB.Exec(`
		INSERT INTO comments (post_id, user_id, author, content, likes, dislikes, user_vote, timestamp, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`, comment.PostID, comment.UserID, comment.Author, comment.Content, comment.Likes, comment.Dislikes,
		comment.UserVote, comment.Timestamp, comment.ParentID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(commentID), nil
}

func (cc *CommentController) GetCommentsByPostID(postID string) ([]models.Comment, error) {
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %w", err)
	}
	if postIDInt <= 0 {
		return nil, fmt.Errorf("invalid post ID")
	}

	rows, err := cc.DB.Query(`
        WITH RECURSIVE CommentTree AS (
            -- Base case: get top-level comments
            SELECT 
                id, post_id, user_id, parent_id, author, content, 
                likes, dislikes, user_vote, timestamp,
                0 as depth,
                CAST(id as CHAR(50)) as path
            FROM comments 
            WHERE post_id = ? AND parent_id IS NULL
            
            UNION ALL
            
            -- Recursive case: get replies with depth limit
            SELECT 
                c.id, c.post_id, c.user_id, c.parent_id, c.author, c.content,
                c.likes, c.dislikes, c.user_vote, c.timestamp,
                ct.depth + 1,
                CONCAT(ct.path, ',', c.id)
            FROM comments c
            INNER JOIN CommentTree ct ON c.parent_id = ct.id
            WHERE ct.depth < 6  -- Limit depth to 6 levels
        )
        SELECT * FROM CommentTree
        ORDER BY path, depth, timestamp;
    `, postIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	commentMap := make(map[int]*models.Comment)
	var topLevelComments []*models.Comment

	for rows.Next() {
		var comment models.Comment
		var depth int
		var path string
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.ParentID,
			&comment.Author, &comment.Content, &comment.Likes, &comment.Dislikes,
			&comment.UserVote, &comment.Timestamp, &depth, &path,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		// Initialize Replies slice
		comment.Replies = make([]models.Comment, 0)

		// Store in map
		commentMap[comment.ID] = &comment

		if !comment.ParentID.Valid {
			// This is a top-level comment
			topLevelComments = append(topLevelComments, commentMap[comment.ID])
		} else {
			// This is a reply - add it to its parent's replies
			parentID := int(comment.ParentID.Int64)

			if parent, exists := commentMap[parentID]; exists {
				// Check if parent is a top-level comment
				isTopLevelParent := false
				for _, topComment := range topLevelComments {
					if topComment.ID == parentID {
						isTopLevelParent = true
						break
					}
				}

				if isTopLevelParent {
					// Add reply directly to parent since it's a top-level comment
					parent.Replies = append(parent.Replies, *commentMap[comment.ID])
				} else {
					// Find the top-level ancestor and add the reply there
					for _, topComment := range topLevelComments {
						var addReplyToTopLevel func(*models.Comment) bool
						addReplyToTopLevel = func(c *models.Comment) bool {
							for i := range c.Replies {
								if c.Replies[i].ID == parentID {
									c.Replies[i].Replies = append(c.Replies[i].Replies, *commentMap[comment.ID])
									return true
								}
								if addReplyToTopLevel(&c.Replies[i]) {
									return true
								}
							}
							return false
						}
						if addReplyToTopLevel(topComment) {
							break
						}
					}
				}
			}
		}
	}

	// Convert to slice of values
	result := make([]models.Comment, len(topLevelComments))
	for i, comment := range topLevelComments {
		result[i] = *comment
	}

	return result, nil
}

func (cc *CommentController) GetCommentCountByPostID(postID int) (int, error) {
	var count int
	err := cc.DB.QueryRow(`
        SELECT COUNT(*) 
        FROM comments 
        WHERE post_id = ?
    `, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch comment count: %w", err)
	}
	return count, nil
}

// DeleteComment deletes a comment by its ID
func (cc *CommentController) DeleteComment(commentID int) error {
	// Execute the delete query
	result, err := cc.DB.Exec(`
        DELETE FROM comments 
        WHERE id = ?
    `, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Check if the comment was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no comment found with ID: %d", commentID)
	}

	return nil
}

// IsCommentAuthor checks if the given user is the author of the comment
func (cc *CommentController) IsCommentAuthor(commentID, userID int) (bool, error) {
	var count int
	err := cc.DB.QueryRow(`
        SELECT COUNT(*) 
        FROM comments 
        WHERE id = ? AND user_id = ?
    `, commentID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check comment author: %w", err)
	}

	return count > 0, nil
}

func (cc *CommentController) UpdateComment(commentID int, content string) error {
	result, err := cc.DB.Exec(`
        UPDATE comments 
        SET content = ?
        WHERE id = ?
    `, content, commentID)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no comment found with ID: %d", commentID)
	}

	return nil
}

