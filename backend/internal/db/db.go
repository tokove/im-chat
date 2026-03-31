package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath, dbName string) {
	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database dir, err: %v", err)
	}

	dbFile := filepath.Join(dbPath, dbName)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database, err: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database, err: %v", err)
	}

	// db.SetMaxOpenConns(30)
	// db.SetMaxIdleConns(5)
	// db.SetConnMaxLifetime(0)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			log.Printf("Failed to execute %s, err: %v", p, err)
		}
	}

	tables := []string{
		// Users
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			refresh_token_web TEXT,
  			refresh_token_web_at DATETIME,
  			refresh_token_mobile TEXT,
  			refresh_token_mobile_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Privates
		`CREATE TABLE IF NOT EXISTS privates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user1_id INTEGER NOT NULL,
			user2_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user1_id, user2_id),
			CHECK(user1_id < user2_id),
			FOREIGN KEY(user1_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY(user2_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Messages
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_id INTEGER NOT NULL,
			private_id INTEGER,
			message_type TEXT NOT NULL,
			content TEXT NOT NULL,
			delivered INTEGER NOT NULL DEFAULT 0,
			read INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(from_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY(private_id) REFERENCES privates(id) ON DELETE CASCADE
		);`,
	}
	for _, t := range tables {
		if _, err := db.Exec(t); err != nil {
			log.Printf("Failed to create table %s, err: %v", t, err)
		}
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_messages_private_id ON messages(private_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages(from_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_privates_user1_id ON privates(user1_id);`,
		`CREATE INDEX IF NOT EXISTS idx_privates_user2_id ON privates(user2_id);`,
	}
	for _, i := range indexes {
		if _, err := db.Exec(i); err != nil {
			log.Printf("Failed to create index %s, err: %v", i, err)
		}
	}

	DB = db

	log.Println("DB initialised")
}

func Close() {
	if DB == nil {
		return
	}
	if err := DB.Close(); err != nil {
		log.Printf("Failed to close DB: %v", err)
	}
}
