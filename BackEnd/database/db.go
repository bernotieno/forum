package database

import (
	"database/sql"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
	_ "github.com/mattn/go-sqlite3"
)

var GloabalDB *sql.DB

func Init(env string) (*sql.DB, error) {
	var DB *sql.DB
	var err error
	if env == "Test" {
		DB, err = sql.Open("sqlite3", ":memory:")
		if err != nil {
			logger.Error("Failed to open Test database connection: %v", err)
			return nil, err
		}
	} else {
		DB, err = sql.Open("sqlite3", "./BackEnd/database/storage/forum.db")
		if err != nil {
			logger.Error("Failed to open database connection: %v", err)
			return nil, err
		}
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
		return nil, err
	}

	// Create Posts table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            author TEXT NOT NULL,
            user_id INTEGER NOT NULL,
            category TEXT NOT NULL,
            likes INTEGER DEFAULT 0,
            dislikes INTEGER DEFAULT 0,
            user_vote TEXT,
            content TEXT NOT NULL,
            image_url TEXT,
            timestamp DATETIME NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		logger.Error("Failed to create posts table: %v", err)
		return nil, err
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
            FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
            FOREIGN KEY (parent_id) REFERENCES comments (id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		logger.Error("Failed to create comments table: %v", err)
		return nil, err
	}

	// Create Sessions table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS sessions (
            session_token TEXT PRIMARY KEY,
            user_id INTEGER NOT NULL,
            expires_at DATETIME NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		logger.Error("Failed to create sessions table: %v", err)
		return nil, err
	}

	// Create CSRF Tokens table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS csrf_tokens (
            session_token TEXT NOT NULL,
            csrf_token TEXT NOT NULL,
            expires_at DATETIME NOT NULL,
            PRIMARY KEY (session_token),
            FOREIGN KEY (session_token) REFERENCES sessions (session_token) ON DELETE CASCADE
        );
    `)
	if err != nil {
		logger.Error("Failed to create csrf_tokens table: %v", err)
		return nil, err
	}

	// Create Likes table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS likes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            user_vote TEXT CHECK(user_vote IN ('like', 'dislike')),
            FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		logger.Error("Failed to create likes table: %v", err)
		return nil, err
	}

	// Create Comment Votes table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS comment_votes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            comment_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            vote_type TEXT CHECK(vote_type IN ('like', 'dislike')),
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (comment_id) REFERENCES comments (id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
            UNIQUE(comment_id, user_id)
        );
    `)
	if err != nil {
		logger.Error("Failed to create comment_votes table: %v", err)
		return nil, err
	}

	return DB, nil
}
