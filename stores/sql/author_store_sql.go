// stores/sql/author_store.go
package stores

import (
	"database/sql"
	"errors"
	"fmt"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type SQLAuthorStore struct {
	db *sql.DB
}

var _ stores.AuthorStore = (*SQLAuthorStore)(nil)

func NewSQLAuthorStore(db *sql.DB) *SQLAuthorStore {
	return &SQLAuthorStore{db: db}
}

func (s *SQLAuthorStore) CreateAuthor(a models.Author) (models.Author, error) {
	result, err := s.db.Exec(`
		INSERT INTO authors (first_name, last_name, bio)
		VALUES (?, ?, ?)`,
		a.FirstName, a.LastName, a.Bio)
	if err != nil {
		return models.Author{}, fmt.Errorf("failed to insert author: %w", err)
	}

	id, _ := result.LastInsertId()
	a.ID = int(id)
	return a, nil
}

func (s *SQLAuthorStore) GetAuthor(id int) (models.Author, error) {
	row := s.db.QueryRow(`
		SELECT id, first_name, last_name, bio
		FROM authors WHERE id = ?`, id)

	var a models.Author
	err := row.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Bio)
	if err == sql.ErrNoRows {
		return models.Author{}, errors.New("author not found")
	} else if err != nil {
		return models.Author{}, fmt.Errorf("failed to scan author: %w", err)
	}
	return a, nil
}

func (s *SQLAuthorStore) UpdateAuthor(id int, a models.Author) (models.Author, error) {
	_, err := s.db.Exec(`
		UPDATE authors SET first_name = ?, last_name = ?, bio = ?
		WHERE id = ?`,
		a.FirstName, a.LastName, a.Bio, id)
	if err != nil {
		return models.Author{}, fmt.Errorf("failed to update author: %w", err)
	}
	a.ID = id
	return a, nil
}

func (s *SQLAuthorStore) DeleteAuthor(id int) error {
	_, err := s.db.Exec("DELETE FROM authors WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}
	return nil
}

func (s *SQLAuthorStore) ListAll() []models.Author {
	rows, err := s.db.Query("SELECT id, first_name, last_name, bio FROM authors")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var authors []models.Author
	for rows.Next() {
		var a models.Author
		err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Bio)
		if err != nil {
			continue
		}
		authors = append(authors, a)
	}
	return authors
}