# BS – Online Bookstore Application

A full-featured, secure, and scalable online bookstore built with **Go (Golang)** and a modern **JavaScript frontend**. Supports both **JSON file storage** and **SQLite database** backends with seamless switching between them.

---

## Main Features

###  **Authentication & Authorization**
- JWT-based authentication with secure token handling
- Role-based access control (Customer vs Admin)
- Password strength validation
- Rate limiting to prevent brute-force attacks
- Persistent login sessions

###  **Core Functionality**
- **Books Management**: Create, read, update, delete books
- **Authors Management**: Full CRUD operations for authors  
- **Order System**: Place orders, view order history, stock management
- **User Management**: Register customers, admin user management
- **Sales Reports**: Real-time sales analytics (admin only)

###  **Dual Storage Support**
- **JSON Mode**: Simple file-based storage (`database.json`)
- **SQL Mode**: Robust SQLite database (`bookstore.db`)
- **Zero code changes** required to switch between modes
- **Identical API behavior** regardless of storage backend

###  **Modern Frontend**
- Responsive, role-aware user interface
- Seamless navigation without logout/re-login
- Real-time data updates and error handling
- Professional UI with Font Awesome icons
- Built-in form validation 

###  **Security & Production Ready**
- **Rate Limiting**: Protects against DoS and brute force attacks
- **Input Validation**: Server-side validation middleware
- **Graceful Shutdown**: Proper server termination handling
- **Error Handling**: Comprehensive error responses

###  **Quality Assurance**
- Comprehensive unit tests for all handlers and stores
- Complete API documentation
- Cross-browser compatibility

---

##  Getting Started

### Prerequisites
- Go 1.24+
- Node.js (for frontend development, optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone [https://github.com/GhitaBellamine05/Online_Bookstore_go_project.git]
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Run the application**

   **JSON Mode (default):**
   ```bash
   go run main.go
   ```

   **SQL Mode:**
   ```bash
   # Windows PowerShell
   $env:STORE_TYPE="sql"
   go run main.go

   # Linux/Mac
   STORE_TYPE=sql go run main.go
   ```

4. **Access the application**
   - Open your browser at `http://localhost:8080`
   - Register as a customer or create an admin user via API

---

## Storage Modes

### JSON Storage Mode
- **File**: `database.json` (created automatically)
- **Use Case**: Development, small-scale deployments
- **Advantages**: Simple setup, human-readable format
- **Command**: `go run main.go` (default)

### SQL Storage Mode  
- **File**: `bookstore.db` (created automatically)
- **Use Case**: Production, larger datasets, better performance
- **Advantages**: ACID compliance, better concurrency, indexing
- **Command**: `STORE_TYPE=sql go run main.go`

>  **Switch between modes anytime** – your data persists in separate files!

---

##  Default Users

### Admin User (in both sql and json implementation)
```powershell
$body = @{
    name = "admin"
    username = "admin"
    email = "admin@gmail.com"
    password = "123456789"
    role = "admin"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/auth/register" `
                  -Method POST `
                  -Body $body `
                  -ContentType "application/json"
```

### Customer Users
- All registered users are **customers by default**
- Can browse books/authors, place orders, view personal dashboard
- **Cannot** access admin features (books/authors management, reports)

---


if you want to test an admin user , use the following login parameters : 
   username = "admin"
   password = "123456789"

   
##  API Endpoints

### Authentication
- `POST /auth/register` - Register new user
- `POST /auth/login` - Authenticate existing user

### Books
- `GET /books` - List all books (public)
- `GET /books/{id}` - Get specific book (public)  
- `POST /books` - Create book (admin only)
- `PUT /books/{id}` - Update book (admin only)
- `DELETE /books/{id}` - Delete book (admin only)

### Authors
- `GET /authors` - List all authors (public)
- `GET /authors/{id}` - Get specific author (public)
- `POST /authors` - Create author (admin only)
- `PUT /authors/{id}` - Update author (admin only)  
- `DELETE /authors/{id}` - Delete author (admin only)

### Orders
- `GET /orders` - Get user's orders (authenticated)
- `POST /orders` - Create new order (authenticated)

### Users & Reports
- `GET /users/me` - Get current user profile
- `GET /users?role=customer` - Get all customers (admin only)
- `GET /reports` - Get sales reports (admin only)

---

##  Testing

### Run Unit Tests
```bash
go test ./tests/... -v
```

### Test Coverage
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Manual Testing
- **Rate Limiting**: Rapid login attempts should be blocked after 5 requests


To test it, run in terminal : 
```bash
powershell -ExecutionPolicy Bypass -File .\test.ps1
```

- **Role-Based Access**: Customers cannot access admin features
- **Data Persistence**: Restart server and verify data integrity
- **Cross-Mode Compatibility**: Switch between JSON/SQL modes seamlessly

---

##  Project Structure

```
bookwise/
├── main.go                 # Application entry point
├── go.mod                  # Go dependencies
├── frontend/               # Web frontend
│   ├── index.html          # Main HTML page
│   ├── app.js              # JavaScript logic
│   └── style.css           # CSS styles
├── handlers/               # HTTP request handlers
├── stores/                 # Data storage implementations
│   ├── interfaces.go       # Store interfaces
│   ├── json/               # JSON file storage
│   └── sql/                # SQLite database storage
├── models/                 # Data models
├── middleware/             # Request middleware
├── auth/                   # JWT authentication
├── utils/                  # Utility functions
└── tests/                  # Unit tests
```

---

##  Performance & Scalability

- **Rate Limiting**: Prevents abuse and ensures fair usage
- **Efficient Queries**: Optimized database operations
- **Memory Management**: Proper resource cleanup
- **Concurrent Safe**: Thread-safe store implementations


---
