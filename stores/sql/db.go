package stores

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
)

func NewSQLiteDB(dbPath string) *sql.DB {
	dir := filepath.Dir(dbPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Fatal("Failed to create data directory:", err)
    }
    
	db, err := sql.Open("sqlite3", "./bookstore.db")
	if err != nil {
		log.Fatal("Failed to open SQLite DB:", err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	initSchema(db)
	return db
}

func initSchema(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		bio TEXT
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		address_street TEXT,
		address_city TEXT,
		address_state TEXT,
		address_postal_code TEXT,
		address_country TEXT,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		genres TEXT NOT NULL, -- JSON array
		published_at TEXT NOT NULL,
		price REAL NOT NULL,
		stock INTEGER NOT NULL,
		FOREIGN KEY(author_id) REFERENCES authors(id)
	);

	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		total_price REAL NOT NULL,
		created_at TEXT NOT NULL,
		status TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL,
		book_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		FOREIGN KEY(order_id) REFERENCES orders(id),
		FOREIGN KEY(book_id) REFERENCES books(id)
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}
}