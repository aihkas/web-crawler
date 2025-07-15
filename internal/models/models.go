package models

import "time"

// Analysis represents a single web page analysis record, mirroring the database schema.
type Analysis struct {
	ID                int64      `json:"id"`
	URL               string     `json:"url"`
	Status            string     `json:"status"`
	ErrorMsg          string     `json:"error_msg,omitempty"`
	PageTitle         string     `json:"page_title,omitempty"`
	HTMLVersion       string     `json:"html_version,omitempty"`
	HeadingCounts     map[string]int `json:"heading_counts,omitempty"`
	InternalLinkCount int        `json:"internal_link_count"`
	ExternalLinkCount int        `json:"external_link_count"`
	InaccessibleLinks []InaccessibleLink `json:"inaccessible_links,omitempty"`
	HasLoginForm      bool       `json:"has_login_form"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// InaccessibleLink stores details about links that returned an error status.
type InaccessibleLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}
