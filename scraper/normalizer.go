package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hbollon/go-edlib"
)

// TeamNameNormalizer handles team name normalization using AI-powered similarity
type TeamNameNormalizer struct {
	teamMappings        map[string]string
	similarityThreshold float64
}

// NewTeamNameNormalizer creates a new team name normalizer
func NewTeamNameNormalizer() *TeamNameNormalizer {
	return &TeamNameNormalizer{
		teamMappings:        getStandardTeamMappings(),
		similarityThreshold: 0.7, // 70% similarity threshold
	}
}

// getStandardTeamMappings returns a map of common team name variations to standard names
func getStandardTeamMappings() map[string]string {
	return map[string]string{
		// Real Madrid variations
		"real madrid":                "Real Madrid",
		"real madrid cf":             "Real Madrid",
		"real madrid club de fútbol": "Real Madrid",
		"rm":                         "Real Madrid",
		"madrid":                     "Real Madrid",

		// Barcelona variations
		"barcelona":    "Barcelona",
		"fc barcelona": "Barcelona",
		"barça":        "Barcelona",
		"barca":        "Barcelona",
		"fcb":          "Barcelona",

		// Atlético Madrid variations
		"atletico madrid":    "Atlético Madrid",
		"atletico de madrid": "Atlético Madrid",
		"atleti":             "Atlético Madrid",
		"atm":                "Atlético Madrid",

		// Manchester United variations
		"manchester united":    "Manchester United",
		"man utd":              "Manchester United",
		"manchester utd":       "Manchester United",
		"man u":                "Manchester United",
		"manchester united fc": "Manchester United",

		// Liverpool variations
		"liverpool":    "Liverpool",
		"liverpool fc": "Liverpool",
		"lfc":          "Liverpool",

		// Juventus variations
		"juventus":    "Juventus",
		"juventus fc": "Juventus",
		"juve":        "Juventus",

		// Other common teams
		"villarreal":       "Villarreal",
		"villarreal cf":    "Villarreal",
		"getafe":           "Getafe",
		"getafe cf":        "Getafe",
		"valencia":         "Valencia",
		"valencia cf":      "Valencia",
		"sevilla":          "Sevilla",
		"sevilla fc":       "Sevilla",
		"athletic bilbao":  "Athletic Bilbao",
		"athletic club":    "Athletic Bilbao",
		"real betis":       "Real Betis",
		"real sociedad":    "Real Sociedad",
		"rayo vallecano":   "Rayo Vallecano",
		"elche":            "Elche",
		"elche cf":         "Elche",
		"girona":           "Girona",
		"girona fc":        "Girona",
		"celta de vigo":    "Celta de Vigo",
		"celta vigo":       "Celta de Vigo",
		"rc celta de vigo": "Celta de Vigo",
		"deportivo alaves": "Deportivo Alavés",
		"alaves":           "Deportivo Alavés",
		"ca osasuna":       "CA Osasuna",
		"osasuna":          "CA Osasuna",
		"levante ud":       "Levante UD",
		"levante":          "Levante UD",
		"mallorca":         "Mallorca",
		"rcd mallorca":     "Mallorca",
		"rcd espanyol":     "RCD Espanyol",
		"espanyol":         "RCD Espanyol",
		"oviedo":           "Oviedo",
		"real oviedo":      "Oviedo",

		// Champions League teams
		"manchester city":    "Manchester City",
		"manchester city fc": "Manchester City",
		"man city":           "Manchester City",
		"as monaco":          "AS Monaco",
		"monaco":             "AS Monaco",
		"sl benfica":         "SL Benfica",
		"benfica":            "SL Benfica",
		"olympiacos fc":      "Olympiacos FC",
		"olympiacos":         "Olympiacos FC",
		"kairat almaty":      "Kairat Almaty",
		"kairat almaty fc":   "Kairat Almaty",
	}
}

// NormalizeEvent normalizes a ticket event using AI-powered similarity matching
func (n *TeamNameNormalizer) NormalizeEvent(event *TicketEvent) *TicketEvent {
	normalized := &TicketEvent{
		DateTime: n.normalizeDateTime(event.DateTime),
		Event:    n.normalizeEventName(event.Event),
		Link:     event.Link,   // Keep link as is
		Source:   event.Source, // Keep source as is
	}
	return normalized
}

