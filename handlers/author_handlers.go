package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type AuthorHandler struct {
	store stores.AuthorStore
}

func NewAuthorHandler(store stores.AuthorStore) *AuthorHandler {
	return &AuthorHandler{store: store}
}

func (h *AuthorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var author models.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	created, err := h.store.CreateAuthor(author)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *AuthorHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	author, err := h.store.GetAuthor(id)
	if err != nil {
		http.Error(w, `{"error":"author not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(author)
}

func (h *AuthorHandler) Update(w http.ResponseWriter, r *http.Request) {
    idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
    if idStr == "" {
        http.Error(w, `{"error":"author ID is required"}`, http.StatusBadRequest)
        return
    }
    if err != nil {
        http.Error(w, `{"error":"invalid author ID"}`, http.StatusBadRequest)
        return
    }

    var author models.Author
    if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
        http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
        return
    }
    author.ID = id
    updatedAuthor, err := h.store.UpdateAuthor(id, author)
    if err != nil {
        http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedAuthor)
}


func (h *AuthorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)	
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteAuthor(id); err != nil {
		http.Error(w, `{"error":"author not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthorHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	authors := h.store.ListAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}