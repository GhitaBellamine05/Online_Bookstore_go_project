package stores

import (
    "errors"
    "sync"
    "online-bookstore/models"
    "online-bookstore/stores"
)

type InMemoryUserStore struct {
    mu    sync.RWMutex
    users map[string]models.User 
    nextID int
}

var _ stores.UserStore = (*InMemoryUserStore)(nil)

func NewInMemoryUserStore(initial map[int]models.User, nextID int) *InMemoryUserStore {
    users := make(map[string]models.User)
    for _, user := range initial {
        users[user.Username] = user
    }
    return &InMemoryUserStore{
        users:  users,
        nextID: nextID,
    }
}

func (s *InMemoryUserStore) CreateUser(user models.User) (models.User, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.users[user.Username]; exists {
        return models.User{}, errors.New("username already exists")
    }
    
    user.ID = s.nextID
    s.users[user.Username] = user
    s.nextID++
    return user, nil
}

func (s *InMemoryUserStore) GetUserByUsername(username string) (models.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    user, exists := s.users[username]
    if !exists {
        return models.User{}, errors.New("user not found")
    }
    return user, nil
}

func (s *InMemoryUserStore) GetUserByID(id int) (models.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    for _, user := range s.users {
        if user.ID == id {
            return user, nil
        }
    }
    return models.User{}, errors.New("user not found")
}

func (s *InMemoryUserStore) ListAll() []models.User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    var users []models.User
    for _, user := range s.users {
        users = append(users, user)
    }
    return users
}

func (s *InMemoryUserStore) GetAllUsers() map[int]models.User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    result := make(map[int]models.User)
    for _, user := range s.users {
        result[user.ID] = user
    }
    return result
}

func (s *InMemoryUserStore) GetNextID() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.nextID
}