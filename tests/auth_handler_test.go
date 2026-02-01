package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"online-bookstore/handlers"
	"online-bookstore/models"
)

func TestAuthHandler_Register(t *testing.T) {
	_, _, userStore, _ := SetupTestStores()
	handler := handlers.NewAuthHandler(userStore)

	t.Run("successful registration", func(t *testing.T) {
		registerData := map[string]interface{}{
			"name":     "Test User",
			"username": "testuser",
			"email":    "test@example.com",
			"password": "SecurePass123!",
		}

		req := NewTestRequest("POST", "/auth/register", registerData)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		var response models.AuthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", response.Username)
		}
		if response.Role != "customer" {
			t.Errorf("Expected role 'customer', got %v", response.Role)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		registerData := map[string]interface{}{
			"username": "incomplete",
			"password": "weakpass", //HERE we are missing name and email
		}

		req := NewTestRequest("POST", "/auth/register", registerData)
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		// Register first user
		registerData1 := map[string]interface{}{
			"name":     "First User",
			"username": "duplicate",
			"email":    "first@example.com",
			"password": "SecurePass123!",
		}
		req1 := NewTestRequest("POST", "/auth/register", registerData1)
		rr1 := httptest.NewRecorder()
		handler.Register(rr1, req1)

		// Try to register same username
		registerData2 := map[string]interface{}{
			"name":     "Second User",
			"username": "duplicate",
			"email":    "second@example.com",
			"password": "AnotherPass123!",
		}
		req2 := NewTestRequest("POST", "/auth/register", registerData2)
		rr2 := httptest.NewRecorder()
		handler.Register(rr2, req2)

		if rr2.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for duplicate username, got %d", http.StatusBadRequest, rr2.Code)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	_, _, userStore, _ := SetupTestStores()
	handler := handlers.NewAuthHandler(userStore)

	// First, register a user
	registerData := map[string]interface{}{
		"name":     "Login User",
		"username": "loginuser",
		"email":    "login@example.com",
		"password": "LoginPass123!",
	}
	registerReq := NewTestRequest("POST", "/auth/register", registerData)
	registerRR := httptest.NewRecorder()
	handler.Register(registerRR, registerReq)

	t.Run("successful login", func(t *testing.T) {
		loginData := models.LoginRequest{
			Username: "loginuser",
			Password: "LoginPass123!",
		}
		req := NewTestRequest("POST", "/auth/login", loginData)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var response models.AuthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response.Token == "" {
			t.Error("Expected token in response")
		}
		if response.Username != "loginuser" {
			t.Errorf("Expected username 'loginuser', got %v", response.Username)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		loginData := models.LoginRequest{
			Username: "loginuser",
			Password: "wrongpassword",
		}
		req := NewTestRequest("POST", "/auth/login", loginData)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		loginData := models.LoginRequest{
			Username: "nonexistent",
			Password: "anypassword",
		}
		req := NewTestRequest("POST", "/auth/login", loginData)
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})
}