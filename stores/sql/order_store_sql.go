package stores

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "time"
    "online-bookstore/models"
    "online-bookstore/stores"
)

type SQLOrderStore struct {
    db        *sql.DB
    bookStore stores.BookStore
    userStore stores.UserStore 
}

var _ stores.OrderStore = (*SQLOrderStore)(nil)

func NewSQLOrderStore(db *sql.DB, bookStore stores.BookStore, userStore stores.UserStore) *SQLOrderStore {
    return &SQLOrderStore{db: db, bookStore: bookStore, userStore: userStore}
}

func (s *SQLOrderStore) CreateOrder(o models.Order) (models.Order, error) {
    _, err := s.userStore.GetUserByID(o.UserID)
    if err != nil {
        return models.Order{}, fmt.Errorf("user %d not found", o.UserID)
    }
    tx, err := s.db.Begin()
    if err != nil {
        return models.Order{}, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback() 

    for _, item := range o.Items {
        var currentStock int
        err = tx.QueryRow("SELECT stock FROM books WHERE id = ?", item.Book.ID).Scan(&currentStock)
        if err != nil {
            return models.Order{}, fmt.Errorf("book %d not found", item.Book.ID)
        }
        if currentStock < item.Quantity {
            return models.Order{}, fmt.Errorf("insufficient stock for book %d (available: %d, requested: %d)", 
                item.Book.ID, currentStock, item.Quantity)
        }

        _, err = tx.Exec("UPDATE books SET stock = stock - ? WHERE id = ?", item.Quantity, item.Book.ID)
        if err != nil {
            return models.Order{}, fmt.Errorf("failed to reduce stock for book %d: %w", item.Book.ID, err)
        }
    }
    o.CreatedAt = time.Now()
    o.Status = "confirmed"
    result, err := tx.Exec(`
        INSERT INTO orders (user_id, total_price, created_at, status)
        VALUES (?, ?, ?, ?)`,
        o.UserID, o.TotalPrice, o.CreatedAt.Format("2006-01-02T15:04:05Z"), o.Status)
    if err != nil {
        return models.Order{}, fmt.Errorf("failed to insert order: %w", err)
    }

    orderID, _ := result.LastInsertId()
    o.ID = int(orderID)
    for _, item := range o.Items {
        _, err := tx.Exec(`
            INSERT INTO order_items (order_id, book_id, quantity)
            VALUES (?, ?, ?)`,
            o.ID, item.Book.ID, item.Quantity)
        if err != nil {
            return models.Order{}, fmt.Errorf("failed to insert order item: %w", err)
        }
    }
    user, err := s.userStore.GetUserByID(o.UserID)
    if err != nil {
        return models.Order{}, fmt.Errorf("failed to load user: %w", err)
    }
    o.User = user
    var items []models.OrderItem
    for _, item := range o.Items {
        book, err := s.bookStore.GetBook(item.Book.ID)
        if err != nil {
            return models.Order{}, fmt.Errorf("failed to load book %d: %w", item.Book.ID, err)
        }
        items = append(items, models.OrderItem{
            Book:     book,
            Quantity: item.Quantity,
        })
    }
    o.Items = items

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return models.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return o, nil
}
func (s *SQLOrderStore) GetOrder(id int) (models.Order, error) {
    row := s.db.QueryRow(`
        SELECT id, user_id, total_price, created_at, status
        FROM orders WHERE id = ?`, id)

    var o models.Order
    var createdAtStr string
    err := row.Scan(&o.ID, &o.UserID, &o.TotalPrice, &createdAtStr, &o.Status)
    if err == sql.ErrNoRows {
        return models.Order{}, errors.New("order not found")
    } else if err != nil {
        return models.Order{}, fmt.Errorf("failed to scan order: %w", err)
    }

    o.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)

    // Load user
    user, err := s.userStore.GetUserByID(o.UserID)
    if err != nil {
        return models.Order{}, fmt.Errorf("failed to load user: %w", err)
    }
    o.User = user

    // Load items
    itemsRows, err := s.db.Query(`
        SELECT oi.book_id, oi.quantity, b.title, b.author_id, b.genres, b.published_at, b.price, b.stock
        FROM order_items oi
        JOIN books b ON oi.book_id = b.id
        WHERE oi.order_id = ?`, o.ID)
    if err != nil {
        return models.Order{}, fmt.Errorf("failed to load items: %w", err)
    }
    defer itemsRows.Close()

    var items []models.OrderItem
    for itemsRows.Next() {
        var item models.OrderItem
        var genresJSON, publishedAtStr string
        var authorID int
        err := itemsRows.Scan(
            &item.Book.ID, &item.Quantity,
            &item.Book.Title, &authorID, &genresJSON, &publishedAtStr,
            &item.Book.Price, &item.Book.Stock)
        if err != nil {
            continue
        }
        item.Book.PublishedAt, _ = time.Parse("2006-01-02T15:04:05Z", publishedAtStr)
        _ = json.Unmarshal([]byte(genresJSON), &item.Book.Genres)

        // Load author
        authorRow := s.db.QueryRow("SELECT id, first_name, last_name, bio FROM authors WHERE id = ?", authorID)
        var author models.Author
        authorRow.Scan(&author.ID, &author.FirstName, &author.LastName, &author.Bio)
        item.Book.Author = author

        items = append(items, item)
    }
    o.Items = items

    return o, nil
}

func (s *SQLOrderStore) GetOrdersInTimeRange(start, end time.Time) ([]models.Order, error) {
    rows, err := s.db.Query(`
        SELECT id, user_id, total_price, created_at, status
        FROM orders
        WHERE created_at BETWEEN ? AND ?`,
        start.Format("2006-01-02T15:04:05Z"),
        end.Format("2006-01-02T15:04:05Z"))
    if err != nil {
        return nil, fmt.Errorf("failed to query orders: %w", err)
    }
    defer rows.Close()

    var orders []models.Order
    for rows.Next() {
        var o models.Order
        var createdAtStr string
        err := rows.Scan(&o.ID, &o.UserID, &o.TotalPrice, &createdAtStr, &o.Status)
        if err != nil {
            continue
        }
        o.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)
        orders = append(orders, o)
    }
    return orders, nil
}

func (s *SQLOrderStore) ListAll() []models.Order {
    rows, err := s.db.Query("SELECT id, user_id, total_price, created_at, status FROM orders")
    if err != nil {
        return nil
    }
    defer rows.Close()

    var orders []models.Order
    for rows.Next() {
        var o models.Order
        var createdAtStr string
        err := rows.Scan(&o.ID, &o.UserID, &o.TotalPrice, &createdAtStr, &o.Status)
        if err != nil {
            continue
        }
        o.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)
        orders = append(orders, o)
    }
    return orders
}