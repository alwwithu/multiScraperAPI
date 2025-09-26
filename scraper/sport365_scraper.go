package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// Sport365Scraper handles Sport365 web scraping operations using ChromeDP
type Sport365Scraper struct {
	baseURL string
}

// NewSport365Scraper creates a new Sport365 scraper instance
func NewSport365Scraper() *Sport365Scraper {
	return &Sport365Scraper{
		baseURL: "https://www.sport365.com",
	}
}

// ScrapeSport365RealMadridMatches scrapes the Sport365 Real Madrid fixtures page using ChromeDP
func (s *Sport365Scraper) ScrapeSport365RealMadridMatches() (*ScrapingResult, error) {
	url := "https://www.sport365.com/football/team/real-madrid/1-1973#/fixtures"

	result := &ScrapingResult{
		Events:    []TicketEvent{},
		Timestamp: time.Now(),
		SourceURL: url,
		Source:    "sport365",
	}

	log.Printf("Scraping Sport365 with ChromeDP: %s", url)

	// Create context with timeout
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var htmlContent string

	// Run ChromeDP tasks
	err := chromedp.Run(ctx,
		// Navigate to the page
		chromedp.Navigate(url),
		// Wait for the page to load and JavaScript to execute
		chromedp.WaitVisible("a.match-row", chromedp.ByQuery),
		// Wait a bit more for dynamic content to load
		chromedp.Sleep(3*time.Second),
		// Get the full HTML content
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		log.Printf("ChromeDP failed to scrape %s: %v", url, err)
		return result, fmt.Errorf("failed to scrape with ChromeDP: %w", err)
	}

	// Parse the HTML content with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return result, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract match events
	doc.Find("a.match-row").Each(func(i int, sel *goquery.Selection) {
		event := s.parseSport365SelectionEvent(sel)
		if event != nil {
			result.Events = append(result.Events, *event)
		}
	})

	result.Total = len(result.Events)

	if result.Total > 0 {
		log.Printf("Successfully scraped %d events from Sport365", result.Total)
	} else {
		log.Printf("No events found at %s", url)
	}

	return result, nil
}

// parseSport365SelectionEvent parses a goquery selection into a TicketEvent
func (s *Sport365Scraper) parseSport365SelectionEvent(sel *goquery.Selection) *TicketEvent {
	// Extract link
	link, _ := sel.Attr("href")
	if link == "" {
		return nil
	}

	// Convert to full URL
	if !strings.HasPrefix(link, "http") {
		link = s.baseURL + link
	}

	// Extract date from status column
	date := strings.TrimSpace(sel.Find(".match-col.status .status-content").Text())

	// Extract home team name
	homeTeam := strings.TrimSpace(sel.Find(".match-col.home-team .team-name").Text())

	// Extract away team name
	awayTeam := strings.TrimSpace(sel.Find(".match-col.away-team .team-name").Text())

	// Create event name
	event := fmt.Sprintf("%s vs. %s", homeTeam, awayTeam)

	// Format datetime (Sport365 only provides date, no time)
	datetime := date

	return &TicketEvent{
		DateTime: datetime,
		Event:    event,
		Link:     link,
		Source:   "sport365",
	}
}
