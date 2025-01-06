package database

import (
	"database/sql"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	_ "github.com/mattn/go-sqlite3"
)

func Init() *sql.DB {

	DB, err := sql.Open("sqlite3", "./BackEnd/database/storage/forum.db")
	if err != nil {
		logger.Error("Failed to open database connection: %v", err)
		return nil
	}
	// Create Users and Posts table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE,
			username TEXT UNIQUE,
			password TEXT
		);
	`)
	if err != nil {
		logger.Error("Failed to create users table: %v", err)
		return nil
	}
	return DB
}
