package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func Init() *sql.DB {
	var err error
	DB, err := sql.Open("sqlite3", "./BackEnd/database/storage/forum.db") 
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	return DB
}
