package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"online-bookstore/auth"
	"online-bookstore/config"
	"online-bookstore/handlers"
	"online-bookstore/middleware"
	"online-bookstore/models"
	"online-bookstore/stores"
	jsonstores "online-bookstore/stores/json"
	sqlstores "online-bookstore/stores/sql"
	"online-bookstore/utils"
	"golang.org/x/time/rate"


)

const PORT = "1010"
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
                log.Printf("User role: %s", claims.Role)
        
        if claims.Role != "admin" {
            http.Error(w, `{"error":"admin access required"}`, http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    }
}
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login" || r.URL.Path == "/auth/register" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		_, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func debugMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("  %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func main() {
	var (
		bookStore   stores.BookStore
		authorStore stores.AuthorStore
		userStore   stores.UserStore
		orderStore  stores.OrderStore
		reportStore stores.ReportStore
	)
	
	var (
		authLimiter = middleware.NewRateLimiter(rate.Every(10*time.Second), 5)    // 5 requests per 10 seconds
		apiLimiter  = middleware.NewRateLimiter(rate.Every(1*time.Second), 20)   // 20 requests per second  
		adminLimiter = middleware.NewRateLimiter(rate.Every(30*time.Second), 10) // 10 requests per 30 seconds
	)
	storeType := config.GetStoreType()
	log.Printf("- Using store type: %s", storeType)

	if storeType == config.SQLStore {
		db := sqlstores.NewSQLiteDB("./bookstore.db")
		bookStore = sqlstores.NewSQLBookStore(db)
		authorStore = sqlstores.NewSQLAuthorStore(db)
		userStore = sqlstores.NewSQLUserStore(db)
		orderStore = sqlstores.NewSQLOrderStore(db, bookStore, userStore)
		reportStore = sqlstores.NewSQLReportStore(db)
	} else {
		db, err := utils.LoadDB("database.json")
		if err != nil {
			log.Fatalf("Failed to load DB: %v", err)
		}
		authorStore = jsonstores.NewInMemoryAuthorStore(db.Authors, db.NextIDs["author"])
		bookStore = jsonstores.NewInMemoryBookStore(db.Books, db.NextIDs["book"], authorStore)
		userStore = jsonstores.NewInMemoryUserStore(db.Users, db.NextIDs["user"])
		orderStore = jsonstores.NewInMemoryOrderStore(db.Orders, db.NextIDs["order"], bookStore, userStore)
		reportStore = jsonstores.NewFileBasedReportStore()
	}

	handlers.SetReportStore(reportStore)

	mux := http.NewServeMux()

	// ==================== API ROUTES ====================
	authHandler := handlers.NewAuthHandler(userStore)
	
	// Auth routes (no auth required)
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }
    debugMiddleware(authLimiter.Limit(authHandler.Login))(w, r)
})

	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }
    debugMiddleware(authLimiter.Limit(authHandler.Register))(w, r)
})
	// ==================== BOOKS ROUTES ====================
bh := handlers.NewBookHandler(bookStore)

// GET /books (list all) - Anyone can read
// POST /books (create) - Admin only
mux.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        bh.ListAll(w, r)
    case http.MethodPost:
    	adminMiddleware(adminLimiter.Limit(middleware.ValidateBook(bh.Create)))(w, r)
    default:
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
})

// GET/PUT/DELETE /books/{id} - Handle individual books
mux.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
    // Extract ID from path: /books/123
    path := r.URL.Path
    if len(path) <= len("/books/") {
        http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
        return
    }
    
    idStr := path[len("/books/"):]
    if strings.HasSuffix(idStr, "/") {
        idStr = strings.TrimSuffix(idStr, "/")
    }
        _, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
        return
    }
    
    ctx := context.WithValue(r.Context(), "id", idStr)
    r = r.WithContext(ctx)
    
    switch r.Method {
    case http.MethodGet:
        bh.Get(w, r)
    case http.MethodPut:
   		 adminMiddleware(adminLimiter.Limit(middleware.ValidateBook(bh.Update)))(w, r)
	case http.MethodDelete:
    	adminMiddleware(adminLimiter.Limit(bh.Delete))(w, r)
    default:
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
})

// ==================== AUTHORS ROUTES ====================
ah := handlers.NewAuthorHandler(authorStore)

// GET /authors (list all) - Anyone can read
// POST /authors (create) - Admin only
mux.HandleFunc("/authors", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        ah.ListAll(w, r)
    case http.MethodPost:
   		 adminMiddleware(adminLimiter.Limit(middleware.ValidateAuthor(ah.Create)))(w, r)
    default:
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
})

