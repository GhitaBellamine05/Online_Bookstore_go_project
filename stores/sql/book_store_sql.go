package stores

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
	"fmt"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type SQLBookStore struct {
	db *sql.DB
}

var _ stores.BookStore = (*SQLBookStore)(nil)

func NewSQLBookStore(db *sql.DB) *SQLBookStore {
	return &SQLBookStore{db: db}
}

func (s *SQLBookStore) CreateBook(book models.Book) (models.Book, error) {
    genresJSON, _ := json.Marshal(book.Genres)
    result, err := s.db.Exec(`
        INSERT INTO books (title, author_id, genres, published_at, price, stock)
        VALUES (?, ?, ?, ?, ?, ?)`,
        book.Title, book.Author.ID, string(genresJSON),
        book.PublishedAt.Format("2006-01-02T15:04:05Z"), book.Price, book.Stock)
    if err != nil {
        return models.Book{}, fmt.Errorf("failed to insert book: %w", err)
    }

    id, _ := result.LastInsertId()
    book.ID = int(id)
    
    // Load the complete author information for immediate response
    authorRow := s.db.QueryRow("SELECT id, first_name, last_name, bio FROM authors WHERE id = ?", book.Author.ID)
    var author models.Author
    err = authorRow.Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
    if err != nil {
        return models.Book{}, fmt.Errorf("failed to load author: %w", err)
    }
    book.Author = author
    
    return book, nil
}

func (s *SQLBookStore) GetBook(id int) (models.Book, error) {
	row := s.db.QueryRow(`
		SELECT id, title, author_id, genres, published_at, price, stock
		FROM books WHERE id = ?`, id)

	var book models.Book
	var genresJSON, publishedAtStr string
	var authorID int
	err := row.Scan(&book.ID, &book.Title, &authorID, &genresJSON, &publishedAtStr, &book.Price, &book.Stock)
	if err == sql.ErrNoRows {
		return models.Book{}, errors.New("book not found")
	} else if err != nil {
		return models.Book{}, fmt.Errorf("failed to scan book: %w", err)
	}

	book.PublishedAt, _ = time.Parse("2006-01-02T15:04:05Z", publishedAtStr)
	_ = json.Unmarshal([]byte(genresJSON), &book.Genres)

	// Load author
	authorRow := s.db.QueryRow("SELECT id, first_name, last_name, bio FROM authors WHERE id = ?", authorID)
	var author models.Author
	err = authorRow.Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
	if err != nil {
		return models.Book{}, fmt.Errorf("failed to load author: %w", err)
	}
	book.Author = author

	return book, nil
}

func (s *SQLBookStore) UpdateBook(id int, book models.Book) (models.Book, error) {
	genresJSON, _ := json.Marshal(book.Genres)
	_, err := s.db.Exec(`
		UPDATE books SET
			title = ?, author_id = ?, genres = ?, published_at = ?, price = ?, stock = ?
		WHERE id = ?`,
		book.Title, book.Author.ID, string(genresJSON),
		book.PublishedAt.Format("2006-01-02T15:04:05Z"), book.Price, book.Stock, id)
	if err != nil {
		return models.Book{}, fmt.Errorf("failed to update book: %w", err)
	}
	book.ID = id
	return book, nil
}

func (s *SQLBookStore) DeleteBook(id int) error {
	_, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}
	return nil
}

func (s *SQLBookStore) SearchBooks(c models.SearchCriteria) ([]models.Book, error) {
	query := `
		SELECT b.id, b.title, b.author_id, b.genres, b.published_at, b.price, b.stock,
		       a.id, a.first_name, a.last_name, a.bio
		FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE 1=1`
	args := []interface{}{}

	if c.Title != "" {
		query += " AND b.title = ?"
		args = append(args, c.Title)
	}
	if c.AuthorID != 0 {
		query += " AND b.author_id = ?"
		args = append(args, c.AuthorID)
	}
	if c.Genre != "" {
		query += " AND b.genres LIKE ?"
		args = append(args, "%"+c.Genre+"%")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var genresJSON, publishedAtStr string
		var authorID int
		var author models.Author
		err := rows.Scan(
			&b.ID, &b.Title, &authorID, &genresJSON, &publishedAtStr, &b.Price, &b.Stock,
			&author.ID, &author.FirstName, &author.LastName, &author.Bio)
		if err != nil {
			continue
		}
		b.PublishedAt, _ = time.Parse("2006-01-02T15:04:05Z", publishedAtStr)
		_ = json.Unmarshal([]byte(genresJSON), &b.Genres)
		b.Author = author
		books = append(books, b)
	}
	return books, nil
}

func (s *SQLBookStore) ListAll() []models.Book {
	rows, err := s.db.Query(`
		SELECT b.id, b.title, b.author_id, b.genres, b.published_at, b.price, b.stock,
		       a.id, a.first_name, a.last_name, a.bio
		FROM books b
		JOIN authors a ON b.author_id = a.id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var genresJSON, publishedAtStr string
		var authorID int
		var author models.Author
		err := rows.Scan(
			&b.ID, &b.Title, &authorID, &genresJSON, &publishedAtStr, &b.Price, &b.Stock,
			&author.ID, &author.FirstName, &author.LastName, &author.Bio)
		if err != nil {
			continue
		}
		b.PublishedAt, _ = time.Parse("2006-01-02T15:04:05Z", publishedAtStr)
		_ = json.Unmarshal([]byte(genresJSON), &b.Genres)
		b.Author = author
		books = append(books, b)
	}
	return books
}

//  NEW: Stock management
// stores/sql/book_store.go
func (s *SQLBookStore) ReduceStock(bookID, quantity int) error {
    // You need 3 parameters: quantity, bookID, quantity (for stock >= check)
    result, err := s.db.Exec("UPDATE books SET stock = stock - ? WHERE id = ? AND stock >= ?", quantity, bookID, quantity)
    if err != nil {
        return fmt.Errorf("failed to reduce stock: %w", err)
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        // Check if book exists
        var count int
        err = s.db.QueryRow("SELECT COUNT(*) FROM books WHERE id = ?", bookID).Scan(&count)
        if err != nil {
            return fmt.Errorf("failed to check book existence: %w", err)
        }
        if count == 0 {
            return fmt.Errorf("book %d not found", bookID)
        }
        // Book exists but insufficient stock
        var currentStock int
        err = s.db.QueryRow("SELECT stock FROM books WHERE id = ?", bookID).Scan(&currentStock)
        if err != nil {
            return fmt.Errorf("failed to get current stock: %w", err)
        }
        return fmt.Errorf("insufficient stock for book %d (available: %d, requested: %d)", bookID, currentStock, quantity)
    }
    return nil
}
func (s *SQLBookStore) GetStock(bookID int) (int, error) {
    var stock int
    err := s.db.QueryRow("SELECT stock FROM books WHERE id = ?", bookID).Scan(&stock)
    if err == sql.ErrNoRows {
        return 0, fmt.Errorf("book %d not found", bookID)
    } else if err != nil {
        return 0, fmt.Errorf("failed to get stock: %w", err)
    }
    return stock, nil
}