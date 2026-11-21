package main

import (
	"fmt"
	"go-scraper-cli/internal/fetcher"
	"go-scraper-cli/internal/parser"
	"os"
	"time"
)

func main() {
	targetURL := "http://example.com"
	f := fetcher.NewFetcher(10 * time.Second)

	htmlBody, err := f.FecthHTML(targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	page, err := parser.NewParsedPage(htmlBody, targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parse: %v\n", err)
	}

	fmt.Printf("Judul Halaman : %s\n", page.Title)
	fmt.Printf("Jumlah Link : %d\n", len(page.Links))
	for i, link := range page.Links {
		fmt.Printf(" %d. %s\n", i+1, link)
	}
}
