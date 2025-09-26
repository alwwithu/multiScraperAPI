package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"normalizer/scraper"

	"github.com/gorilla/mux"
)

// WebServer handles HTTP requests for the web interface
type WebServer struct {
	scraper *scraper.Scraper
	port    string
}

// NewWebServer creates a new web server instance
func NewWebServer(port string) *WebServer {
	return &WebServer{
		scraper: scraper.NewScraper(),
		port:    port,
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/scrape", ws.handleScrape).Methods("GET")
	api.HandleFunc("/health", ws.handleHealth).Methods("GET")

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// CORS middleware
	r.Use(corsMiddleware)

	// Logging middleware
	r.Use(loggingMiddleware)

	fmt.Printf("🚀 Web server starting on http://localhost:%s\n", ws.port)
	fmt.Printf("📱 Frontend available at: http://localhost:%s\n", ws.port)
	fmt.Printf("🔗 API endpoints:\n")
	fmt.Printf("   - GET /api/scrape - Scrape tickets\n")
	fmt.Printf("   - GET /api/health - Health check\n")

	return http.ListenAndServe(":"+ws.port, r)
}

// handleScrape handles the scraping API endpoint
func (ws *WebServer) handleScrape(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	source := query.Get("source")
	if source == "" {
		source = "hellotickets"
	}

	normalize := query.Get("normalize") == "true"
	filter := query.Get("filter")
	dateFrom := query.Get("from")
	dateTo := query.Get("to")

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Scrape tickets based on source
	var result *scraper.ScrapingResult
	var err error

	switch source {
	case "hellotickets":
		result, err = scraper.NewScraper().ScrapeRealMadridTickets()
	case "vividseats":
		result, err = scraper.NewVividSeatsScraper().ScrapeVividSeatsRealMadridTickets()
	case "sport365":
		result, err = scraper.NewSport365Scraper().ScrapeSport365RealMadridMatches()
	case "all":
		// Scrape from all sources
		helloResult, err1 := scraper.NewScraper().ScrapeRealMadridTickets()
		vividResult, err2 := scraper.NewVividSeatsScraper().ScrapeVividSeatsRealMadridTickets()
		sportResult, err3 := scraper.NewSport365Scraper().ScrapeSport365RealMadridMatches()

		if err1 != nil && err2 != nil && err3 != nil {
			http.Error(w, fmt.Sprintf("Failed to scrape from all sources: %v, %v, %v", err1, err2, err3), http.StatusInternalServerError)
			return
		}

		// Combine results
		result = &scraper.ScrapingResult{
			Events:    []scraper.TicketEvent{},
			Timestamp: time.Now(),
			SourceURL: "multiple_sources",
			Source:    "all",
		}

		if helloResult != nil {
			result.Events = append(result.Events, helloResult.Events...)
		}
		if vividResult != nil {
			result.Events = append(result.Events, vividResult.Events...)
		}
		if sportResult != nil {
			result.Events = append(result.Events, sportResult.Events...)
		}
		result.Total = len(result.Events)
	default:
		http.Error(w, "Invalid source. Use: hellotickets, vividseats, sport365, or all", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Scraping failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply normalization if requested
	if normalize {
		normalizer := scraper.NewTeamNameNormalizer()
		result = normalizer.NormalizeScrapingResult(result)
	}

	// Apply filters
	if filter != "" {
		result = result.FilterByKeyword(filter)
	}

	if dateFrom != "" || dateTo != "" {
		var startDate, endDate time.Time
		var parseErr error

		if dateFrom != "" {
			startDate, parseErr = time.Parse("2006-01-02", dateFrom)
			if parseErr != nil {
				http.Error(w, fmt.Sprintf("Invalid from date: %s", dateFrom), http.StatusBadRequest)
				return
			}
		} else {
			startDate = time.Now().AddDate(-1, 0, 0) // 1 year ago
		}

		if dateTo != "" {
			endDate, parseErr = time.Parse("2006-01-02", dateTo)
			if parseErr != nil {
				http.Error(w, fmt.Sprintf("Invalid to date: %s", dateTo), http.StatusBadRequest)
				return
			}
		} else {
			endDate = time.Now().AddDate(2, 0, 0) // 2 years from now
		}

		result = result.FilterByDate(startDate, endDate)
	}

	// Return JSON response
	json.NewEncoder(w).Encode(result)
}

// handleHealth handles the health check endpoint
func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"services": map[string]string{
			"hellotickets": "available",
			"vividseats":   "available",
			"sport365":     "available",
		},
	}

	json.NewEncoder(w).Encode(health)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}

func main() {
	port := flag.String("port", "8080", "Port to run the web server on")
	flag.Parse()

	// Check if we're in the right directory
	if _, err := os.Stat("./web"); os.IsNotExist(err) {
		log.Fatal("Web directory not found. Please run from the project root directory.")
	}

	// Check if web files exist
	webFiles := []string{"index.html", "styles.css", "script.js"}
	for _, file := range webFiles {
		path := filepath.Join("./web", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("Required web file not found: %s", path)
		}
	}

	fmt.Println("🌐 Real Madrid Ticket Scraper Web Interface")
	fmt.Println("==========================================")
	fmt.Printf("Starting web server on port %s...\n", *port)
	fmt.Println()

	// Create and start web server
	server := NewWebServer(*port)
	log.Fatal(server.Start())
}
