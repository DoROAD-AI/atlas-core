// risks.go
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	v1 "github.com/DoROAD-AI/atlas/api/v1"
	"github.com/DoROAD-AI/atlas/types"
	"github.com/gin-gonic/gin"
)

// ----------------------------------------------------------------------------
// DATA STRUCTS
// ----------------------------------------------------------------------------

// RiskData represents the complete country risk data structure.  It's keyed
// by the ISO2 country code (e.g., "AF", "AL").
type RiskData map[string]CountryRiskInfo

// OuterRiskJSON is the top-level structure of the risk JSON file.
type OuterRiskJSON struct {
	Metadata RiskMetadata               `json:"metadata"` // Metadata about the risk data (e.g., generation timestamp).
	Data     map[string]CountryRiskInfo `json:"data"`     // Map of ISO2 country codes to CountryRiskInfo structs.
}

// RiskMetadata holds metadata about the risk data.
type RiskMetadata struct {
	Generated struct {
		Timestamp int64  `json:"timestamp"` // Unix timestamp of when the data was generated.
		Date      string `json:"date"`      // Human-readable date and time of when the data was generated.
	} `json:"generated"`
}

// CountryRiskInfo holds risk information for a specific country.
type CountryRiskInfo struct {
	CountryID           int                 `json:"country-id" example:"1000"`                      // Internal country ID.
	CountryISO          string              `json:"country-iso" example:"AF"`                       // ISO 3166-1 alpha-2 country code.
	CountryEng          string              `json:"country-eng" example:"Afghanistan"`              // English name of the country.
	CountryFra          string              `json:"country-fra" example:"Afghanistan"`              // French name of the country.
	AdvisoryState       int                 `json:"advisory-state" example:"3"`                     // Advisory level (numerical representation).
	DatePublished       RiskDate            `json:"date-published"`                                 // Date and time when the advisory was published.
	HasAdvisoryWarning  int                 `json:"has-advisory-warning" example:"1"`               // Flag indicating if there's an advisory warning (1 = yes, 0 = no).
	HasRegionalAdvisory int                 `json:"has-regional-advisory" example:"0"`              // Flag indicating if there are regional advisories (1 = yes, 0 = no).
	HasContent          int                 `json:"has-content" example:"1"`                        // Flag indicating if there's content associated with the advisory (1 = yes, 0 = no).
	RecentUpdatesType   string              `json:"recent-updates-type" example:"Editorial change"` // Description of the most recent update.
	Eng                 RiskLanguageDetails `json:"eng"`                                            // English-specific risk details.
	Fra                 RiskLanguageDetails `json:"fra"`                                            // French-specific risk details.
}

// RiskDate represents the date and time information for risk advisories.
type RiskDate struct {
	Timestamp int64  `json:"timestamp"` // Unix timestamp of the publication date.
	Date      string `json:"date"`      // Human-readable date and time of the publication.
	ASP       string `json:"asp"`       // ASP.NET-formatted date and time.
}

// RiskLanguageDetails holds language-specific risk information.
type RiskLanguageDetails struct {
	Name          string `json:"name" example:"Afghanistan"`                         // Country name in the specific language.
	URLSlug       string `json:"url-slug" example:"afghanistan"`                     // URL-friendly slug of the country name.
	FriendlyDate  string `json:"friendly-date" example:"January 22, 2025 14:48 EST"` // Human-readable publication date.
	AdvisoryText  string `json:"advisory-text" example:"Avoid all travel"`           // Advisory text (e.g., "Avoid all travel", "Exercise normal security precautions").
	RecentUpdates string `json:"recent-updates" example:"Editorial change"`          // Description of the most recent update.
}

// riskData holds the loaded risk data.
var riskData RiskData

// ----------------------------------------------------------------------------
// LOADING / INITIAL SETUP
// ----------------------------------------------------------------------------

// LoadRiskData loads and parses the risk data from the JSON file.
// This function reads the JSON data from the specified file and unmarshals it
// into the `riskData` global variable.
//
// Parameters:
//   - filename: The path to the JSON file containing the risk data.
//
// Returns:
//   - An error if the file cannot be read or parsed, or if the data is
//     invalid. Returns nil on success.
//
// For enterprise use, this function ensures that the risk data is loaded
// correctly and efficiently, handling potential errors gracefully.
func LoadRiskData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read risk data file: %w", err)
	}

	var outer OuterRiskJSON
	if err := json.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("failed to parse risk data: %w", err)
	}

	if outer.Data == nil {
		return fmt.Errorf("risk data file is missing 'data' field")
	}

	riskData = outer.Data
	return nil
}

// RegisterRiskRoutes registers the risk-related API endpoints.
// This function sets up the routing for all risk-related endpoints under the
// "/v2/risks" path.
//
// Parameters:
//   - r: A pointer to a gin.RouterGroup. This is the router group to which
//     the risk routes will be added.
//
// For enterprise use, this function provides a clear and organized way to
// manage API endpoints, making it easier to maintain and scale the API.
func RegisterRiskRoutes(r *gin.RouterGroup) {
	risks := r.Group("/risks")
	{
		risks.GET("", GetAllRiskData)
		risks.GET("/:countryCode", GetRiskByCountry) // This endpoint will be modified
		risks.GET("/advisory/:level", GetCountriesByAdvisoryLevel)
	}
}

