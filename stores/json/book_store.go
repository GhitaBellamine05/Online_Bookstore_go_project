package stores

import (
	"errors"
	"fmt"
	"sync"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type InMemoryBookStore struct {
	mu     sync.RWMutex
	books  map[int]models.Book
	nextID int
	authorStore stores.AuthorStore 
}

var _ stores.BookStore = (*InMemoryBookStore)(nil)

func NewInMemoryBookStore(initial map[int]models.Book, nextID int,authorStore stores.AuthorStore) *InMemoryBookStore {
	if initial == nil {
		initial = make(map[int]models.Book)
	}
	return &InMemoryBookStore{
		books:  initial,
		nextID: nextID,
		authorStore: authorStore,
	}
}
func (s *InMemoryBookStore) CreateBook(book models.Book) (models.Book, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.authorStore != nil && book.Author.ID != 0 {
        // Get full author from author store
        author, err := s.authorStore.GetAuthor(book.Author.ID)
        if err != nil {
            return models.Book{}, fmt.Errorf("author %d not found", book.Author.ID)
        }
        book.Author = author
    } else if book.Author.FirstName != "" && book.Author.LastName != "" {
    } else {
        return models.Book{}, errors.New("author ID or complete author name is required")
    }

    book.ID = s.nextID
    s.books[book.ID] = book
    s.nextID++
    return book, nil
}

func (s *InMemoryBookStore) GetBook(id int) (models.Book, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
	if s.books == nil {
        return models.Book{}, errors.New("books store not initialized")
    }
    
    book, ok := s.books[id]
    if !ok {
        return models.Book{}, errors.New("book not found")
    }
    
    if s.authorStore != nil && book.Author.ID != 0 && (book.Author.FirstName == "" || book.Author.LastName == "") {
        author, err := s.authorStore.GetAuthor(book.Author.ID)
        if err == nil {
            book.Author = author
        }
    }
    
    return book, nil
}

func (s *InMemoryBookStore) UpdateBook(id int, book models.Book) (models.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.books[id]; !exists {
		return models.Book{}, errors.New("book not found")
	}
	book.ID = id
	s.books[id] = book
	return book, nil
}

func (s *InMemoryBookStore) DeleteBook(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.books, id)
	return nil
}

func (s *InMemoryBookStore) SearchBooks(c models.SearchCriteria) ([]models.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []models.Book
	for _, b := range s.books {
		if c.Title != "" && b.Title != c.Title {
			continue
		}
		if c.AuthorID != 0 && b.Author.ID != c.AuthorID {
			continue
		}
		if c.Genre != "" {
			found := false
			for _, g := range b.Genres {
				if g == c.Genre {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		res = append(res, b)
	}
	return res, nil
}

func (s *InMemoryBookStore) ListAll() []models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var books []models.Book
	for _, b := range s.books {
		books = append(books, b)
	}
	return books
}

func (s *InMemoryBookStore) GetAllBooks() map[int]models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[int]models.Book, len(s.books))
	for k, v := range s.books {
		result[k] = v
	}
	return result
}

func (s *InMemoryBookStore) GetNextID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextID
}

func (s *InMemoryBookStore) ReduceStock(bookID, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, exists := s.books[bookID]
	if !exists {
		return fmt.Errorf("book %d not found", bookID)
	}
	if book.Stock < quantity {
		return fmt.Errorf("insufficient stock for book %d (available: %d, requested: %d)", bookID, book.Stock, quantity)
	}
	book.Stock -= quantity
	s.books[bookID] = book
	return nil
}

func (s *InMemoryBookStore) GetStock(bookID int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	book, exists := s.books[bookID]
	if !exists {
		return 0, fmt.Errorf("book %d not found", bookID)
	}
	return book.Stock, nil
}