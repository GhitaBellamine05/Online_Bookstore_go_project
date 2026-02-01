package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"online-bookstore/models"
	"online-bookstore/stores"
	"online-bookstore/auth"
)


type BookHandler struct {
	store stores.BookStore
}

func NewBookHandler(store stores.BookStore) *BookHandler {
	return &BookHandler{store: store}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	created, err := h.store.CreateBook(book)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	book, err := h.store.GetBook(id)
	if err != nil {
		http.Error(w, `{"error":"book not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	updated, err := h.store.UpdateBook(id, book)
	if err != nil {
		http.Error(w, `{"error":"book not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteBook(id); err != nil {
		http.Error(w, `{"error":"book not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	authorIDStr := r.URL.Query().Get("author_id")
	genre := r.URL.Query().Get("genre")

	authorID := 0
	if authorIDStr != "" {
		authorID, _ = strconv.Atoi(authorIDStr)
	}

	criteria := models.SearchCriteria{
		Title:    title,
		AuthorID: authorID,
		Genre:    genre,
	}

	books, err := h.store.SearchBooks(criteria)
	if err != nil {
		http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	books := h.store.ListAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        _, err := auth.ValidateTokenFromRequest(r)
        if err != nil {
            http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}