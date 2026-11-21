package exporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"go-scraper-cli/internal/parser"
)

func SaveJSON(filename string, data *parser.ScrapeResult) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal meng-encode JSON: %w", err)
	}

	payload = append(payload, '\n')

	if err := os.WriteFile(filename, payload, 0644); err != nil {
		return fmt.Errorf("gagal menulis file JSON %s: %w", filename, err)
	}

	return nil
}

func SaveCSV(filename string, data *parser.ScrapeResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("gagal membuat file CSV %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"source_url", "page_title", "url", "type"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("gagal menulis header CSV: %w", err)
	}

	for _, link := range data.Links {
		row := []string{
			data.SourceURL,
			data.Title,
			link.URL,
			link.Type,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("gagal menulis baris CSV: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error saat flushing CSV: %w", err)
	}

	return nil
}
