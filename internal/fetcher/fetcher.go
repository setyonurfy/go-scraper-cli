package fetcher

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (compatible; GoScraperCLI/1.0)"

type Fetcher struct {
	client    *http.Client
	userAgent string
}

func New(timeout time.Duration, insecureSkipVerify bool) *Fetcher {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	return &Fetcher{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		userAgent: defaultUserAgent,
	}
}

func (f *Fetcher) FetchHTML(targetURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}

	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal melakukan request ke %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status HTTP tidak diharapkan: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca body respons: %w", err)
	}

	return string(body), nil
}
