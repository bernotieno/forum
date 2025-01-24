package tests

import (
	"database/sql"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/database"
)

func setupTestDB(t *testing.T) *sql.DB {
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
