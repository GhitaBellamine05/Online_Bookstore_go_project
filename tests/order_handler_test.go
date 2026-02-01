package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"online-bookstore/handlers"
	"online-bookstore/models"
)

func TestOrderHandler_Create(t *testing.T) {
	bookStore, authorStore, userStore, orderStore := SetupTestStores()
	
	author := models.Author{FirstName: "Test", LastName: "Author"}
	createdAuthor, _ := authorStore.CreateAuthor(author)
	
	book := models.Book{
		Title:       "Test Book",
		Author:      models.Author{ID: createdAuthor.ID},
		Genres:      []string{"Fiction"},
		PublishedAt: time.Now(),
		Price:       19.99,
		Stock:       10,
	}
	createdBook, _ := bookStore.CreateBook(book)

	handler := handlers.NewOrderHandler(orderStore, bookStore, userStore)

	t.Run("successful order creation", func(t *testing.T) {
		authHandler := handlers.NewAuthHandler(userStore)
		registerData := map[string]interface{}{
			"name":     "Customer User",
			"username": "customer",
			"email":    "customer@example.com",
			"password": "CustomerPass123!",
		}
		userReq := NewTestRequest("POST", "/auth/register", registerData)
		userRR := httptest.NewRecorder()
		authHandler.Register(userRR, userReq)
		
		userToken, err := ExtractTokenFromResponse(userRR)
		if err != nil {
			t.Fatalf("Failed to extract user token: %v", err)
		}
		orderData := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"book_id":   createdBook.ID,
					"quantity":  2,
				},
			},
			"total_price": 39.98,
		}

		req := NewAuthRequest("POST", "/orders", userToken, orderData)
		rr := httptest.NewRecorder()

		handler.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Logf("Response body: %s", rr.Body.String())
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		var response models.Order
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if len(response.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(response.Items))
		}
		if response.TotalPrice != 39.98 {
			t.Errorf("Expected total price 39.98, got %f", response.TotalPrice)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":     "Customer Two",
			"username": "customer2",
			"email":    "customer2@example.com",
			"password": "CustomerPass123!",
		}
		userReq := NewTestRequest("POST", "/auth/register", userData)
		userRR := httptest.NewRecorder()
		authHandler := handlers.NewAuthHandler(userStore)
		authHandler.Register(userRR, userReq)
		
		userToken, err := ExtractTokenFromResponse(userRR)
		if err != nil {
			t.Fatalf("Failed to extract user token: %v", err)
		}
		orderData := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"book_id":   createdBook.ID,
					"quantity":  10,
				},
			},
			"total_price": 199.90,
		}

		req := NewAuthRequest("POST", "/orders", userToken, orderData)
		rr := httptest.NewRecorder()

		handler.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Logf("Response body: %s", rr.Body.String())
			t.Errorf("Expected status %d for insufficient stock, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

func TestOrderHandler_GetAll(t *testing.T) {
    bookStore, authorStore, userStore, orderStore := SetupTestStores()

    author := models.Author{FirstName: "Test", LastName: "Author"}
    createdAuthor, _ := authorStore.CreateAuthor(author)
    book := models.Book{Title: "Test Book", Author: models.Author{ID: createdAuthor.ID}, Price: 19.99, Stock: 10}
    createdBook, _ := bookStore.CreateBook(book)
    
    authHandler := handlers.NewAuthHandler(userStore)
    registerData := map[string]interface{}{
        "name":     "Test User",
        "username": "testuser",
        "email":    "test@example.com",
        "password": "TestPass123!",
    }
    userReq := NewTestRequest("POST", "/auth/register", registerData)
    userRR := httptest.NewRecorder()
    authHandler.Register(userRR, userReq)
   
    userToken, _ := ExtractTokenFromResponse(userRR)
    
    orderHandler := handlers.NewOrderHandler(orderStore, bookStore, userStore)
   orderData := map[string]interface{}{
        "items": []map[string]interface{}{
            {
                "book_id":   createdBook.ID,
                "quantity":  1,
            },
        },
        "total_price": 19.99,
    }
    
    orderReq := NewAuthRequest("POST", "/orders", userToken, orderData)
    orderRR := httptest.NewRecorder()
    orderHandler.Create(orderRR, orderReq)
    
    req := NewAuthRequest("GET", "/orders", userToken, nil)
    rr := httptest.NewRecorder()
    orderHandler.ListAll(rr, req)
    
    if rr.Code != http.StatusOK {
        t.Logf("Response body: %s", rr.Body.String())
        t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
    }

    var orders []models.Order
    if err := json.Unmarshal(rr.Body.Bytes(), &orders); err != nil {
        t.Fatalf("Failed to parse response: %v", err)
    }
    if len(orders) != 1 {
        t.Logf("Found %d orders instead of 1", len(orders))
        t.Errorf("Expected 1 order, got %d", len(orders))
    }
}