// GET/PUT/DELETE /authors/{id} - Handle individual authors
mux.HandleFunc("/authors/", func(w http.ResponseWriter, r *http.Request) {
    // Extract ID from path: /authors/123
    path := r.URL.Path
    if len(path) <= len("/authors/") {
        http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
        return
    }
    
    idStr := path[len("/authors/"):]
    if strings.HasSuffix(idStr, "/") {
        idStr = strings.TrimSuffix(idStr, "/")
    }
    
    _, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
        return
    }
    ctx := context.WithValue(r.Context(), "id", idStr)
    r = r.WithContext(ctx)
    
    switch r.Method {
    case http.MethodGet:
        ah.Get(w, r)
    case http.MethodPut:
   		 adminMiddleware(adminLimiter.Limit(middleware.ValidateAuthor(ah.Update)))(w, r)
	case http.MethodDelete:
   		 adminMiddleware(adminLimiter.Limit(ah.Delete))(w, r)
    default:
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
})
// ==================== Orders ROUTES ====================

oh := handlers.NewOrderHandler(orderStore, bookStore, userStore)
mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
   		 debugMiddleware(authMiddleware(apiLimiter.Limit(oh.ListAll)))(w, r)
    case http.MethodPost:
		handler := authMiddleware(apiLimiter.Limit(middleware.ValidateOrder(oh.Create)))
		debugMiddleware(handler)(w, r)
    default:
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
})
	uh := handlers.NewUserHandler(userStore)
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		debugMiddleware(authMiddleware(apiLimiter.Limit(uh.GetCurrent)))(w, r)	})
	
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		debugMiddleware(authMiddleware(apiLimiter.Limit(uh.ListAll)))(w, r)
	})
	mux.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		debugMiddleware(handlers.ListReports)(w, r)
	})

	// ====================  FILES ====================
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") ||
			strings.HasPrefix(r.URL.Path, "/books") ||
			strings.HasPrefix(r.URL.Path, "/authors") ||
			strings.HasPrefix(r.URL.Path, "/orders") ||
			strings.HasPrefix(r.URL.Path, "/users") ||
			strings.HasPrefix(r.URL.Path, "/reports") {
			http.Error(w, `{"error":"endpoint not found"}`, http.StatusNotFound)
			return
		}
			http.ServeFile(w, r, "./frontend/index.html")
	})

	// ==================== BACKGROUND REPORT GENERATION ====================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(500 * time.Millisecond)
		generateAndSaveReport(ctx, bookStore, orderStore, reportStore)
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		timer := time.NewTimer(time.Until(nextMidnight))
		<-timer.C
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				generateAndSaveReport(ctx, bookStore, orderStore, reportStore)
			case <-ctx.Done():
				return
			}
		}
	}()

	// ==================== SERVER STARTUP ====================
	server := &http.Server{Addr: ":" + PORT, Handler: mux}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		log.Printf(" - Server starting on http://localhost:%s", PORT)
		log.Printf(" - Frontend: http://localhost:%s", PORT)
		log.Printf(":)  API endpoints ready")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
		done <- true
	}()
	<-shutdown
	log.Println(" Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(" Server forced to shutdown: %v", err)
	}

	// Save JSON database if using JSON store
	if storeType == config.JSONStore {
		finalDB := &utils.DB{
			Books:   bookStore.(interface{ GetAllBooks() map[int]models.Book }).GetAllBooks(),
			Authors: authorStore.(interface{ GetAllAuthors() map[int]models.Author }).GetAllAuthors(),
			Users:   userStore.(interface{ GetAllUsers() map[int]models.User }).GetAllUsers(),
			Orders:  orderStore.(interface{ GetAllOrders() map[int]models.Order }).GetAllOrders(),
			NextIDs: map[string]int{
				"book":   bookStore.(interface{ GetNextID() int }).GetNextID(),
				"author": authorStore.(interface{ GetNextID() int }).GetNextID(),
				"user":   userStore.(interface{ GetNextID() int }).GetNextID(),
				"order":  orderStore.(interface{ GetNextID() int }).GetNextID(),
			},
		}
		if err := utils.SaveDB("database.json", finalDB); err != nil {
			log.Printf(":( Failed to save database.json: %v", err)
		} else {
			log.Println(":) Data saved to database.json")
		}
	}

	<-done
	log.Println(" Server stopped")
}

func getPathValue(r *http.Request, key string) string {
	if val := r.Context().Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func generateAndSaveReport(
	ctx context.Context,
	bookStore stores.BookStore,
	orderStore stores.OrderStore,
	reportStore stores.ReportStore,
) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	report, err := stores.GenerateSalesReport(bookStore, orderStore, startTime, endTime)
	if err != nil || report.TotalOrders == 0 {
		log.Println("- No orders in the last 24 hours. Skipping report.")
		return
	}

	if err := reportStore.SaveReport(report); err != nil {
		log.Printf(":( Report save failed: %v", err)
		return
	}

	log.Printf(":) Report saved | Orders: %d, Revenue: $%.2f",
		report.TotalOrders,
		report.TotalRevenue)
}