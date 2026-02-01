package stores

import (
    "errors"
    "fmt"
    "sync"
    "time"
    "online-bookstore/models"
    "online-bookstore/stores"
)

type InMemoryOrderStore struct {
    mu        sync.RWMutex
    orders    map[int]models.Order
    nextID    int
    bookStore stores.BookStore
    userStore stores.UserStore 
}

var _ stores.OrderStore = (*InMemoryOrderStore)(nil)

func NewInMemoryOrderStore(initial map[int]models.Order, nextID int, bookStore stores.BookStore, userStore stores.UserStore) *InMemoryOrderStore {
    if initial == nil {
        initial = make(map[int]models.Order)
    }
    return &InMemoryOrderStore{
        orders:    initial,
        nextID:    nextID,
        bookStore: bookStore,
        userStore: userStore,
    }
}

func (s *InMemoryOrderStore) CreateOrder(o models.Order) (models.Order, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    user, err := s.userStore.GetUserByID(o.UserID)
    if err != nil {
        return models.Order{}, fmt.Errorf("user %d not found", o.UserID)
    }
    for _, item := range o.Items {
        currentStock, err := s.bookStore.GetStock(item.Book.ID)
        if err != nil {
            return models.Order{}, fmt.Errorf("book %d not found", item.Book.ID)
        }
        if currentStock < item.Quantity {
            return models.Order{}, fmt.Errorf("insufficient stock for book %d (available: %d, requested: %d)", 
                item.Book.ID, currentStock, item.Quantity)
        }
    }
    for _, item := range o.Items {
        if err := s.bookStore.ReduceStock(item.Book.ID, item.Quantity); err != nil {
            return models.Order{}, err
        }
    }

    o.ID = s.nextID
    o.CreatedAt = time.Now()
    o.Status = "confirmed"
    o.User = user  
    s.orders[o.ID] = o
    s.nextID++
    return o, nil
}
func (s *InMemoryOrderStore) GetOrder(id int) (models.Order, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    o, ok := s.orders[id]
    if !ok {
        return models.Order{}, errors.New("order not found")
    }
        if o.User.ID == 0 {
        user, err := s.userStore.GetUserByID(o.UserID)
        if err != nil {
            fmt.Printf("Warning: Could not load user %d for order %d: %v\n", o.UserID, id, err)
        } else {
            o.User = user
        }
    }
    
    return o, nil
}

func (s *InMemoryOrderStore) GetOrdersInTimeRange(start, end time.Time) ([]models.Order, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    var res []models.Order
    for _, o := range s.orders {
        if !o.CreatedAt.Before(start) && !o.CreatedAt.After(end) {
            res = append(res, o)
        }
    }
    return res, nil
}

func (s *InMemoryOrderStore) ListAll() []models.Order {
    s.mu.RLock()
    defer s.mu.RUnlock()
    var orders []models.Order
    for _, o := range s.orders {
        if o.User.ID == 0 {
            user, err := s.userStore.GetUserByID(o.UserID)
            if err == nil {
                o.User = user
            }
        }
        orders = append(orders, o)
    }
    return orders
}
func (s *InMemoryOrderStore) GetAllOrders() map[int]models.Order {
    s.mu.RLock()
    defer s.mu.RUnlock()
    result := make(map[int]models.Order, len(s.orders))
    for k, v := range s.orders {
        result[k] = v
    }
    return result
}

func (s *InMemoryOrderStore) GetNextID() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.nextID
}