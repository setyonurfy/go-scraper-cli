package main

import (
	"fmt"
	"os"
	"time"

	"go-scraper-cli/internal/exporter"
	"go-scraper-cli/internal/fetcher"
	"go-scraper-cli/internal/parser"
)

func main() {
	targetURL := "https://quotes.toscrape.com/"
	f := fetcher.New(10*time.Second, false)

	htmlBody, err := f.FetchHTML(targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetch: %v\n", err)
		os.Exit(1)
	}

	result, err := parser.Parse(htmlBody, targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parse: %v\n", err)
		os.Exit(1)
	}

	internalCount := 0
	externalCount := 0
	for _, link := range result.Links {
		switch link.Type {
		case "internal":
			internalCount++
		case "external":
			externalCount++
		}
	}

	fmt.Printf("Judul Halaman : %s\n", result.Title)
	fmt.Printf("Total Link    : %d (Internal: %d, Eksternal: %d)\n", len(result.Links), internalCount, externalCount)

	if err := exporter.SaveJSON("hasil.json", result); err != nil {
		fmt.Fprintf(os.Stderr, "Error save JSON: %v\n", err)
		os.Exit(1)
	}

	if err := exporter.SaveCSV("hasil.csv", result); err != nil {
		fmt.Fprintf(os.Stderr, "Error save CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Berhasil menyimpan ke hasil.json dan hasil.csv")
}
