package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go-scraper-cli/internal/crawler"
	"go-scraper-cli/internal/exporter"
	"go-scraper-cli/internal/parser"
)

type cliConfig struct {
	rawURL       string
	rawURLs      string
	outPath      string
	format       string
	internalOnly bool
	workers      int
	depth        int
	maxPages     int
	delay        time.Duration
	timeout      time.Duration
}

func parseCLI() (cliConfig, error) {
	cfg := cliConfig{}

	flag.StringVar(&cfg.rawURL, "url", "https://quotes.toscrape.com/", "Single target URL to scrape")
	flag.StringVar(&cfg.rawURLs, "urls", "", "Comma-separated target URLs (overrides --url if set)")
	flag.StringVar(&cfg.outPath, "out", "", "Output file path (default: hasil.<format>)")
	flag.StringVar(&cfg.format, "format", "json", "Output format: json, csv, or text")
	flag.BoolVar(&cfg.internalOnly, "internal-only", false, "Filter output to internal links only")
	flag.IntVar(&cfg.workers, "workers", 3, "Number of concurrent workers")
	flag.DurationVar(&cfg.delay, "delay", 300*time.Millisecond, "Rate limit delay between requests (e.g. 300ms, 1s)")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "Timeout per request (e.g. 10s, 30s)")
	flag.IntVar(&cfg.depth, "depth", 1, "Crawl depth (1 = initial URLs only, 2 = follow internal links once)")
	flag.IntVar(&cfg.maxPages, "max-pages", 20, "Maximum total pages to crawl (0 = unlimited)")

	flag.Parse()
	if cfg.maxPages < 0 {
		return cfg, errors.New("flag --max-pages tidak boleh negatif")
	}
	if cfg.depth < 1 {
		return cfg, errors.New("flag --depth minimal 1")
	}

	if strings.TrimSpace(cfg.rawURLs) == "" && strings.TrimSpace(cfg.rawURL) == "" {
		return cfg, errors.New("flag --url atau --urls wajib diisi")
	}

	if cfg.workers < 1 {
		return cfg, errors.New("flag --workers minimal 1")
	}

	cfg.format = strings.ToLower(strings.TrimSpace(cfg.format))
	switch cfg.format {
	case "json", "csv", "text":
	default:
		return cfg, fmt.Errorf("format tidak didukung: %s (pilih json, csv, atau text)", cfg.format)
	}

	if cfg.outPath == "" && (cfg.format == "json" || cfg.format == "csv") {
		cfg.outPath = "hasil." + cfg.format
	}

	return cfg, nil
}

func resolveTargets(cfg cliConfig) []string {
	if strings.TrimSpace(cfg.rawURLs) != "" {
		parts := strings.Split(cfg.rawURLs, ",")
		targets := make([]string, 0, len(parts))
		for _, p := range parts {
			u := strings.TrimSpace(p)
			if u != "" {
				targets = append(targets, u)
			}
		}
		return targets
	}
	return []string{strings.TrimSpace(cfg.rawURL)}
}

func mergeResults(targets []string, outcomes []crawler.Outcome, internalOnly bool) *parser.ScrapeResult {
	merged := &parser.ScrapeResult{
		SourceURL: strings.Join(targets, ","),
		Title:     fmt.Sprintf("Multi-Page Scrape (%d targets)", len(targets)),
		ScrapedAt: time.Now().UTC(),
		Links:     make([]parser.ScrapedLink, 0),
	}

	seen := make(map[string]bool)
	for _, out := range outcomes {
		if out.Err != nil || out.Result == nil {
			continue
		}
		if len(targets) == 1 {
			merged.SourceURL = out.Result.SourceURL
			merged.Title = out.Result.Title
			merged.ScrapedAt = out.Result.ScrapedAt
		}
		for _, link := range out.Result.Links {
			if internalOnly && link.Type != "internal" {
				continue
			}
			dedupKey := link.URL + "|" + link.Type
			if !seen[dedupKey] {
				seen[dedupKey] = true
				merged.Links = append(merged.Links, link)
			}
		}
	}

	return merged
}

func run(ctx context.Context, cfg cliConfig) error {
	targets := resolveTargets(cfg)
	if len(targets) == 0 {
		return errors.New("tidak ada URL target yang valid")
	}

	outcomes := crawler.Run(ctx, targets, crawler.Config{
		Workers:            cfg.workers,
		Delay:              cfg.delay,
		Timeout:            cfg.timeout,
		MaxDepth:           cfg.depth,
		MaxPages:           cfg.maxPages,
		InsecureSkipVerify: false,
	})

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("proses dihentikan: %w", err)
	}

	errCount := 0
	for _, out := range outcomes {
		if out.Err != nil {
			errCount++
			fmt.Fprintf(os.Stderr, "Gagal [%s]: %v\n", out.URL, out.Err)
			continue
		}
		fmt.Printf("Selesai [%s] -> Judul: %q | Link: %d\n", out.URL, out.Result.Title, len(out.Result.Links))
	}

	if errCount == len(outcomes) {
		return errors.New("semua URL target gagal di-scrape")
	}

	finalRes := mergeResults(targets, outcomes, cfg.internalOnly)

	switch cfg.format {
	case "json":
		if err := exporter.SaveJSON(cfg.outPath, finalRes); err != nil {
			return err
		}
		fmt.Printf("Berhasil menyimpan ke %s\n", cfg.outPath)
	case "csv":
		if err := exporter.SaveCSV(cfg.outPath, finalRes); err != nil {
			return err
		}
		fmt.Printf("Berhasil menyimpan ke %s\n", cfg.outPath)
	case "text":
		for _, l := range finalRes.Links {
			fmt.Printf("  [%s] %s\n", l.Type, l.URL)
		}
	}

	return nil
}

func main() {
	cfg, err := parseCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
