package stores

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"online-bookstore/models"
	"online-bookstore/stores"
)

type SQLReportStore struct {
	db *sql.DB
}

var _ stores.ReportStore = (*SQLReportStore)(nil)

func NewSQLReportStore(db *sql.DB) *SQLReportStore {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS sales_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		total_revenue REAL NOT NULL,
		total_orders INTEGER NOT NULL,
		top_selling_books TEXT NOT NULL  -- JSON
	);
	`)
	if err != nil {
		panic(fmt.Errorf("failed to create reports table: %w", err))
	}
	return &SQLReportStore{db: db}
}

func (s *SQLReportStore) SaveReport(report models.SalesReport) error {
	topBooksJSON, _ := json.Marshal(report.TopSellingBooks)
	_, err := s.db.Exec(`
		INSERT INTO sales_reports (timestamp, total_revenue, total_orders, top_selling_books)
		VALUES (?, ?, ?, ?)`,
		report.Timestamp.Format("2006-01-02T15:04:05Z"),
		report.TotalRevenue,
		report.TotalOrders,
		string(topBooksJSON))
	return err
}

func (s *SQLReportStore) ListReports(start, end time.Time) ([]models.SalesReport, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, total_revenue, total_orders, top_selling_books
		FROM sales_reports
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC`,
		start.Format("2006-01-02T15:04:05Z"),
		end.Format("2006-01-02T15:04:05Z"))
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []models.SalesReport
	for rows.Next() {
		var r models.SalesReport
		var timestampStr, topBooksJSON string
		err := rows.Scan(&timestampStr, &r.TotalRevenue, &r.TotalOrders, &topBooksJSON)
		if err != nil {
			continue
		}
		r.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", timestampStr)
		_ = json.Unmarshal([]byte(topBooksJSON), &r.TopSellingBooks)
		reports = append(reports, r)
	}
	return reports, nil
}