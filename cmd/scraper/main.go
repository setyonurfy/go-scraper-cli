package main

import (
	"fmt"
	"go-scraper-cli/internal/fetcher"
	"os"
	"time"
)

func main() {
	targetURL := "http://example.com"
	f := fetcher.New(10 * time.Second)

	htmlBody, err := f.FecthHTML(targetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	previewLen := len(htmlBody)
	if previewLen < 300 {
		previewLen = 300
	}

	fmt.Printf("berhasil mengambil HTML dari %s (total %d byte)\n", targetURL, len(htmlBody))
	fmt.Println("---Cupllikan HTML---")
	fmt.Println(htmlBody[:previewLen])
}
