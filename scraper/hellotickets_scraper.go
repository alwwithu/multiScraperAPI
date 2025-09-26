package scraper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Scraper handles web scraping operations
type Scraper struct {
	collector *colly.Collector
	baseURL   string
}

// NewScraper creates a new scraper instance
func NewScraper() *Scraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Set up rate limiting to be respectful
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       1 * time.Second,
	})

	return &Scraper{
		collector: c,
		baseURL:   "https://www.hellotickets.com",
	}
}

// ScrapeRealMadridTickets scrapes the Real Madrid tickets page
func (s *Scraper) ScrapeRealMadridTickets() (*ScrapingResult, error) {
	url := "https://www.hellotickets.com/real-madrid-cf-tickets/p-598?qs=real%20mar"

	result := &ScrapingResult{
		Events:    []TicketEvent{},
		Timestamp: time.Now(),
		SourceURL: url,
		Source:    "hellotickets",
	}

	s.collector.OnHTML("li.performance.performances-list__item", func(e *colly.HTMLElement) {
		event := s.parseTicketEvent(e)
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

// parseTicketEvent extracts essential ticket event data from HTML element
func (s *Scraper) parseTicketEvent(e *colly.HTMLElement) *TicketEvent {
	// Extract link
	link := e.ChildAttr("a.performance__link", "href")
	if link == "" {
		return nil
	}

	// Convert to full URL
	if !strings.HasPrefix(link, "http") {
		link = s.baseURL + link
	}

	// Extract date information
	dateMonth := strings.TrimSpace(e.ChildText(".performance__date-month"))
	day := strings.TrimSpace(e.ChildText(".performance__date-day p:first-child"))
	timeStr := strings.TrimSpace(e.ChildText(".performance__date-day p:last-child"))

	// Extract event name
	event := strings.TrimSpace(e.ChildText(".performance__description__name"))

	// Combine date and time into single string
	datetime := fmt.Sprintf("%s %s %s", dateMonth, day, timeStr)

	return &TicketEvent{
		DateTime: datetime,
		Event:    event,
		Link:     link,
		Source:   "hellotickets",
	}
}
