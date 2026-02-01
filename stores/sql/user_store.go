package stores

import (
    "database/sql"
    "errors"
    "fmt"
	"time"
    "online-bookstore/models"
    "online-bookstore/stores"
)

type SQLUserStore struct {
    db *sql.DB
}

var _ stores.UserStore = (*SQLUserStore)(nil)

func NewSQLUserStore(db *sql.DB) *SQLUserStore {
    return &SQLUserStore{db: db}
}

func (s *SQLUserStore) CreateUser(user models.User) (models.User, error) {
    result, err := s.db.Exec(`
        INSERT INTO users (
            username, password, role, name, email,
            address_street, address_city, address_state,
            address_postal_code, address_country, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        user.Username, user.Password, user.Role, user.Name, user.Email,
        user.Address.Street, user.Address.City, user.Address.State,
        user.Address.PostalCode, user.Address.Country, user.CreatedAt.Format("2006-01-02T15:04:05Z"))
    if err != nil {
        if err.Error() == "UNIQUE constraint failed: users.username" {
            return models.User{}, errors.New("username already exists")
        }
        return models.User{}, fmt.Errorf("failed to create user: %w", err)
    }

    id, _ := result.LastInsertId()
    user.ID = int(id)
    return user, nil
}

func (s *SQLUserStore) GetUserByUsername(username string) (models.User, error) {
    row := s.db.QueryRow(`
        SELECT id, username, password, role, name, email,
               address_street, address_city, address_state,
               address_postal_code, address_country, created_at
        FROM users WHERE username = ?`, username)
    
    var user models.User
    var createdAtStr string
    err := row.Scan(
        &user.ID, &user.Username, &user.Password, &user.Role, &user.Name, &user.Email,
        &user.Address.Street, &user.Address.City, &user.Address.State,
        &user.Address.PostalCode, &user.Address.Country, &createdAtStr)
    if err == sql.ErrNoRows {
        return models.User{}, errors.New("user not found")
    } else if err != nil {
        return models.User{}, fmt.Errorf("failed to get user: %w", err)
    }
    
    user.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)
    return user, nil
}

func (s *SQLUserStore) GetUserByID(id int) (models.User, error) {
    row := s.db.QueryRow(`
        SELECT id, username, password, role, name, email,
               address_street, address_city, address_state,
               address_postal_code, address_country, created_at
        FROM users WHERE id = ?`, id)
    
    var user models.User
    var createdAtStr string
    err := row.Scan(
        &user.ID, &user.Username, &user.Password, &user.Role, &user.Name, &user.Email,
        &user.Address.Street, &user.Address.City, &user.Address.State,
        &user.Address.PostalCode, &user.Address.Country, &createdAtStr)
    if err == sql.ErrNoRows {
        return models.User{}, errors.New("user not found")
    } else if err != nil {
        return models.User{}, fmt.Errorf("failed to get user: %w", err)
    }
    
    user.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)
    return user, nil
}

func (s *SQLUserStore) ListAll() []models.User {
    rows, err := s.db.Query(`
        SELECT id, username, password, role, name, email,
               address_street, address_city, address_state,
               address_postal_code, address_country, created_at
        FROM users`)
    if err != nil {
        return nil
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        var createdAtStr string
        err := rows.Scan(
            &user.ID, &user.Username, &user.Password, &user.Role, &user.Name, &user.Email,
            &user.Address.Street, &user.Address.City, &user.Address.State,
            &user.Address.PostalCode, &user.Address.Country, &createdAtStr)
        if err != nil {
            continue
        }
        user.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAtStr)
        users = append(users, user)
    }
    return users
}