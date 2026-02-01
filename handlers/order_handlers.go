// handlers/order_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"online-bookstore/auth"
	"online-bookstore/models"
	"online-bookstore/stores"
	"strconv"
)

type OrderHandler struct {
	store      stores.OrderStore
	bookStore  stores.BookStore
	userStore  stores.UserStore
}

func NewOrderHandler(store stores.OrderStore, bookStore stores.BookStore, userStore stores.UserStore) *OrderHandler {
	return &OrderHandler{
		store:     store,
		bookStore: bookStore,
		userStore: userStore,
	}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ValidateTokenFromRequest(r)
	if err != nil {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return
	}
	var orderReq struct {
		Items      []struct {
			BookID   int `json:"book_id"`
			Quantity int `json:"quantity"`
		} `json:"items"`
		TotalPrice float64 `json:"total_price"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	var validatedItems []models.OrderItem
	totalCalculated := 0.0

	for _, item := range orderReq.Items {
		if item.Quantity <= 0 {
			http.Error(w, `{"error":"quantity must be greater than 0"}`, http.StatusBadRequest)
			return
		}
		book, err := h.bookStore.GetBook(item.BookID)
		if err != nil {
			http.Error(w, `{"error":"book not found"}`, http.StatusNotFound)
			return
		}

		if book.Stock < item.Quantity {
			http.Error(w, `{"error":"insufficient stock for book: `+book.Title+`"}`, http.StatusBadRequest)
			return
		}

		itemTotal := book.Price * float64(item.Quantity)
		totalCalculated += itemTotal

		validatedItems = append(validatedItems, models.OrderItem{
			Book:     book, // Full book data!
			Quantity: item.Quantity,
		})
	}
	if totalCalculated != orderReq.TotalPrice && 
	   (totalCalculated-orderReq.TotalPrice > 0.01 || orderReq.TotalPrice-totalCalculated > 0.01) {
		http.Error(w, `{"error":"total price mismatch"}`, http.StatusBadRequest)
		return
	}
	user, err := h.userStore.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	order := models.Order{
		UserID:     claims.UserID,
		User:       user,
		Items:      validatedItems,
		TotalPrice: totalCalculated,
		Status:     "confirmed",
	}

	createdOrder, err := h.store.CreateOrder(order)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := getPathValue(r, "id")  
    id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	order, err := h.store.GetOrder(id)
	if err != nil {
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	orders := h.store.ListAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}