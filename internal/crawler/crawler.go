package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"web-crawler/internal/models"
)

// AnalyzeURL fetches a URL, parses its HTML, and extracts key information.
func AnalyzeURL(pageURL string) (*models.Analysis, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	doc := html.NewTokenizer(strings.NewReader(string(bodyBytes)))
	result := &models.Analysis{
		URL:           pageURL,
		HeadingCounts: make(map[string]int),
	}
	var allLinks []string

	// First Pass: Collect all page metadata and links.
	for {
		tokenType := doc.Next()
		if tokenType == html.ErrorToken {
			if doc.Err() == io.EOF {
				break
			}
			return nil, fmt.Errorf("error parsing HTML: %w", doc.Err())
		}
		
		switch tokenType {
		case html.DoctypeToken:
			doctype := doc.Token().Data
			if strings.Contains(strings.ToLower(doctype), "html") {
				result.HTMLVersion = "HTML5"
			} else {
				result.HTMLVersion = "Older or unknown"
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			token := doc.Token()
			switch token.Data {
			case "title":
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
							continue
						}
						resolvedURL := base.ResolveReference(linkURL)
						allLinks = append(allLinks, resolvedURL.String())
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

	// Second Pass: Concurrently check all collected links for accessibility.
	result.InaccessibleLinks = checkAllLinks(allLinks)

	return result, nil
}

// checkAllLinks uses goroutines to check a list of URLs and returns the ones that are inaccessible.
func checkAllLinks(links []string) []models.InaccessibleLink {
	var wg sync.WaitGroup
	resultsChan := make(chan models.InaccessibleLink, len(links))
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	uniqueLinks := make(map[string]struct{})

	for _, link := range links {
		if _, exists := uniqueLinks[link]; exists {
			continue
		}
		uniqueLinks[link] = struct{}{}
		
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			req, err := http.NewRequest("HEAD", l, nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "web-crawler-bot/1.0")

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 && resp.StatusCode < 600 {
				resultsChan <- models.InaccessibleLink{
					URL:        l,
					StatusCode: resp.StatusCode,
				}
			}
		}(link)
	}

	wg.Wait()
	close(resultsChan)

	var inaccessibleLinks []models.InaccessibleLink
	for res := range resultsChan {
		inaccessibleLinks = append(inaccessibleLinks, res)
	}
	return inaccessibleLinks
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
