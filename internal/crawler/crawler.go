package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"web-crawler/internal/models"
)

// AnalyzeURL fetches a URL, parses its HTML, and extracts key information.
func AnalyzeURL(pageURL string) (*models.AnalysisResult, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse the base URL to help distinguish internal vs. external links.
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Tokenize and parse the HTML body.
	doc := html.NewTokenizer(resp.Body)
	result := &models.AnalysisResult{
		URL:           pageURL,
		HeadingCounts: make(map[string]int),
	}

	for {
		tokenType := doc.Next()

		switch tokenType {
		case html.ErrorToken:
			err := doc.Err()
			if err == io.EOF {
				return result, nil
			}
			return nil, fmt.Errorf("error parsing HTML: %w", err)

		case html.DoctypeToken:
			// Extract HTML version from Doctype.
			doctype := doc.Token().Data
			if strings.Contains(strings.ToLower(doctype), "html") {
				result.HTMLVersion = "HTML5"
			} else {
				result.HTMLVersion = "Older or unknown"
			}


		case html.StartTagToken, html.SelfClosingTagToken:
			token := doc.Token()
			// Find title, headings, links, and forms.
			switch token.Data {
			case "title":
				// The next token should be the text token inside the <title> tag.
				if doc.Next() == html.TextToken {
					result.PageTitle = strings.TrimSpace(doc.Token().Data)
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.HeadingCounts[token.Data]++
			case "a":
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						linkURL, err := url.Parse(attr.Val)
						if err != nil {
							continue // Ignore malformed URLs
						}
						// If the link URL is relative, resolve it against the base URL.
						resolvedURL := base.ResolveReference(linkURL)
						if resolvedURL.Host == base.Host {
							result.InternalLinkCount++
						} else {
							result.ExternalLinkCount++
						}
					}
				}
			case "form":
				if !result.HasLoginForm {
					result.HasLoginForm = checkForPasswordInput(doc)
				}
			}
		}
	}
}

func checkForPasswordInput(t *html.Tokenizer) bool {
	depth := 1
	for depth > 0 {
		tokenType := t.Next()
		if tokenType == html.ErrorToken {
			return false
		}

		token := t.Token()
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			depth++
			if token.Data == "input" {
				isPassword := false
				for _, attr := range token.Attr {
					if attr.Key == "type" && strings.ToLower(attr.Val) == "password" {
						isPassword = true
						break
					}
				}
				if isPassword {
					return true
				}
			}
		case html.EndTagToken:
			depth--
		}
	}
	return false
}
