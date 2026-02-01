package stores

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"online-bookstore/models"
	"online-bookstore/stores"
)

const ReportDir = "output-reports"
const ReportFileLayout = "report_02012006150405"

type FileBasedReportStore struct{}

var _ stores.ReportStore = (*FileBasedReportStore)(nil)

func NewFileBasedReportStore() *FileBasedReportStore {
	os.MkdirAll(ReportDir, os.ModePerm)
	return &FileBasedReportStore{}
}

func (s *FileBasedReportStore) SaveReport(report models.SalesReport) error {
	t := report.Timestamp.Truncate(time.Second)
	filename := t.Format(ReportFileLayout) + ".json"
	path := filepath.Join(ReportDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func (s *FileBasedReportStore) ListReports(start, end time.Time) ([]models.SalesReport, error) {
	files, err := os.ReadDir(ReportDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.SalesReport{}, nil
		}
		return nil, err
	}

	var reports []models.SalesReport
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") || !strings.HasPrefix(f.Name(), "report_") {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".json")
		t, err := time.Parse(ReportFileLayout, name)
		if err != nil {
			continue
		}
		if t.Before(start) || t.After(end) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ReportDir, f.Name()))
		if err != nil {
			continue
		}
		var rep models.SalesReport
		if err := json.Unmarshal(data, &rep); err != nil {
			continue
		}
		reports = append(reports, rep)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.Before(reports[j].Timestamp)
	})

	return reports, nil
}