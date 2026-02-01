package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"online-bookstore/auth"
	"online-bookstore/handlers"
	"online-bookstore/middleware"
	"online-bookstore/models"
	"online-bookstore/stores"
)
func adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}
		
		tokenString := authHeader[7:]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		if claims.Role != "admin" {
			http.Error(w, `{"error":"admin access required"}`, http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}
func createAuthorTestServer(authorStore, userStore interface{}) *httptest.Server {
	ah := handlers.NewAuthorHandler(authorStore.(stores.AuthorStore))
	authHandler := handlers.NewAuthHandler(userStore.(stores.UserStore))
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/authors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ah.ListAll(w, r)
		case http.MethodPost:
			adminMiddleware(middleware.ValidateAuthor(ah.Create))(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestAuthorHandler_Create(t *testing.T) {
	_, authorStore, userStore, _ := SetupTestStores()
	server := createAuthorTestServer(authorStore, userStore)
	defer server.Close()

	t.Run("successful author creation", func(t *testing.T) {
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

		authorData := models.Author{
			FirstName: "New",
			LastName:  "Author",
			Bio:       "New author bio",
		}

		authorBody, _ := json.Marshal(authorData)
		req, _ := http.NewRequest("POST", server.URL+"/authors", bytes.NewReader(authorBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Author create request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		var response models.Author
		json.NewDecoder(resp.Body).Decode(&response)
		if response.FirstName != "New" {
			t.Errorf("Expected first name 'New', got %s", response.FirstName)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		registerData := map[string]interface{}{
			"name":     "Admin User",
			"username": "admin2",
			"email":    "admin2@example.com",
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

		// Test with empty first name
		authorData := models.Author{
			FirstName: "", 
			LastName:  "Author",
			Bio:       "Test bio",
		}

		authorBody, _ := json.Marshal(authorData)
		req, _ := http.NewRequest("POST", server.URL+"/authors", bytes.NewReader(authorBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Author create request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing fields, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

func TestAuthorHandler_GetAll(t *testing.T) {
	_, authorStore, _, _ := SetupTestStores()
	author1 := models.Author{FirstName: "Author", LastName: "One", Bio: "Bio 1"}
	author2 := models.Author{FirstName: "Author", LastName: "Two", Bio: "Bio 2"}
	authorStore.CreateAuthor(author1)
	authorStore.CreateAuthor(author2)

	handler := handlers.NewAuthorHandler(authorStore)
	req := httptest.NewRequest("GET", "/authors", nil)
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var authors []models.Author
	if err := json.Unmarshal(rr.Body.Bytes(), &authors); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(authors))
	}
}