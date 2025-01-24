package tests

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create test uploads directory
	err := os.MkdirAll("uploads", 0755)
	if err != nil {
		t.Fatalf("Failed to create uploads directory: %v", err)
	}

	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func teardownTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}

	// Clean up test directories
	dirsToClean := []string{
		"uploads",
		"logs",
		"./BackEnd/database/storage",
	}

	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Failed to remove directory %s: %v", dir, err)
		}
	}
}

func clearDatabaseTables(db *sql.DB) error {
	tables := []string{
		"users",
		"posts",
		"comments",
		"likes",
		"comment_votes",
		"sessions",
	}

	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
