package stores

import (
    "sort"
    "time"
    "online-bookstore/models"
)

func GenerateSalesReport(bookStore BookStore, orderStore OrderStore, startTime, endTime time.Time) (models.SalesReport, error) {
    orders, err := orderStore.GetOrdersInTimeRange(startTime, endTime)
    if err != nil {
        return models.SalesReport{}, err
    }

    if len(orders) == 0 {
        return models.SalesReport{
            Timestamp:       time.Now(),
            TotalRevenue:    0,
            TotalOrders:     0,
            TopSellingBooks: []models.TopSellingBook{},
        }, nil
    }

    report := models.SalesReport{
        Timestamp:       time.Now(),
        TotalRevenue:    0,
        TotalOrders:     len(orders),
        TopSellingBooks: make([]models.TopSellingBook, 0),
    }

    bookSales := make(map[int]int)
    totalRevenue := 0.0

    for _, order := range orders {
        totalRevenue += order.TotalPrice
        
        for _, item := range order.Items {
            if item.Book.ID <= 0 {
                continue
            }
            book, err := bookStore.GetBook(item.Book.ID)
            if err != nil {
                continue 
            }
            
            bookSales[book.ID] += item.Quantity
        }
    }

    report.TotalRevenue = totalRevenue

    for bookID, quantity := range bookSales {
        book, err := bookStore.GetBook(bookID)
        if err != nil {
            continue 
        }
        
        report.TopSellingBooks = append(report.TopSellingBooks, models.TopSellingBook{
            Book:         book,
            QuantitySold: quantity,
        })
    }
    sort.Slice(report.TopSellingBooks, func(i, j int) bool {
        return report.TopSellingBooks[i].QuantitySold > report.TopSellingBooks[j].QuantitySold
    })

    return report, nil
}