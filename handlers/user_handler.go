package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"online-bookstore/auth"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type UserHandler struct {
	store stores.UserStore
}

func NewUserHandler(store stores.UserStore) *UserHandler {
	return &UserHandler{store: store}
}

func (h *UserHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.store.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	users := h.store.ListAll()

	if role != "" {
		filtered := []models.User{}
		for _, user := range users {
			if user.Role == role {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	// Sanitize passwords
	for i := range users {
		users[i].Password = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = getPathValue(r, "id")
	}
	
	if idStr == "" {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID format"}`, http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByID(id)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getPathValue(r *http.Request, key string) string {
	if val := r.Context().Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}