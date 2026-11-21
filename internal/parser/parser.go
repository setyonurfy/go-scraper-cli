package parser

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type ScrapedLink struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type ScrapeResult struct {
	SourceURL string        `json:"source_url"`
	Title     string        `json:"title"`
	ScrapedAt time.Time     `json:"scraped_at"`
	Links     []ScrapedLink `json:"links"`
}

func Parse(rawHTML string, baseURLString string) (*ScrapeResult, error) {
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, fmt.Errorf("base URL tidak valid (%s): %w", baseURLString, err)
	}

	tokenizer := html.NewTokenizer(strings.NewReader(rawHTML))
	result := &ScrapeResult{
		SourceURL: baseURLString,
		ScrapedAt: time.Now().UTC(),
		Links:     make([]ScrapedLink, 0),
	}
	seenLinks := make(map[string]bool)

	inTitle := false
	var titleBuilder strings.Builder

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err().Error() == "EOF" {
				result.Title = strings.TrimSpace(titleBuilder.String())
				return result, nil
			}
			return nil, fmt.Errorf("gagal mem-parsing HTML: %w", tokenizer.Err())

		case html.StartTagToken:
			token := tokenizer.Token()
			switch token.Data {
			case "title":
				inTitle = true
			case "a":
				rawHref, ok := getAttr(token, "href")
				if !ok {
					continue
				}

				resolvedURL, parsedLink, valid := normalizeURL(baseURL, rawHref)
				if !valid {
					continue
				}

				if !seenLinks[resolvedURL] {
					seenLinks[resolvedURL] = true
					linkType := classifyLink(baseURL.Host, parsedLink.Host)
					result.Links = append(result.Links, ScrapedLink{
						URL:  resolvedURL,
						Type: linkType,
					})
				}
			}

		case html.EndTagToken:
			token := tokenizer.Token()
			if token.Data == "title" {
				inTitle = false
			}

		case html.TextToken:
			if inTitle {
				titleBuilder.WriteString(string(tokenizer.Text()))
			}
		}
	}
}

func classifyLink(baseHost, targetHost string) string {
	if strings.EqualFold(baseHost, targetHost) {
		return "internal"
	}
	return "external"
}

func getAttr(token html.Token, key string) (string, bool) {
	for _, attr := range token.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func normalizeURL(baseURL *url.URL, rawHref string) (string, *url.URL, bool) {
	cleaned := strings.TrimSpace(rawHref)
	if cleaned == "" || strings.HasPrefix(cleaned, "#") {
		return "", nil, false
	}

	lower := strings.ToLower(cleaned)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") {
		return "", nil, false
	}

	parsedRef, err := url.Parse(cleaned)
	if err != nil {
		return "", nil, false
	}

	resolved := baseURL.ResolveReference(parsedRef)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", nil, false
	}

	resolved.Fragment = ""

	return resolved.String(), resolved, true
}
