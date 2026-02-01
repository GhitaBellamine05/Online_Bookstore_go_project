package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"online-bookstore/stores"
)

var globalReportStore stores.ReportStore

func SetReportStore(store stores.ReportStore) {
	globalReportStore = store
}

func ListReports(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			http.Error(w, `{"error":"invalid start_date format (YYYY-MM-DD)"}`, http.StatusBadRequest)
			return
		}
	} else {
		start = time.Now().AddDate(-1, 0, 0)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			http.Error(w, `{"error":"invalid end_date format (YYYY-MM-DD)"}`, http.StatusBadRequest)
			return
		}
		end = end.Add(24 * time.Hour)
	} else {
		end = time.Now().Add(24 * time.Hour)
	}

	if globalReportStore == nil {
		http.Error(w, `{"error":"report store not initialized"}`, http.StatusInternalServerError)
		return
	}

	reports, err := globalReportStore.ListReports(start, end)
	if err != nil {
		http.Error(w, `{"error":"failed to list reports"}`, http.StatusInternalServerError)
		return
	}
	totalOrders := 0
	totalRevenue := 0.0
	
	for _, report := range reports {
		totalOrders += report.TotalOrders
		totalRevenue += report.TotalRevenue
	}
	response := map[string]interface{}{
		"total_orders":  totalOrders,
		"total_revenue": totalRevenue,
		"reports":       reports,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}