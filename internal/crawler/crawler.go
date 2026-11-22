package crawler

import (
	"context"
	"sync"
	"time"

	"go-scraper-cli/internal/fetcher"
	"go-scraper-cli/internal/parser"
)

type Config struct {
	Workers            int
	Delay              time.Duration
	Timeout            time.Duration
	MaxDepth           int
	MaxPages           int
	InsecureSkipVerify bool
}

type Outcome struct {
	URL    string
	Result *parser.ScrapeResult
	Err    error
}

func Run(ctx context.Context, urls []string, cfg Config) []Outcome {
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}
	if cfg.MaxDepth < 1 {
		cfg.MaxDepth = 1
	}

	var ticker *time.Ticker
	var rateCh <-chan time.Time
	if cfg.Delay > 0 {
		ticker = time.NewTicker(cfg.Delay)
		defer ticker.Stop()
		rateCh = ticker.C
	}

	visited := make(map[string]bool)
	currentBatch := make([]string, 0, len(urls))
	for _, u := range urls {
		if cfg.MaxPages > 0 && len(visited) >= cfg.MaxPages {
			break
		}
		if u != "" && !visited[u] {
			visited[u] = true
			currentBatch = append(currentBatch, u)
		}
	}

	var allOutcomes []Outcome

	for depth := 1; depth <= cfg.MaxDepth && len(currentBatch) > 0; depth++ {
		if ctx.Err() != nil {
			break
		}

		batchOutcomes := runBatch(ctx, currentBatch, cfg, rateCh)
		allOutcomes = append(allOutcomes, batchOutcomes...)

		if depth < cfg.MaxDepth {
			if cfg.MaxPages > 0 && len(visited) >= cfg.MaxPages {
				break
			}

			nextBatch := make([]string, 0)
			for _, out := range batchOutcomes {
				if out.Err != nil || out.Result == nil {
					continue
				}
				for _, link := range out.Result.Links {
					if cfg.MaxPages > 0 && len(visited) >= cfg.MaxPages {
						break
					}
					if link.Type == "internal" && !visited[link.URL] {
						visited[link.URL] = true
						nextBatch = append(nextBatch, link.URL)
					}
				}
				if cfg.MaxPages > 0 && len(visited) >= cfg.MaxPages {
					break
				}
			}
			currentBatch = nextBatch
		}
	}
	return allOutcomes
}

func runBatch(ctx context.Context, urls []string, cfg Config, rateCh <-chan time.Time) []Outcome {
	jobs := make(chan string, len(urls))
	results := make(chan Outcome, len(urls))

	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f := fetcher.New(cfg.Timeout, cfg.InsecureSkipVerify)

			for {
				select {
				case <-ctx.Done():
					return
				case targetURL, ok := <-jobs:
					if !ok {
						return
					}

					if rateCh != nil {
						select {
						case <-ctx.Done():
							return
						case <-rateCh:
						}
					}

					if ctx.Err() != nil {
						return
					}

					body, err := f.FetchHTML(targetURL)
					if err != nil {
						results <- Outcome{URL: targetURL, Err: err}
						continue
					}

					parsed, err := parser.Parse(body, targetURL)
					if err != nil {
						results <- Outcome{URL: targetURL, Err: err}
						continue
					}

					results <- Outcome{URL: targetURL, Result: parsed}
				}
			}
		}()
	}

	for _, u := range urls {
		select {
		case <-ctx.Done():
			break
		case jobs <- u:
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	outcomes := make([]Outcome, 0, len(urls))
	for out := range results {
		outcomes = append(outcomes, out)
	}
	return outcomes
}
