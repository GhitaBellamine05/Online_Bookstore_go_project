package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"online-bookstore/models"
	"regexp"
	"strings"
)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidPrice(price float64) bool {
	return price > 0
}

func isValidStock(stock int) bool {
	return stock >= 0
}

func isValidQuantity(quantity int) bool {
	return quantity > 0
}

func sendValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func ValidateBook(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}
		var book models.Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			sendValidationError(w, "invalid JSON format")
			return
		}
		if strings.TrimSpace(book.Title) == "" {
			sendValidationError(w, "book title is required")
			return
		}
		if book.Author.ID == 0 && (book.Author.FirstName == "" || book.Author.LastName == "") {
			sendValidationError(w, "author ID or full author name is required")
			return
		}
		if len(book.Genres) == 0 {
			sendValidationError(w, "at least one genre is required")
			return
		}
		if !isValidPrice(book.Price) {
			sendValidationError(w, "price must be greater than 0")
			return
		}
		if !isValidStock(book.Stock) {
			sendValidationError(w, "stock cannot be negative")
			return
		}
		body, _ := json.Marshal(book)
		r.Body = http.MaxBytesReader(w, io.NopCloser(strings.NewReader(string(body))), 1048576)
		next.ServeHTTP(w, r)
	}
}
func ValidateCustomer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}
		var customer models.User
		if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
			sendValidationError(w, "invalid JSON format")
			return
		}
		if strings.TrimSpace(customer.Name) == "" {
			sendValidationError(w, "customer name is required")
			return
		}
		if !isValidEmail(customer.Email) {
			sendValidationError(w, "invalid email format")
			return
		}
		if strings.TrimSpace(customer.Address.Street) == "" ||
			strings.TrimSpace(customer.Address.City) == "" ||
			strings.TrimSpace(customer.Address.Country) == "" {
			sendValidationError(w, "complete address is required (street, city, country)")
			return
		}
		body, _ := json.Marshal(customer)
		r.Body = http.MaxBytesReader(w, io.NopCloser(strings.NewReader(string(body))), 1048576)
		next.ServeHTTP(w, r)
	}
}
func ValidateOrder(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		var orderReq struct {
			Items []struct {
				BookID   int     `json:"book_id"`
				Quantity int     `json:"quantity"`
			} `json:"items"`
			TotalPrice float64 `json:"total_price"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
			sendValidationError(w, "invalid JSON format")
			return
		}
		
		if len(orderReq.Items) == 0 {
			sendValidationError(w, "at least one item is required")
			return
		}
		
		for i, item := range orderReq.Items {
			if item.BookID == 0 {
				sendValidationError(w, fmt.Sprintf("item %d: book ID is required", i+1))
				return
			}
			if !isValidQuantity(item.Quantity) {
				sendValidationError(w, fmt.Sprintf("item %d: quantity must be greater than 0", i+1))
				return
			}
		}
		
		if !isValidPrice(orderReq.TotalPrice) {
			sendValidationError(w, "total price must be greater than 0")
			return
		}
		body, _ := json.Marshal(orderReq)
		r.Body = http.MaxBytesReader(w, io.NopCloser(strings.NewReader(string(body))), 1048576)
		next.ServeHTTP(w, r)
	}
}
func ValidateAuthor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		var author models.Author
		if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
			sendValidationError(w, "invalid JSON format")
			return
		}
		if r.Method == http.MethodPost {
			if author.ID != 0 {
				sendValidationError(w, "author ID should not be provided when creating")
				return
			}
		}
		
		if r.Method == http.MethodPut {
		}

		if strings.TrimSpace(author.FirstName) == "" {
			sendValidationError(w, "first name is required")
			return
		}

		if strings.TrimSpace(author.LastName) == "" {
			sendValidationError(w, "last name is required")
			return
		}
		body, _ := json.Marshal(author)
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		next.ServeHTTP(w, r)
	}
}