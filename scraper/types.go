package scraper

import "time"

// TicketEvent represents a single ticket event with essential information only
type TicketEvent struct {
	DateTime string `json:"datetime"` // e.g., "27 Sep Sat 4:15pm"
	Event    string `json:"event"`    // e.g., "Atlético de Madrid vs. Real Madrid CF"
	Link     string `json:"link"`     // e.g., "/spain/madrid/sports/..."
	Source   string `json:"source"`   // e.g., "hellotickets" or "vividseats"
}

// ScrapingResult contains all scraped events and metadata
type ScrapingResult struct {
	Events    []TicketEvent `json:"events"`
	Total     int           `json:"total"`
	Timestamp time.Time     `json:"timestamp"`
	SourceURL string        `json:"source_url"`
	Source    string        `json:"source"` // "hellotickets" or "vividseats"
}
