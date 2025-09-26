package scraper

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// FormatAsTable formats the scraping results as a readable table
func (r *ScrapingResult) FormatAsTable() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintf(w, "DATETIME\tEVENT\tLINK\tSOURCE\n")
	fmt.Fprintf(w, "--------\t-----\t----\t------\n")

	// Data rows
	for _, event := range r.Events {
		// Truncate long fields for better display
		eventName := truncate(event.Event, 50)
		link := truncate(event.Link, 60)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			event.DateTime,
			eventName,
			link,
			event.Source,
		)
	}

	w.Flush()
	return sb.String()
}

// FormatAsJSON formats the scraping results as JSON
func (r *ScrapingResult) FormatAsJSON(indent bool) (string, error) {
	var data []byte
	var err error

	if indent {
		data, err = json.MarshalIndent(r, "", "  ")
	} else {
		data, err = json.Marshal(r)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// SaveToFile saves the results to a file in the specified format
func (r *ScrapingResult) SaveToFile(filename, format string) error {
	var content string
	var err error

	switch strings.ToLower(format) {
	case "json":
		content, err = r.FormatAsJSON(true)
		if err != nil {
			return err
		}
	case "table", "txt":
		content = r.FormatAsTable()
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, table, txt)", format)
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// FilterByKeyword filters events by keyword in event name
func (r *ScrapingResult) FilterByKeyword(keyword string) *ScrapingResult {
	if keyword == "" {
		return r
	}

	filtered := &ScrapingResult{
		Events:    []TicketEvent{},
		Timestamp: r.Timestamp,
		SourceURL: r.SourceURL,
		Source:    r.Source,
	}

	keywordLower := strings.ToLower(keyword)
	for _, event := range r.Events {
		if strings.Contains(strings.ToLower(event.Event), keywordLower) ||
			strings.Contains(strings.ToLower(event.DateTime), keywordLower) ||
			strings.Contains(strings.ToLower(event.Source), keywordLower) {
			filtered.Events = append(filtered.Events, event)
		}
	}

	filtered.Total = len(filtered.Events)
	return filtered
}

// FilterByDate filters events by date range
func (r *ScrapingResult) FilterByDate(startDate, endDate time.Time) *ScrapingResult {
	filtered := &ScrapingResult{
		Events:    []TicketEvent{},
		Timestamp: r.Timestamp,
		SourceURL: r.SourceURL,
		Source:    r.Source,
	}

	for _, event := range r.Events {
		// Parse the datetime string to extract date
		eventDate, err := parseEventDate(event.DateTime)
		if err != nil {
			// If we can't parse the date, include the event
			filtered.Events = append(filtered.Events, event)
			continue
		}

		// Check if event date is within range
		if (eventDate.After(startDate) || eventDate.Equal(startDate)) &&
			(eventDate.Before(endDate) || eventDate.Equal(endDate)) {
			filtered.Events = append(filtered.Events, event)
		}
	}

	filtered.Total = len(filtered.Events)
	return filtered
}

// parseEventDate attempts to parse various date formats from event datetime strings
func parseEventDate(dateTimeStr string) (time.Time, error) {
	// Common date formats to try
	formats := []string{
		"02 Jan Mon 3:04pm", // "27 Sep Sat 4:15pm"
		"Jan 02 Mon 3:04pm", // "Sep 27 Sat 4:15pm"
		"02 Jan 2006",       // "27 Sep 2025"
		"Jan 02 2006",       // "Sep 27 2025"
		"02 Jan",            // "27 Sep"
		"Jan 02",            // "Sep 27"
		"2006-01-02",        // "2025-09-27"
		"02/01/2006",        // "27/09/2025"
		"01/02/2006",        // "09/27/2025"
	}

	// Try each format
	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			// If year is not specified, assume current year or next year
			if t.Year() == 0 {
				now := time.Now()
				if t.Month() < now.Month() || (t.Month() == now.Month() && t.Day() < now.Day()) {
					t = t.AddDate(now.Year()+1, 0, 0)
				} else {
					t = t.AddDate(now.Year(), 0, 0)
				}
			}
			return t, nil
		}
	}

	// If no format matches, return current time
	return time.Now(), fmt.Errorf("unable to parse date: %s", dateTimeStr)
}

// GetSummary returns a summary of the scraping results
func (r *ScrapingResult) GetSummary() string {
	if len(r.Events) == 0 {
		return "No events found"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Scraping Summary:\n")
	fmt.Fprintf(&sb, "================\n")
	fmt.Fprintf(&sb, "Total Events: %d\n", r.Total)
	fmt.Fprintf(&sb, "Source URL: %s\n", r.SourceURL)
	fmt.Fprintf(&sb, "Scraped At: %s\n", r.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&sb, "\n")

	// Date range (simplified - just show first and last events)
	if len(r.Events) > 0 {
		firstEvent := r.Events[0]
		lastEvent := r.Events[len(r.Events)-1]
		fmt.Fprintf(&sb, "Date Range: %s to %s\n",
			firstEvent.DateTime,
			lastEvent.DateTime)
	}

	// Sample events
	fmt.Fprintf(&sb, "\nSample Events:\n")
	limit := 3
	if len(r.Events) < limit {
		limit = len(r.Events)
	}

	for i := 0; i < limit; i++ {
		event := r.Events[i]
		fmt.Fprintf(&sb, "- %s: %s\n",
			event.DateTime, event.Event)
	}

	return sb.String()
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
