package database

import (
	"database/sql"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	_ "github.com/mattn/go-sqlite3"
)

var GloabalDB *sql.DB

func Init() *sql.DB {
	DB, err := sql.Open("sqlite3", "./BackEnd/database/storage/forum.db")
	if err != nil {
		logger.Error("Failed to open database connection: %v", err)
		return nil
	}
	GloabalDB = DB
	// Create Users table
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
	// Create Posts table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			user_id INTEGER NOT NULL, -- Added user_id column
			category TEXT NOT NULL,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			user_vote TEXT,
			content TEXT NOT NULL,
			image_url TEXT,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id) -- Added comma before FOREIGN KEY
		);
	`)
	if err != nil {
		logger.Error("Failed to create posts table: %v", err)
		return nil
	}
	// Create Comments table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			parent_id INTEGER DEFAULT NULL,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			user_vote TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (parent_id) REFERENCES comments(id)
		);
	`)
	if err != nil {
		logger.Error("Failed to create comments table: %v", err)
		return nil
	}
	// create sessions table
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		session_token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL
	);
   `)
	if err != nil {
		logger.Error("Failed to create sessions table: %v", err)
		return nil
	}
	// create csrf_tokens table
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS csrf_tokens (
		session_token TEXT NOT NULL, 
		csrf_token TEXT NOT NULL,   
		expires_at DATETIME NOT NULL,
		PRIMARY KEY (session_token)
   );
  `)
	if err != nil {
		logger.Error("Failed to create csrf_tokens table: %v", err)
		return nil
	}
	return DB
}
