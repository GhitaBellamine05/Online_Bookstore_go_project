package tests

import (
	"testing"
	"online-bookstore/models"
	jsonstores "online-bookstore/stores/json"
)

func TestInMemoryUserStore_CreateUser(t *testing.T) {
	db := CreateTestDB()
	store := jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])

	user := models.User{
		Username: "testuser",
		Password: "hashedpassword",
		Role:     "customer",
		Name:     "Test User",
		Email:    "test@example.com",
	}

	createdUser, err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if createdUser.ID != 1 {
		t.Errorf("Expected ID 1, got %d", createdUser.ID)
	}
	if createdUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", createdUser.Username)
	}
	if createdUser.Role != "customer" {
		t.Errorf("Expected role 'customer', got %s", createdUser.Role)
	}
}

func TestInMemoryUserStore_GetUserByUsername(t *testing.T) {
	db := CreateTestDB()
	store := jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])

	user := models.User{
		Username: "findme",
		Password: "hashedpassword",
		Role:     "customer",
		Name:     "Find Me",
		Email:    "findme@example.com",
	}

	store.CreateUser(user)

	foundUser, err := store.GetUserByUsername("findme")
	if err != nil {
		t.Fatalf("Expected user to be found, got error: %v", err)
	}

	if foundUser.Username != "findme" {
		t.Errorf("Expected username 'findme', got %s", foundUser.Username)
	}
}

func TestInMemoryUserStore_GetUserByUsername_NotFound(t *testing.T) {
	db := CreateTestDB()
	store := jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])

	_, err := store.GetUserByUsername("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent user, got nil")
	}
}

func TestInMemoryUserStore_GetAllUsers(t *testing.T) {
	db := CreateTestDB()
	store := jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])

	store.CreateUser(models.User{Username: "user1", Password: "pass1", Role: "customer", Name: "User 1", Email: "u1@example.com"})
	store.CreateUser(models.User{Username: "user2", Password: "pass2", Role: "admin", Name: "User 2", Email: "u2@example.com"})

	users := store.GetAllUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}