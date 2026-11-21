package parser

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type ParsedPage struct {
	Title string
	Links []string
}

func NewParsedPage(rawHTML string, baseURLString string) (*ParsedPage, error) {
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, fmt.Errorf("base URL tidak valid (%s): %w", baseURLString, err)
	}

	tokenizer := html.NewTokenizer(strings.NewReader(rawHTML))
	result := &ParsedPage{}
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

				resolvedURL, valid := normalizeURL(baseURL, rawHref)
				if !valid {
					continue
				}

				if !seenLinks[resolvedURL] {
					seenLinks[resolvedURL] = true
					result.Links = append(result.Links, resolvedURL)
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

func getAttr(token html.Token, key string) (string, bool) {
	for _, attr := range token.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func normalizeURL(baseURL *url.URL, rawHref string) (string, bool) {
	cleaned := strings.TrimSpace(rawHref)
	if cleaned == "" || strings.HasPrefix(cleaned, "#") {
		return "", false
	}

	lower := strings.ToLower(cleaned)
	if strings.HasPrefix(lower, "javascript") || strings.HasPrefix(lower, "mailto") || strings.HasPrefix(lower, "tel:") {
		return "", false
	}

	parsedRef, err := url.Parse(cleaned)
	if err != nil {
		return "", false
	}

	resolved := baseURL.ResolveReference(parsedRef)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}

	resolved.Fragment = ""
	return resolved.String(), true
}
