package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"online-bookstore/handlers"
	"online-bookstore/middleware"
	"online-bookstore/models"
	"online-bookstore/stores"
	"testing"
	"time"
)
func createBookTestServer(bookStore, userStore interface{}) *httptest.Server {
	bh := handlers.NewBookHandler(bookStore.(stores.BookStore))
	authHandler := handlers.NewAuthHandler(userStore.(stores.UserStore))
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			bh.ListAll(w, r)
		case http.MethodPost:
			adminMiddleware(middleware.ValidateBook(bh.Create))(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestBookHandler_Create(t *testing.T) {
	bookStore, authorStore, userStore, _ := SetupTestStores()
	author := models.Author{
		FirstName: "Test",
		LastName:  "Author",
		Bio:       "Test bio",
	}
	createdAuthor, _ := authorStore.CreateAuthor(author)
	server := createBookTestServer(bookStore, userStore)
	defer server.Close()

	t.Run("successful book creation", func(t *testing.T) {
		// Create admin user
		registerData := map[string]interface{}{
			"name":     "Admin User",
			"username": "admin",
			"email":    "admin@example.com",
			"password": "AdminPass123!",
			"role":     "admin",
		}
		
		registerBody, _ := json.Marshal(registerData)
		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewReader(registerBody))
		if err != nil {
			t.Fatalf("Register request failed: %v", err)
		}
		defer resp.Body.Close()
		
		var authResp models.AuthResponse
		json.NewDecoder(resp.Body).Decode(&authResp)
		adminToken := authResp.Token

		bookData := models.Book{
			Title:       "Test Book",
			Author:      models.Author{ID: createdAuthor.ID},
			Genres:      []string{"Fiction"},
			PublishedAt: time.Now(),
			Price:       19.99,
			Stock:       10,
		}

		bookBody, _ := json.Marshal(bookData)
		req, _ := http.NewRequest("POST", server.URL+"/books", bytes.NewReader(bookBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Book create request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("unauthorized book creation", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":     "Regular User",
			"username": "regular",
			"email":    "user@example.com",
			"password": "UserPass123!",
			// no role field by  default customer
		}
		
		userBody, _ := json.Marshal(userData)
		resp, err := http.Post(server.URL+"/auth/register", "application/json", bytes.NewReader(userBody))
		if err != nil {
			t.Fatalf("Register request failed: %v", err)
		}
		defer resp.Body.Close()
		
		var authResp models.AuthResponse
		json.NewDecoder(resp.Body).Decode(&authResp)
		userToken := authResp.Token

		bookData := models.Book{
			Title:       "Unauthorized Book",
			Author:      models.Author{ID: createdAuthor.ID},
			Genres:      []string{"Fiction"},
			PublishedAt: time.Now(),
			Price:       19.99,
			Stock:       10,
		}

		bookBody, _ := json.Marshal(bookData)
		req, _ := http.NewRequest("POST", server.URL+"/books", bytes.NewReader(bookBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Book create request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
		}
	})
}

func TestBookHandler_GetAll(t *testing.T) {
	bookStore, authorStore, _, _ := SetupTestStores()
	author := models.Author{FirstName: "Test", LastName: "Author"}
	createdAuthor, _ := authorStore.CreateAuthor(author)
	
	book1 := models.Book{
		Title:       "Book 1",
		Author:      models.Author{ID: createdAuthor.ID},
		Genres:      []string{"Fiction"},
		PublishedAt: time.Now(),
		Price:       19.99,
		Stock:       10,
	}
	book2 := models.Book{
		Title:       "Book 2",
		Author:      models.Author{ID: createdAuthor.ID},
		Genres:      []string{"Non-Fiction"},
		PublishedAt: time.Now(),
		Price:       24.99,
		Stock:       5,
	}
	bookStore.CreateBook(book1)
	bookStore.CreateBook(book2)

	handler := handlers.NewBookHandler(bookStore)
	req := httptest.NewRequest("GET", "/books", nil)
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var books []models.Book
	if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}