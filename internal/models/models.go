package models

// AnalysisResult holds the extracted information from a crawled web page.
type AnalysisResult struct {
	URL               string
	PageTitle         string
	HTMLVersion       string
	HeadingCounts     map[string]int
	InternalLinkCount int
	ExternalLinkCount int
	HasLoginForm bool
}

// InaccessibleLink stores details about links that returned an error status.
type InaccessibleLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}
