package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"online-bookstore/models"
	"online-bookstore/stores"
	jsonstores "online-bookstore/stores/json"
	"online-bookstore/utils"
)
func CreateTestDB() *utils.DB {
	return &utils.DB{
		Books:   make(map[int]models.Book),
		Authors: make(map[int]models.Author),
		Users:   make(map[int]models.User),
		Orders:  make(map[int]models.Order),
		NextIDs: map[string]int{
			"book":   1,
			"author": 1,
			"user":   1,
			"order":  1,
		},
	}
}
func SetupTestStores() (stores.BookStore, stores.AuthorStore, stores.UserStore, stores.OrderStore) {
	db := CreateTestDB()
	authorStore := jsonstores.NewInMemoryAuthorStore(db.Authors, db.NextIDs["author"])
	bookStore := jsonstores.NewInMemoryBookStore(db.Books, db.NextIDs["book"], authorStore)
	userStore := jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])
	orderStore := jsonstores.NewInMemoryOrderStore(db.Orders, db.NextIDs["order"], bookStore, userStore)
	return bookStore, authorStore, userStore, orderStore
}
func NewTestRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}
func NewAuthRequest(method, url, token string, body interface{}) *http.Request {
	req := NewTestRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}
func ExtractTokenFromResponse(response *httptest.ResponseRecorder) (string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		return "", err
	}
	if token, ok := result["token"].(string); ok {
		return token, nil
	}
	return "", nil
}