// ----------------------------------------------------------------------------
// HELPER FUNCTIONS
// ----------------------------------------------------------------------------

// getCountryRiskInfo is a helper function to retrieve risk info by country code.
// This function performs a direct lookup in the `riskData` map using the
// provided country code (ISO2). It returns a pointer to the `CountryRiskInfo`
// struct and a boolean indicating whether the country code was found.
//
// Parameters:
//   - countryCode: The ISO2 country code (case-insensitive).
//
// Returns:
//   - A pointer to the CountryRiskInfo struct if found, otherwise nil.
//   - A boolean indicating whether the country code was found (true) or not
//     (false).
func getCountryRiskInfo(countryCode string) (*CountryRiskInfo, bool) {
	countryCode = strings.ToUpper(countryCode)
	info, ok := riskData[countryCode]
	return &info, ok
}

// findCountryCode is a NEW helper function to find the ISO2 code by various identifiers.
func findCountryCode(identifier string) (string, bool) {
	identifier = strings.ToUpper(identifier)

	// 1. Direct lookup (ISO2) - fastest
	if _, ok := riskData[identifier]; ok {
		return identifier, true
	}

	// 2. Lookup in v1.Countries (ISO3, names)
	for _, country := range v1.Countries {
		if strings.ToUpper(country.CCA3) == identifier ||
			strings.ToUpper(country.CCA2) == identifier || //check for iso2 as well
			strings.EqualFold(country.Name.Common, identifier) ||
			strings.EqualFold(country.Name.Official, identifier) {
			// Found a match, return the ISO2 code
			if riskInfo, ok := getCountryRiskInfo(country.CCA2); ok {
				return strings.ToUpper(riskInfo.CountryISO), true
			}

		}
		//check translations
		for _, translation := range country.Translations {
			if strings.EqualFold(translation.Common, identifier) ||
				strings.EqualFold(translation.Official, identifier) {
				if riskInfo, ok := getCountryRiskInfo(country.CCA2); ok {
					return strings.ToUpper(riskInfo.CountryISO), true
				}
			}
		}
	}

	return "", false // Not found
}

// ----------------------------------------------------------------------------
// ROUTE HANDLERS
// ----------------------------------------------------------------------------

// GetAllRiskData handles GET /v2/risks
// @Summary Get all country risk data
// @Description Retrieves risk advisories for all countries.
// @Tags Risks
// @Accept json
// @Produce json
// @Success 200 {object} RiskData
// @Failure 404 {object} types.ErrorResponse
// @Router /risks [get]
// GetAllRiskData retrieves the complete risk data for all countries.
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a comprehensive overview of risk advisories globally, enabling
// large-scale risk assessments and strategic planning.
func GetAllRiskData(c *gin.Context) {
	if len(riskData) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No risk data found or not loaded."})
		return
	}
	c.JSON(http.StatusOK, riskData)
}

// GetRiskByCountry handles GET /v2/risks/:countryCode
// @Summary Get risk data for a specific country
// @Description Retrieves risk advisory information for a given country identifier (ISO2, ISO3, or country name).
// @Tags Risks
// @Accept json
// @Produce json
// @Param countryCode path string true "Country identifier (ISO2, ISO3, or country name)"
// @Success 200 {object} CountryRiskInfo
// @Failure 404 {object} types.ErrorResponse
// @Router /risks/{countryCode} [get]
// GetRiskByCountry retrieves risk information for a specific country,
// supporting lookups by ISO2, ISO3, and country name.
//
// Parameters:
//   - countryCode: The country identifier (e.g., "CA", "CAN", "Canada").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides flexible access to risk data, allowing users to query using
// different identifiers they might have available.
func GetRiskByCountry(c *gin.Context) {
	identifier := c.Param("countryCode")

	// Use the helper function to find the ISO2 code
	countryCode, found := findCountryCode(identifier)
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Risk data not found for this country identifier"})
		return
	}

	riskInfo, _ := getCountryRiskInfo(countryCode) // We know it exists now
	c.JSON(http.StatusOK, riskInfo)
}

// GetCountriesByAdvisoryLevel handles GET /v2/risks/advisory/:level
// @Summary Get countries by advisory level
// @Description Retrieves a list of countries that have a specific advisory level.
// @Tags Risks
// @Accept json
// @Produce json
// @Param level path int true "Advisory level"
// @Success 200 {array} string
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /risks/advisory/{level} [get]
// GetCountriesByAdvisoryLevel retrieves a list of countries with a specific advisory level.
//
// Parameters:
//   - level: The advisory level (integer).
//
// For enterprise, governmental, commercial, and military use, this endpoint
// enables quick identification of countries based on their risk level,
// facilitating risk-based decision-making and resource allocation.  For
// example, a military user might want to quickly find all countries at
// advisory level 3.
func GetCountriesByAdvisoryLevel(c *gin.Context) {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid advisory level. Must be an integer."})
		return
	}

	var countries []string
	for code, info := range riskData {
		if info.AdvisoryState == level {
			countries = append(countries, code)
		}
	}

	if len(countries) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: fmt.Sprintf("No countries found at advisory level %d", level)})
		return
	}

	c.JSON(http.StatusOK, countries)
}
