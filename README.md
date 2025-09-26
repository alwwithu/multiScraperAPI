# Real Madrid Ticket Scraper

A modern web application that scrapes Real Madrid ticket information from multiple sources including HelloTickets, VividSeats, and Sport365. Features a beautiful web interface with AI-powered normalization for clean, standardized data.

## Features

- 🎫 **Scrapes ticket data** from HelloTickets, VividSeats, and Sport365 Real Madrid pages
- 📅 **Extracts comprehensive information**: Date, time, event name, venue, ticket links, prices, and availability
- 🔍 **Filtering capabilities**: Filter by keywords, date ranges
- 📊 **Multiple output formats**: Table view, JSON export
- 💾 **File export**: Save results to JSON or text files
- 🔧 **Easy setup**: Simple one-command startup
- 🔄 **Multi-source support**: Choose between HelloTickets, VividSeats, Sport365, or all sources
- 🌐 **Web interface**: Modern, responsive web interface

## Data Extracted

For each Real Madrid match, the scraper extracts only the essential information:

- **Date & Time**: Match date and time (e.g., "27 Sep Sat 4:15pm")
- **Event**: Match description (e.g., "Atlético de Madrid vs. Real Madrid CF")
- **Link**: Ticket purchase link
- **Source**: Which website the data came from (HelloTickets or VividSeats)

## Installation

### Prerequisites

- Go 1.25

### Run from source

```bash
git clone <repository-url>
cd normalizer
go mod download
go run web_server.go
```

## Usage

### Web Interface

Start the web server and open your browser:

```bash
# Start the web interface
go run web_server.go

# Open http://localhost:8080 in your browser
```

The web interface provides:
- 🎯 Interactive source selection (HelloTickets, VividSeats, Sport365, or all)
- 🤖 AI normalization toggle for clean team names
- 📅 Date range filtering
- 🔍 Real-time keyword filtering
- 📊 Table and JSON views
- 📱 Responsive design for all devices
- ⚡ Real-time progress indicators

### API Usage

You can also use the REST API directly:

```bash
# Scrape from all sources with normalization
curl "http://localhost:8080/api/scrape?source=all&normalize=true"

# Filter by keyword
curl "http://localhost:8080/api/scrape?source=all&filter=Champions"

# Filter by date range
curl "http://localhost:8080/api/scrape?source=all&from=2025-10-01&to=2025-12-31"

# Health check
curl "http://localhost:8080/api/health"
```

### API Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `source` | Data source: hellotickets, vividseats, sport365, or all | `source=all` |
| `normalize` | Enable AI normalization | `normalize=true` |
| `filter` | Filter events by keyword | `filter=Champions` |
| `from` | Filter events from date (YYYY-MM-DD) | `from=2025-10-01` |
| `to` | Filter events to date (YYYY-MM-DD) | `to=2025-12-31` |

## Example Output

### Table Format
```
Ticket Events:
==============
ID       DATE    DAY  TIME    EVENT                                     LINK                              VENUE                          SCARCITY
---      ----    ---  ----    -----                                     ----                              -----                          --------
2263527  27 Sep  Sat  4:15pm  Atlético de Madrid vs. Real Madrid CF     /spain/madrid/sports/atletico...  Riyadh Air Metropolitano •...  This date is an absolu...
2294096  30 Sep  Tue  9:45pm  Kairat Almaty FC vs. Real Madrid CF -...  /kazakhstan/almaty/sports/ka...   Almaty Central Stadium •...    Almost sold out — on...
```

### JSON Format
```json
{
  "events": [
    {
      "id": "2263527",
      "date": "27 Sep",
      "day": "Sat", 
      "time": "4:15pm",
      "datetime": "2025-09-27T16:15:00Z",
      "event": "Atlético de Madrid vs. Real Madrid CF",
      "link": "/spain/madrid/sports/atletico-madrid-tickets/2025-09-27,1615/2263527/2?itemListId=internal&itemListName=internal&itemSublist=internal&performerId=598",
      "venue": "Riyadh Air Metropolitano • Madrid",
      "scarcity": "This date is an absolute best-seller"
    }
  ],
  "total": 1,
  "timestamp": "2025-09-25T12:55:12Z",
  "source_url": "https://www.hellotickets.com/real-madrid-cf-tickets/p-598"
}
```

## Project Structure

```
normalizer/
├── main.go              # Main application entry point
├── go.mod              # Go module file
├── go.sum              # Go dependencies checksum
├── cmd/
│   └── cli.go          # Command-line interface logic
├── scraper/
│   ├── types.go        # Data structures
│   ├── parser.go       # HTML parsing and scraping logic
│   └── formatter.go    # Output formatting and filtering
├── .gitignore          # Git ignore rules
└── README.md           # This file
```

## Dependencies

- **[colly/v2](https://github.com/gocolly/colly)**: Web scraping framework
- **[goquery](https://github.com/PuerkitoBio/goquery)**: HTML parsing (jQuery-like)

## Technical Details

### Web Scraping Approach

The scraper uses the Colly framework to:
1. Visit the Real Madrid tickets page on HelloTickets.com
2. Parse HTML elements with class `li.performance.performances-list__item`
3. Extract data from nested elements using CSS selectors
4. Parse dates and times into structured data
5. Apply rate limiting to be respectful to the server

### HTML Structure Targeted

The scraper looks for this HTML structure:
```html
<li id="2263527" class="performance performances-list__item">
  <a href="/spain/madrid/sports/..." class="performance__link"></a>
  <div class="performance__date-container">
    <p class="performance__date-month">27 Sep</p>
    <span class="performance__date-day">
      <p>Sat</p>
      <p>4:15pm</p>
    </span>
  </div>
  <div class="performance__description">
    <a class="performance__description__name">Atlético de Madrid vs. Real Madrid CF</a>
    <p class="performance__description__venue-city">Riyadh Air Metropolitano • Madrid</p>
    <p class="performance__scarcity-message">This date is an absolute best-seller</p>
  </div>
</li>
```

## Development

### Running Tests

```bash
# Test with sample HTML data
go run main.go test

# Test CLI functionality  
go run main.go -test -verbose
```

### Building

```bash
# Build executable
go build -o normalizer.exe main.go

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o normalizer-linux main.go

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -o normalizer-macos main.go
```

## Legal Notice

This tool is for educational and personal use only. Please respect the website's terms of service and robots.txt. The scraper includes rate limiting to be respectful to the server.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Run tests to ensure they pass
6. Submit a pull request

## License

This project is licensed under the MIT License.
