package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultUserAgent = "Mozzila/5.0 (compatible; GoScraperCLI/1.0)"

type Fetcher struct {
	client    *http.Client
	userAgent string
}

func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: defaultUserAgent,
	}

}

func (f *Fetcher) FecthHTML(targetURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal melakukan reques ke %st: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status http tidak diharapkan: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca body respons: %w", err)
	}
	return string(body), nil
}
