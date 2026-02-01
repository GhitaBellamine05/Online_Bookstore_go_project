package stores

import (
    "time"
    "online-bookstore/models"
)

type BookStore interface {
    CreateBook(book models.Book) (models.Book, error)
    GetBook(id int) (models.Book, error)
    UpdateBook(id int, book models.Book) (models.Book, error)
    DeleteBook(id int) error
    SearchBooks(criteria models.SearchCriteria) ([]models.Book, error)
    ListAll() []models.Book
    ReduceStock(bookID, quantity int) error
    GetStock(bookID int) (int, error)
}

type AuthorStore interface {
    CreateAuthor(a models.Author) (models.Author, error)
    GetAuthor(id int) (models.Author, error)
    UpdateAuthor(id int, a models.Author) (models.Author, error)
    DeleteAuthor(id int) error
    ListAll() []models.Author
}

type UserStore interface {
    CreateUser(user models.User) (models.User, error)
    GetUserByUsername(username string) (models.User, error)
    GetUserByID(id int) (models.User, error)
    ListAll() []models.User
}

type OrderStore interface {
    CreateOrder(o models.Order) (models.Order, error)
    GetOrder(id int) (models.Order, error)
    GetOrdersInTimeRange(start, end time.Time) ([]models.Order, error)
    ListAll() []models.Order
}

type ReportStore interface {
    SaveReport(report models.SalesReport) error
    ListReports(startDate, endDate time.Time) ([]models.SalesReport, error)
}