package scraper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// VividSeatsScraper handles VividSeats web scraping operations
type VividSeatsScraper struct {
	collector *colly.Collector
	baseURL   string
}

// NewVividSeatsScraper creates a new VividSeats scraper instance
func NewVividSeatsScraper() *VividSeatsScraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Set up rate limiting to be respectful
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       1 * time.Second,
	})

	return &VividSeatsScraper{
		collector: c,
		baseURL:   "https://www.vividseats.com",
	}
}

// ScrapeVividSeatsRealMadridTickets scrapes the VividSeats Real Madrid tickets page
func (s *VividSeatsScraper) ScrapeVividSeatsRealMadridTickets() (*ScrapingResult, error) {
	url := "https://www.vividseats.com/real-madrid-tickets--sports-soccer/performer/3053"

	result := &ScrapingResult{
		Events:    []TicketEvent{},
		Timestamp: time.Now(),
		SourceURL: url,
		Source:    "vividseats",
	}

	s.collector.OnHTML("div[data-testid*='production-listing']", func(e *colly.HTMLElement) {
		event := s.parseVividSeatsTicketEvent(e)
		if event != nil {
			result.Events = append(result.Events, *event)
		}
	})

	s.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
	})

	err := s.collector.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	result.Total = len(result.Events)
	return result, nil
}

// parseVividSeatsTicketEvent extracts essential ticket event data from VividSeats HTML element
func (s *VividSeatsScraper) parseVividSeatsTicketEvent(e *colly.HTMLElement) *TicketEvent {
	// Extract link
	link := e.ChildAttr("a.styles_linkContainer__4li3j", "href")
	if link == "" {
		return nil
	}

	// Convert to full URL
	if !strings.HasPrefix(link, "http") {
		link = s.baseURL + link
	}

	// Extract date information from the left column
	day := strings.TrimSpace(e.ChildText("div[data-testid='date-time-left-element'] span.MuiTypography-overline"))
	dateMonth := strings.TrimSpace(e.ChildText("div[data-testid='date-time-left-element'] span.MuiTypography-small-bold"))
	timeStr := strings.TrimSpace(e.ChildText("div[data-testid='date-time-left-element'] span.MuiTypography-caption"))

	// Extract event name
	event := strings.TrimSpace(e.ChildText("span.styles_titleTruncate__XiZ53"))

	// Fix date format - separate year from day if they're concatenated
	formattedDate := s.formatDateWithYear(dateMonth)

	// Combine date and time into single string
	datetime := fmt.Sprintf("%s %s %s", formattedDate, day, timeStr)

	return &TicketEvent{
		DateTime: datetime,
		Event:    event,
		Link:     link,
		Source:   "vividseats",
	}
}

// formatDateWithYear separates the year from the day in date strings like "Jan 182026"
func (s *VividSeatsScraper) formatDateWithYear(dateStr string) string {
	// Look for patterns like "Jan 182026" or "Feb 152026"
	// Split by spaces and check if the last part has 4+ digits at the end
	parts := strings.Fields(dateStr)
	if len(parts) >= 2 {
		dayPart := parts[1]

		// Check if the day part ends with 4 digits (year)
		if len(dayPart) >= 5 {
			// Find where the year starts (last 4 digits)
			for i := len(dayPart) - 4; i >= 0; i-- {
				if i+4 <= len(dayPart) {
					year := dayPart[i : i+4]
					day := dayPart[:i]

					// Check if year is reasonable (2025-2030)
					if year == "2025" || year == "2026" || year == "2027" || year == "2028" || year == "2029" || year == "2030" {
						return fmt.Sprintf("%s %s %s", parts[0], day, year)
					}
				}
			}
		}
	}

	// If no year found, return as is
	return dateStr
}