// normalizeEventName normalizes the event name using team name mapping and similarity
func (n *TeamNameNormalizer) normalizeEventName(eventName string) string {
	// Clean up the event name
	cleaned := strings.TrimSpace(eventName)

	// Handle "vs" variations
	cleaned = regexp.MustCompile(`\bvs\.?\b`).ReplaceAllString(cleaned, "vs")
	cleaned = regexp.MustCompile(`\bv\b`).ReplaceAllString(cleaned, "vs")

	// Split by "vs" to get teams
	parts := strings.Split(cleaned, "vs")
	if len(parts) != 2 {
		return cleaned // Return as is if not a standard match format
	}

	homeTeam := strings.TrimSpace(parts[0])
	awayTeam := strings.TrimSpace(parts[1])

	// Normalize team names
	normalizedHome := n.normalizeTeamName(homeTeam)
	normalizedAway := n.normalizeTeamName(awayTeam)

	return fmt.Sprintf("%s vs %s", normalizedHome, normalizedAway)
}

// normalizeTeamName normalizes a team name using mapping and similarity
func (n *TeamNameNormalizer) normalizeTeamName(teamName string) string {
	// Clean the team name
	cleaned := strings.TrimSpace(strings.ToLower(teamName))

	// Remove common suffixes
	cleaned = regexp.MustCompile(`\b(fc|cf|ud|club|de fútbol|de futbol)\b`).ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	// Check direct mapping first
	if normalized, exists := n.teamMappings[cleaned]; exists {
		return normalized
	}

	// Use AI-powered similarity matching
	bestMatch := n.findBestSimilarTeam(cleaned)
	if bestMatch != "" {
		return bestMatch
	}

	// If no match found, return original with proper capitalization
	return strings.Title(teamName)
}

// findBestSimilarTeam finds the best matching team using similarity algorithms
func (n *TeamNameNormalizer) findBestSimilarTeam(teamName string) string {
	bestMatch := ""
	bestScore := 0.0

	for mappedTeam := range n.teamMappings {
		// Try multiple similarity algorithms
		levenshteinScore, _ := edlib.StringsSimilarity(teamName, mappedTeam, edlib.Levenshtein)
		jaroScore, _ := edlib.StringsSimilarity(teamName, mappedTeam, edlib.Jaro)
		jaroWinklerScore, _ := edlib.StringsSimilarity(teamName, mappedTeam, edlib.JaroWinkler)

		// Convert to float64 and use the best score from all algorithms
		maxScore := float64(levenshteinScore)
		if float64(jaroScore) > maxScore {
			maxScore = float64(jaroScore)
		}
		if float64(jaroWinklerScore) > maxScore {
			maxScore = float64(jaroWinklerScore)
		}

		if maxScore > bestScore && maxScore >= n.similarityThreshold {
			bestScore = maxScore
			bestMatch = n.teamMappings[mappedTeam]
		}
	}

	if bestScore >= n.similarityThreshold {
		log.Printf("Normalized '%s' to '%s' (similarity: %.2f)", teamName, bestMatch, bestScore)
		return bestMatch
	}

	return ""
}

// normalizeDateTime normalizes date/time format
func (n *TeamNameNormalizer) normalizeDateTime(dateTime string) string {
	// Clean up the datetime string
	cleaned := strings.TrimSpace(dateTime)

	// Standardize common variations
	cleaned = regexp.MustCompile(`\bvs\.?\b`).ReplaceAllString(cleaned, "vs")
	cleaned = regexp.MustCompile(`\bv\b`).ReplaceAllString(cleaned, "vs")

	return cleaned
}

// NormalizeScrapingResult normalizes all events in a scraping result
func (n *TeamNameNormalizer) NormalizeScrapingResult(result *ScrapingResult) *ScrapingResult {
	normalized := &ScrapingResult{
		Events:    make([]TicketEvent, len(result.Events)),
		Total:     result.Total,
		Timestamp: result.Timestamp,
		SourceURL: result.SourceURL,
		Source:    result.Source,
	}

	for i, event := range result.Events {
		normalized.Events[i] = *n.NormalizeEvent(&event)
	}

	return normalized
}
