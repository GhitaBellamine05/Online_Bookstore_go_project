package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "online-bookstore/auth"
    "online-bookstore/models"
    "online-bookstore/stores"
)

type AuthHandler struct {
    userStore stores.UserStore
}
func NewAuthHandler(userStore stores.UserStore) *AuthHandler {
    return &AuthHandler{userStore: userStore}
}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req models.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
        return
    }
    if req.Username == "" || req.Password == "" {
        http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
        return
    }
    user, err := h.userStore.GetUserByUsername(req.Username)
    if err != nil {
        http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
        return
    }
    if !models.CheckPasswordHash(req.Password, user.Password) {
        http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
        return
    }
    resp := models.AuthResponse{
        Token:    token,
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
        Name:     user.Name,
        Email:    user.Email,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
        return
    }
    if len(user.Password) < 8 {
        http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
        return
    }
    if user.Name == "" || user.Email == "" {
        http.Error(w, `{"error":"name and email are required"}`, http.StatusBadRequest)
        return
    }
    if user.Role == "" {
        user.Role = "customer"
    }
    if user.Role != "customer" && user.Role != "admin" {
        http.Error(w, `{"error":"invalid role. Must be 'customer' or 'admin'"}`, http.StatusBadRequest)
        return
    }
    user.CreatedAt = time.Now()
    hashedPassword, err := models.HashPassword(user.Password)
    if err != nil {
        http.Error(w, `{"error":"password hashing failed"}`, http.StatusInternalServerError)
        return
    }
    user.Password = hashedPassword
    createdUser, err := h.userStore.CreateUser(user)
    if err != nil {
        if err.Error() == "username already exists" {
            http.Error(w, `{"error":"username already exists"}`, http.StatusBadRequest)
        } else {
            http.Error(w, `{"error":"user creation failed"}`, http.StatusInternalServerError)
        }
        return
    }
    token, err := auth.GenerateToken(createdUser.ID, createdUser.Username, createdUser.Role)
    if err != nil {
        http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
        return
    }
    resp := models.AuthResponse{
        Token:    token,
        UserID:   createdUser.ID,
        Username: createdUser.Username,
        Role:     createdUser.Role,
        Name:     createdUser.Name,
        Email:    createdUser.Email,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}