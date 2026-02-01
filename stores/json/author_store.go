package stores

import (
	"errors"
	//"log"
	"sync"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type InMemoryAuthorStore struct {
	mu      sync.RWMutex
	authors map[int]models.Author
	nextID  int
}

var _ stores.AuthorStore = (*InMemoryAuthorStore)(nil)

func NewInMemoryAuthorStore(initial map[int]models.Author, nextID int) *InMemoryAuthorStore {
	if initial == nil {
		initial = make(map[int]models.Author)
	}
	return &InMemoryAuthorStore{
		authors: initial,
		nextID:  nextID,
	}
}

func (s *InMemoryAuthorStore) CreateAuthor(a models.Author) (models.Author, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.ID = s.nextID
	s.authors[a.ID] = a
	s.nextID++
	return a, nil
}

func (s *InMemoryAuthorStore) GetAuthor(id int) (models.Author, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.authors[id]
	if !ok {
		return models.Author{}, errors.New("author not found")
	}
	return a, nil
}

func (s *InMemoryAuthorStore) UpdateAuthor(id int, author models.Author) (models.Author, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    //log.Printf("UpdateAuthor called with id=%d, author.ID=%d", id, author.ID)
    //log.Printf("Available authors: %v", s.authors)
    
    if _, exists := s.authors[id]; !exists {
        //log.Printf("Author with id=%d not found", id)
        return models.Author{}, errors.New("author not found")
    }
    
    author.ID = id
    s.authors[id] = author
    //og.Printf("Author updated successfully: %+v", author)
    return author, nil
}

func (s *InMemoryAuthorStore) DeleteAuthor(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.authors, id)
	return nil
}

func (s *InMemoryAuthorStore) ListAll() []models.Author {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var authors []models.Author
	for _, a := range s.authors {
		authors = append(authors, a)
	}
	return authors
}

func (s *InMemoryAuthorStore) GetAllAuthors() map[int]models.Author {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[int]models.Author, len(s.authors))
	for k, v := range s.authors {
		result[k] = v
	}
	return result
}

func (s *InMemoryAuthorStore) GetNextID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextID
}