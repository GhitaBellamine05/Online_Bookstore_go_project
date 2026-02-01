package models

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type Address struct {
    Street      string `json:"street,omitempty"`
    City        string `json:"city,omitempty"`
    State       string `json:"state,omitempty"`
    PostalCode  string `json:"postal_code,omitempty"`
    Country     string `json:"country,omitempty"`
}

type User struct {
    ID        int     `json:"id,omitempty"`
    Username  string  `json:"username" validate:"required,min=3"`
    Password  string  `json:"password" validate:"required,min=8"` 
    Role      string  `json:"role" validate:"required"`           
    Name      string  `json:"name" validate:"required"`
    Email     string  `json:"email" validate:"required,email"`
    Address   Address `json:"address"`
    CreatedAt time.Time `json:"created_at,omitempty"`
}

type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}
type AuthResponse struct {
    Token    string `json:"token"`
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    Name     string `json:"name"`
    Email    string `json:"email"`
}

type Claims struct {
    UserID int    `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

type OrderItem struct {
    Book   Book   `json:"book"`
    Quantity int  `json:"quantity" validate:"required,min=1"`
}

type Order struct {
    ID         int         `json:"id,omitempty"`
    UserID     int         `json:"user_id"`        
    User       User        `json:"user"`         
    Items      []OrderItem `json:"items"`
    TotalPrice float64     `json:"total_price"`
    CreatedAt  time.Time   `json:"created_at,omitempty"`
    Status     string      `json:"status,omitempty"`
}

type Book struct {
    ID          int      `json:"id,omitempty"`
    Title       string   `json:"title" validate:"required"`
    Author      Author   `json:"author"`
    Genres      []string `json:"genres" validate:"required"`
    PublishedAt time.Time `json:"published_at" validate:"required"`
    Price       float64  `json:"price" validate:"required"`
    Stock       int      `json:"stock" validate:"required"`
}

type Author struct {
    ID        int    `json:"id,omitempty"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    Bio       string `json:"bio,omitempty"`
}

type SearchCriteria struct {
    Title    string
    AuthorID int
    Genre    string
    Fuzzy    bool
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

type TopSellingBook struct {
    Book         Book  `json:"book"`
    QuantitySold int   `json:"quantity_sold"`
}

type SalesReport struct {
    Timestamp       time.Time          `json:"timestamp"`
    TotalRevenue    float64           `json:"total_revenue"`
    TotalOrders     int               `json:"total_orders"`
    TopSellingBooks []TopSellingBook  `json:"top_selling_books"`
}