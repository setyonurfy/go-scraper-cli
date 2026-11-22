package exporter

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-scraper-cli/internal/parser"
)

func sampleResult() *parser.ScrapeResult {
	return &parser.ScrapeResult{
		SourceURL: "https://example.com",
		Title:     "Contoh",
		ScrapedAt: time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC),
		Links: []parser.ScrapedLink{
			{URL: "https://example.com/about", Type: "internal"},
			{URL: "https://luar.org", Type: "external"},
		},
	}
}

func TestSaveJSON_And_SaveCSV(t *testing.T) {
	dir := t.TempDir()
	res := sampleResult()

	jsonPath := filepath.Join(dir, "out.json")
	if err := SaveJSON(jsonPath, res); err != nil {
		t.Fatalf("SaveJSON() error = %v", err)
	}

	rawJSON, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("ReadFile(json) error = %v", err)
	}

	var decoded parser.ScrapeResult
	if err := json.Unmarshal(rawJSON, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.Title != res.Title || len(decoded.Links) != 2 {
		t.Errorf("hasil decode JSON tidak sesuai: %+v", decoded)
	}

	csvPath := filepath.Join(dir, "out.csv")
	if err := SaveCSV(csvPath, res); err != nil {
		t.Fatalf("SaveCSV() error = %v", err)
	}

	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("os.Open(csv) error = %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll() error = %v", err)
	}
	// 1 baris header + 2 baris data link
	if len(rows) != 3 {
		t.Fatalf("jumlah baris CSV = %d, want 3 (%+v)", len(rows), rows)
	}
}
