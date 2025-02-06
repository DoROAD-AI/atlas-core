// visa.go
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	v1 "github.com/DoROAD-AI/atlas/api/v1" // Import v1 to access Countries data
	"github.com/DoROAD-AI/atlas/types"
	"github.com/gin-gonic/gin"
)

// ----------------------------------------------------------------------------
// DATA STRUCTS
// ----------------------------------------------------------------------------

// VisaData represents the complete visa requirements data structure.
// Keyed by an upper-case ISO3 country code.  This structure is designed for
// efficient lookup of visa requirements.
//
// For enterprise use, this structure provides a comprehensive and readily
// accessible dataset for all visa-related queries.
type VisaData map[string]CountryVisaInfo

// OuterVisaJSON is the JSON structure you actually have at the top level.
// This structure represents the top-level format of the visa data JSON file,
// including metadata like last update time and total countries covered.
type OuterVisaJSON struct {
	LastUpdated    string                     `json:"last_updated"`    // Timestamp of when the data was last updated.
	TotalCountries int                        `json:"total_countries"` // Total number of countries included in the dataset.
	Countries      map[string]CountryVisaInfo `json:"countries"`       // Map of country ISO3 codes to CountryVisaInfo structs.
}

// CountryVisaInfo holds visa information for a specific country.
// This includes the country's name, a link to its Wikipedia page, various
// country codes, passport ranking information, a visa map, and a list of
// visa requirements for other countries.
type CountryVisaInfo struct {
	Name          string                 `json:"name" example:"Saint Vincent and the Grenadines"`                                                                  // Common name of the country.
	WikiURL       string                 `json:"wiki_url" example:"https://en.wikipedia.org/wiki/Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens"` // URL to the Wikipedia page detailing visa requirements for citizens of this country.
	Codes         CountryCodes           `json:"codes"`                                                                                                            // Various codes associated with the country (ISO2, ISO3, region, subregion).
	PassportIndex PassportIndex          `json:"passport_index"`                                                                                                   // Information about the passport ranking of this country.
	VisaMap       VisaMap                `json:"visa_map"`                                                                                                         // Information about the visa map image, including URL and legend.
	Requirements  []VisaRequirementEntry `json:"requirements"`                                                                                                     // List of visa requirements for citizens of this country traveling to other countries.
}

// CountryCodes represents various country codes.
// This struct holds the ISO 3166-1 alpha-2, ISO 3166-1 alpha-3, region, and
// subregion codes for a country. These codes are used for standardized
// identification and categorization.
type CountryCodes struct {
	ISO2      string `json:"iso2" example:"VC"`             // ISO 3166-1 alpha-2 code (e.g., "VC" for Saint Vincent and the Grenadines).
	ISO3      string `json:"iso3" example:"VCT"`            // ISO 3166-1 alpha-3 code (e.g., "VCT" for Saint Vincent and the Grenadines).
	Region    string `json:"region" example:"Americas"`     // Geographic region the country belongs to (e.g., "Americas").
	Subregion string `json:"subregion" example:"Caribbean"` // Geographic subregion the country belongs to (e.g., "Caribbean").
}

// PassportIndex represents passport ranking information.
// This struct contains data about a country's passport ranking, including
// the number of countries accessible visa-free and the passport's overall
// ranking according to a specific source (e.g., Henley Passport Index).
type PassportIndex struct {
	VisaFreeCount int    `json:"visa_free_count" example:"157"`                  // Number of countries accessible visa-free with this passport.
	Ranking       *int   `json:"ranking" example:"25"`                           // Passport ranking.  This is a pointer to allow for null values (no ranking data).
	RankingSource string `json:"ranking_source" example:"Henley Passport Index"` // Source of the passport ranking (e.g., "Henley Passport Index").
}

// VisaMap represents the visa map information.
// This struct contains a URL to a visual representation (map) of visa
// requirements for a country's citizens and a legend explaining the map's
// color coding.
type VisaMap struct {
	MapURL string            `json:"map_url" example:"https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens.png/800px-Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens.png"` // URL to the visa map image.
	Legend map[string]string `json:"legend"`                                                                                                                                                                                                                    // Legend explaining the color coding of the visa map.
}

// VisaRequirementEntry represents a single visa requirement entry.
// This struct details the visa requirements for citizens of a specific
// country traveling to another country, including the visa requirement type,
// allowed stay, notes, and the destination country's codes and region.
type VisaRequirementEntry struct {
	Country         string `json:"country" example:"Afghanistan"`            // Destination country name.
	VisaRequirement string `json:"visa_requirement" example:"Visa required"` // Visa requirement (e.g., "Visa required", "Visa not required", "e-Visa").
	AllowedStay     string `json:"allowed_stay" example:""`                  // Allowed duration of stay (e.g., "90 days", "30 days").
	Notes           string `json:"notes" example:""`                         // Additional notes about the visa requirement.
	ISO2            string `json:"iso2" example:"AF"`                        // ISO 3166-1 alpha-2 code of the destination country.
	ISO3            string `json:"iso3" example:"AFG"`                       // ISO 3166-1 alpha-3 code of the destination country.
	Region          string `json:"region" example:"Asia"`                    // Region of the destination country.
	Subregion       string `json:"subregion" example:"Southern Asia"`        // Subregion of the destination country.
}

// PassportRank represents a single passport's ranking.  This struct is used
// specifically for the passport ranking endpoint.
type PassportRank struct {
	PassportCode  string `json:"passportCode"`  // The ISO3 code of the passport's country.
	Rank          int    `json:"rank"`          // The passport's rank (1 being the highest).
	VisaFreeCount int    `json:"visaFreeCount"` // The number of countries accessible visa-free.
}

// visaData holds the loaded visa data.  This global variable stores the
// complete visa requirements dataset, making it accessible to all handler
// functions.  It is populated by the `LoadVisaData` function.
var visaData VisaData

// ----------------------------------------------------------------------------
// LOADING / INITIAL SETUP
// ----------------------------------------------------------------------------

// LoadVisaData loads and parses the visa data from the JSON file.
// This function reads the JSON data from the specified file, unmarshals it
// into the `visaData` global variable, and populates the `codeToCCA3` map
// for efficient code lookups.
//
// Parameters:
//   - filename: The path to the JSON file containing the visa data.
//
// Returns:
//   - An error if the file cannot be read or parsed, or if the data is
//     invalid.  Returns nil on success.
//
// For enterprise use, this function ensures that the visa data is loaded
// correctly and efficiently, handling potential errors gracefully.
func LoadVisaData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read visa data file: %w", err)
	}
	var outer OuterVisaJSON
	if err := json.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("failed to parse visa data: %w", err)
	}
	if outer.Countries == nil {
		return fmt.Errorf("visa data file is missing 'countries' field")
	}
	visaData = outer.Countries

	// Add ISO2 and ISO3 to the codeToCCA3 map in handlers.go.
	for _, info := range visaData {
		AddCodesToCCA3Map(info.Codes.ISO2, info.Codes.ISO3)
	}
	return nil
}

// RegisterVisaRoutes registers the visa-related API endpoints.
// This function sets up the routing for all visa-related endpoints under the
// "/v2/visas" path.  It groups related endpoints for better organization and
// readability.
//
// Parameters:
//   - r: A pointer to a gin.RouterGroup. This is the router group to which
//     the visa routes will be added.
//
// For enterprise use, this function provides a clear and organized way to
// manage API endpoints, making it easier to maintain and scale the API.
func RegisterVisaRoutes(r *gin.RouterGroup) {
	visas := r.Group("/visas")
	{
		// Base endpoints
		visas.GET("", GetAllVisaData)
		visas.GET("/search", SearchVisaData)
		visas.GET("/requirements", GetVisaRequirements) // Add this back
		visas.GET("/ranking", GetPassportRanking)
		visas.GET("/common-visa-free", GetCommonVisaFreeDestinations)
		visas.GET("/reciprocal/:countryCode1/:countryCode2", GetReciprocalVisaRequirements)

		// Passport-specific endpoints
		passport := visas.Group("/passport/:passportCode")
		{
			passport.GET("", GetPassportData)
			passport.GET("/all", GetVisaRequirementsForPassport)
			passport.GET("/visa-free", GetVisaFreeCountries)
			passport.GET("/visa-on-arrival", GetVisaOnArrivalCountries)
			passport.GET("/e-visa", GetEVisaCountries)
			passport.GET("/visa-required", GetVisaRequiredCountries)
		}

		// Country-specific endpoints
		visas.GET("/:countryCode", GetVisaRequirementsByCountry)
		visas.GET("/:countryCode/filtered", GetFilteredVisaRequirements)

		// Destination endpoints
		dest := visas.Group("/destination")
		{
			dest.GET("/:destinationCode", GetVisaRequirementsForDestination)
			dest.GET("/:destinationCode/sorted", GetSortedVisaRequirementsForDestination)
		}

		// Comparison endpoints
		visas.GET("/compare", CompareVisaRequirementsCountries)
	}
}

// ----------------------------------------------------------------------------
// HELPER FUNCTIONS
// ----------------------------------------------------------------------------

// getCountryVisaInfo is a helper function to retrieve visa info by country code.
// This function performs a direct lookup in the `visaData` map using the
// provided country code (ISO3).  It returns a pointer to the
// `CountryVisaInfo` struct and a boolean indicating whether the country code
// was found.
//
// Parameters:
//   - countryCode: The ISO3 country code (case-insensitive).
//
// Returns:
//   - A pointer to the CountryVisaInfo struct if found, otherwise nil.
//   - A boolean indicating whether the country code was found (true) or not
//     (false).
//
// This helper function improves code readability and efficiency by
// centralizing the visa data lookup logic.
func getCountryVisaInfo(countryCode string) (*CountryVisaInfo, bool) {
	countryCode = strings.ToUpper(countryCode)
	info, ok := visaData[countryCode] // Direct lookup!
	return &info, ok
}

// parseInt is a simple helper to parse a query param as an integer.
// This function attempts to convert a string to an integer.
//
// Parameters:
//   - s: The string to parse.
//
// Returns:
//   - The parsed integer.
//   - An error if the string cannot be converted to an integer.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// isVisaFreeOrOnArrival is moved to handlers.go

// ----------------------------------------------------------------------------
// ROUTE HANDLERS
// ----------------------------------------------------------------------------

// GetAllVisaData handles GET /v2/visas
// @Summary Get all visa data
// @Description Get complete visa requirement data for all countries
// @Tags Visas
// @Accept json
// @Produce json
// @Success 200 {object} VisaData
// @Failure 404 {object} types.ErrorResponse
// @Router /visas [get]
// GetAllVisaData retrieves the complete visa data for all countries.
// This endpoint returns a JSON object containing visa requirements for all
// countries in the dataset.
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a single source of truth for all visa-related information,
// enabling comprehensive analysis and decision-making.
func GetAllVisaData(c *gin.Context) {
	if len(visaData) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No visa data found or not loaded."})
		return
	}
	c.JSON(http.StatusOK, visaData)
}

// EnhancedVisaRequirement is the new, richer response structure.
// @Description EnhancedVisaRequirement represents the detailed visa requirement between two countries.
type EnhancedVisaRequirement struct {
	From             string `json:"from" example:"USA"`
	To               string `json:"to" example:"DEU"`
	VisaRequirement  string `json:"visa_requirement" example:"Visa not required"` // From visas.json
	AllowedStay      string `json:"allowed_stay" example:"90 days"`               // From visas.json
	Notes            string `json:"notes" example:""`                             // From visas.json
	BasicRequirement string `json:"basic_requirement,omitempty" example:"90"`     // From passports.json (optional)
}

// GetVisaRequirements handles GET /v2/visas/requirements
// @Summary     Get visa requirements between two countries (enhanced)
// @Description Get detailed visa requirements for a passport holder from one country traveling to another, using visas.json data primarily and falling back to passports.json if needed.
// @Tags        Visas
// @Accept      json
// @Produce     json
// @Param       fromCountry query string true "Origin country code (e.g., USA, US, 840, etc.)"
// @Param       toCountry   query string true "Destination country code (e.g., DEU, DE, 276, etc.)"
// @Success     200 {object} EnhancedVisaRequirement
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /visas/requirements [get]
func GetVisaRequirements(c *gin.Context) {
	fromCountryInput := strings.ToUpper(c.Query("fromCountry"))
	toCountryInput := strings.ToUpper(c.Query("toCountry"))

	if fromCountryInput == "" || toCountryInput == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "fromCountry and toCountry query parameters are required"})
		return
	}

	// Get CCA3 codes for both countries
	fromCountryCCA3, ok := codeToCCA3[fromCountryInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: fmt.Sprintf("Invalid fromCountry code: %s", fromCountryInput)})
		return
	}
	toCountryCCA3, ok := codeToCCA3[toCountryInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: fmt.Sprintf("Invalid toCountry code: %s", toCountryInput)})
		return
	}

	// --- 1. Try to get detailed info from visaData ---
	fromCountryInfo, fromFound := visaData[fromCountryCCA3]
	if !fromFound {
		// Fallback to basic passport data if detailed info not found
		fallbackToPassportData(c, fromCountryCCA3, toCountryCCA3)
		return
	}

	// Search through requirements for the destination country
	for _, req := range fromCountryInfo.Requirements {
		if strings.EqualFold(req.ISO3, toCountryCCA3) {
			// Found detailed info! Return it.
			c.JSON(http.StatusOK, EnhancedVisaRequirement{
				From:            fromCountryCCA3,
				To:              toCountryCCA3,
				VisaRequirement: req.VisaRequirement,
				AllowedStay:     req.AllowedStay,
				Notes:           req.Notes,
			})
			return
		}
	}

	// If we get here, we didn't find detailed requirements, fall back to basic passport data
	fallbackToPassportData(c, fromCountryCCA3, toCountryCCA3)
	return
}

// fallbackToPassportData handles the fallback to basic passport data when detailed visa info isn't found
func fallbackToPassportData(c *gin.Context, fromCountryCCA3, toCountryCCA3 string) {
	visaRules, ok := Passports[fromCountryCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: fmt.Sprintf("Passport data not found for origin country: %s", fromCountryCCA3)})
		return
	}

	requirement, ok := visaRules[toCountryCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: fmt.Sprintf("Visa requirement data not found for this country pair: %s to %s", fromCountryCCA3, toCountryCCA3)})
		return
	}

	// Return basic info from passports.json
	c.JSON(http.StatusOK, EnhancedVisaRequirement{
		From:             fromCountryCCA3,
		To:               toCountryCCA3,
		BasicRequirement: requirement,
	})
}

// GetVisaRequirementsByCountry handles GET /v2/visas/{countryCode}
// GetVisaRequirementsByCountry retrieves visa information for a specific country.
// This endpoint returns detailed visa requirements for a given country code
// (ISO3).
//
// Parameters:
//   - countryCode: The ISO3 country code (e.g., "USA").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// allows for quick retrieval of visa requirements for a specific country,
// facilitating travel planning and compliance checks.
func GetVisaRequirementsByCountry(c *gin.Context) {
	countryCode := c.Param("countryCode")
	visaInfo, found := getCountryVisaInfo(countryCode)
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirements not found for this country code"})
		return
	}
	c.JSON(http.StatusOK, visaInfo)
}

// filterVisaRequirements applies filtering to a slice of VisaRequirementEntry.
// This function filters a list of visa requirement entries based on the
// provided query parameters.  It supports filtering by visa requirement type,
// allowed stay, notes, region, subregion, and destination country.
//
// Parameters:
//   - requirements: A slice of VisaRequirementEntry structs to filter.
//   - filters: A url.Values object containing the query parameters.
//
// Returns:
//   - A new slice of VisaRequirementEntry structs containing only the entries
//     that match the filter criteria.
//
// This helper function is used internally by the
// `GetFilteredVisaRequirements` handler to provide flexible filtering
// capabilities.
func filterVisaRequirements(requirements []VisaRequirementEntry, filters url.Values) []VisaRequirementEntry {
	var filtered []VisaRequirementEntry

	for _, req := range requirements {
		match := true

		// Apply each filter, if provided
		if visaReq := filters.Get("visa_requirement"); visaReq != "" && !strings.EqualFold(req.VisaRequirement, visaReq) {
			match = false
		}
		if allowedStay := filters.Get("allowed_stay"); allowedStay != "" && !strings.EqualFold(req.AllowedStay, allowedStay) {
			match = false
		}
		if notes := filters.Get("notes"); notes != "" {
			if !strings.Contains(strings.ToLower(req.Notes), strings.ToLower(notes)) {
				match = false
			}
		}
		if region := filters.Get("region"); region != "" && !strings.EqualFold(req.Region, region) {
			match = false
		}
		if subregion := filters.Get("subregion"); subregion != "" && !strings.EqualFold(req.Subregion, subregion) {
			match = false
		}
		if destination := filters.Get("destination"); destination != "" {
			destination = strings.ToUpper(destination)
			if strings.ToUpper(req.ISO2) != destination && strings.ToUpper(req.ISO3) != destination {
				match = false
			}
		}

		if match {
			filtered = append(filtered, req)
		}
	}

	return filtered
}

// GetFilteredVisaRequirements handles GET /v2/visas/{countryCode}/filtered
// GetFilteredVisaRequirements retrieves filtered visa requirements for a specific country.
// This endpoint allows for filtering visa requirements based on various
// criteria, such as visa requirement type, allowed stay, notes, region,
// subregion, and destination country.
//
// Parameters:
//   - countryCode: The ISO3 country code (e.g., "USA").
//   - Query parameters:
//   - visa_requirement: Filter by visa requirement type (e.g., "Visa required").
//   - allowed_stay: Filter by allowed stay duration (e.g., "90 days").
//   - notes: Filter by notes (case-insensitive substring match).
//   - region: Filter by region (e.g., "Europe").
//   - subregion: Filter by subregion (e.g., "Western Europe").
//   - destination: Filter by destination country code (ISO2 or ISO3).
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a powerful way to filter visa requirements based on specific
// needs, enabling targeted analysis and reporting.
func GetFilteredVisaRequirements(c *gin.Context) {
	countryCode := c.Param("countryCode")
	visaInfo, found := getCountryVisaInfo(countryCode)
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirements not found for this country code"})
		return
	}

	filteredRequirements := filterVisaRequirements(visaInfo.Requirements, c.Request.URL.Query())
	c.JSON(http.StatusOK, filteredRequirements)
}

// GetPassportRanking handles GET /v2/visas/ranking
// @Summary Get passport rankings
// @Description Get global passport rankings based on visa-free access
// @Tags Visas
// @Accept json
// @Produce json
// @Success 200 {array} PassportRank
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/ranking [get]
// GetPassportRanking retrieves global passport rankings based on visa-free access.
// This endpoint returns a list of passport rankings, sorted by the number of
// countries accessible visa-free.
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a global perspective on passport power, enabling strategic
// analysis of international mobility and access.
func GetPassportRanking(c *gin.Context) {

	passportRanks := make(map[string]int)

	for passportCode, visaRules := range Passports { // Uses Passports from handlers.go
		visaFreeCount := 0
		for _, requirement := range visaRules {
			if isVisaFreeOrSimilar(requirement) { // Use the helper function!
				visaFreeCount++
			}
		}
		passportRanks[passportCode] = visaFreeCount
	}

	// Convert map to slice for sorting
	var ranks []PassportRank
	for code, count := range passportRanks {
		ranks = append(ranks, PassportRank{PassportCode: code, VisaFreeCount: count})
	}

	// Sort by visa-free count in descending order
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].VisaFreeCount > ranks[j].VisaFreeCount
	})

	// Assign ranks
	for i := 0; i < len(ranks); i++ {
		if i > 0 && ranks[i].VisaFreeCount != ranks[i-1].VisaFreeCount {
			ranks[i].Rank = i + 1
		} else if i == 0 {
			ranks[i].Rank = 1
		} else {
			ranks[i].Rank = ranks[i-1].Rank
		}
	}

	c.JSON(http.StatusOK, ranks)
}

// CompareVisaRequirementsCountries handles GET /v2/visas/compare
// CompareVisaRequirementsCountries compares visa requirements between two countries.
// This endpoint compares visa requirements for travel between two specified
// countries, providing information on the visa requirements in both
// directions, common countries accessible visa-free by both countries, and
// passport ranking information for both countries.
//
// Parameters:
//   - country1: The ISO3 country code of the first country (e.g., "USA").
//   - country2: The ISO3 country code of the second country (e.g., "CAN").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// facilitates comparative analysis of visa requirements, enabling strategic
// planning and risk assessment for international travel.
func CompareVisaRequirementsCountries(c *gin.Context) {
	countryCode1 := c.Query("country1")
	countryCode2 := c.Query("country2")

	if countryCode1 == "" || countryCode2 == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Both country1 and country2 parameters are required"})
		return
	}

	visaInfo1, found1 := getCountryVisaInfo(countryCode1)
	visaInfo2, found2 := getCountryVisaInfo(countryCode2)

	if !found1 || !found2 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirements not found for one or both countries"})
		return
	}

	result := VisaComparisonResult{
		Country1:              visaInfo1.Name,
		Country2:              visaInfo2.Name,
		Requirements:          make(map[string]string),
		CommonAccess:          []CommonAccessResult{},
		Country1PassportIndex: visaInfo1.PassportIndex,
		Country2PassportIndex: visaInfo2.PassportIndex,
	}

	// (1) Requirement for country1 -> country2
	for _, req := range visaInfo1.Requirements {
		if strings.ToUpper(req.ISO3) == strings.ToUpper(visaInfo2.Codes.ISO3) ||
			strings.ToUpper(req.ISO2) == strings.ToUpper(visaInfo2.Codes.ISO2) {
			result.Requirements[fmt.Sprintf("%s_to_%s", visaInfo1.Codes.ISO3, visaInfo2.Codes.ISO3)] = req.VisaRequirement
			break
		}
	}

	// (2) Requirement for country2 -> country1
	for _, req := range visaInfo2.Requirements {
		if strings.ToUpper(req.ISO3) == strings.ToUpper(visaInfo1.Codes.ISO3) ||
			strings.ToUpper(req.ISO2) == strings.ToUpper(visaInfo1.Codes.ISO2) {
			result.Requirements[fmt.Sprintf("%s_to_%s", visaInfo2.Codes.ISO3, visaInfo1.Codes.ISO3)] = req.VisaRequirement
			break
		}
	}

	// (3) Find common access countries
	accessMap1 := make(map[string]string) // ISO3 -> Requirement
	for _, req := range visaInfo1.Requirements {
		accessMap1[strings.ToUpper(req.ISO3)] = req.VisaRequirement
	}
	for _, req := range visaInfo2.Requirements {
		key := strings.ToUpper(req.ISO3)
		if req1, ok := accessMap1[key]; ok {
			// Use the isVisaFreeOrSimilar helper function from handlers.go
			if isVisaFreeOrSimilar(req1) && isVisaFreeOrSimilar(req.VisaRequirement) {
				// Try to find the actual country name from v1 data or fallback
				countryName := req.Country
				for _, cData := range v1.Countries {
					if strings.ToUpper(cData.CCA3) == key {
						countryName = cData.Name.Common
						break
					}
				}
				result.CommonAccess = append(result.CommonAccess, CommonAccessResult{
					CountryCode:  key,
					CountryName:  countryName,
					Requirement1: req1,
					Requirement2: req.VisaRequirement,
				})
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

// VisaComparisonResult represents the result of comparing visa requirements.
// This struct holds the results of the visa comparison between two countries,
// including the visa requirements in both directions, a list of countries
// accessible visa-free by both countries, and passport ranking information
// for both countries.
type VisaComparisonResult struct {
	Country1              string               `json:"country1" example:"USA"`  // Name of the first country.
	Country2              string               `json:"country2" example:"CAN"`  // Name of the second country.
	Requirements          map[string]string    `json:"requirements"`            // Map of visa requirements, keyed by "country1_to_country2" and "country2_to_country1".
	CommonAccess          []CommonAccessResult `json:"common_access"`           // List of countries accessible visa-free by both countries.
	Country1PassportIndex PassportIndex        `json:"country1_passport_index"` // Passport ranking information for the first country.
	Country2PassportIndex PassportIndex        `json:"country2_passport_index"` // Passport ranking information for the second country.
}

// CommonAccessResult represents a country accessible by both compared countries.
// This struct represents a single country that is accessible visa-free (or
// with similar ease of access) by citizens of both countries being compared.
type CommonAccessResult struct {
	CountryCode  string `json:"country_code" example:"MEX"`                // ISO3 code of the common access country.
	CountryName  string `json:"country_name" example:"Mexico"`             // Common name of the common access country.
	Requirement1 string `json:"requirement_1" example:"Visa not required"` // Visa requirement for the first country's citizens.
	Requirement2 string `json:"requirement_2" example:"Visa not required"` // Visa requirement for the second country's citizens.
}

// GetVisaRequirementsForDestination handles GET /v2/visas/destination/{destinationCode}
// @Summary Get visa requirements for destination
// @Description Get visa requirements for all passports visiting a specific destination
// @Tags Visas
// @Accept json
// @Produce json
// @Param destinationCode path string true "Destination country code"
// @Success 200 {object} VisaDestinationInfo
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/destination/{destinationCode} [get]
// GetVisaRequirementsForDestination retrieves visa requirements for a specific destination country.
// This endpoint returns visa requirements for all passports visiting the
// specified destination country.
//
// Parameters:
//   - destinationCode: The ISO3 country code of the destination country (e.g., "FRA").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a comprehensive view of visa requirements for a specific
// destination, enabling efficient travel planning and compliance checks for
// travelers from all origins.
func GetVisaRequirementsForDestination(c *gin.Context) {
	destinationCode := c.Param("destinationCode")
	destinationInfo, destinationFound := getCountryVisaInfo(destinationCode)

	if !destinationFound {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirements not found for this destination country code"})
		return
	}

	destinationResult := VisaDestinationInfo{
		DestinationCountry: destinationInfo.Name,
		Requirements:       []DestinationVisaRequirement{},
	}

	// Iterate through *all* countries' visa data
	for _, sourceCountryInfo := range visaData {
		for _, req := range sourceCountryInfo.Requirements {
			if strings.EqualFold(req.ISO3, destinationInfo.Codes.ISO3) ||
				strings.EqualFold(req.ISO2, destinationInfo.Codes.ISO2) {
				destinationResult.Requirements = append(destinationResult.Requirements, DestinationVisaRequirement{
					PassportCountry: sourceCountryInfo.Name,
					VisaRequirement: req.VisaRequirement,
					AllowedStay:     req.AllowedStay,
					Notes:           req.Notes,
					ISO2:            sourceCountryInfo.Codes.ISO2,
					ISO3:            sourceCountryInfo.Codes.ISO3,
				})
				break // Found the requirement, skip searching further for this source
			}
		}
	}
	c.JSON(http.StatusOK, destinationResult)
}

// VisaDestinationInfo represents the visa requirements for visiting a specific destination.
// This struct holds the visa requirements for all passports visiting a
// specific destination country.
type VisaDestinationInfo struct {
	DestinationCountry string                       `json:"destination_country" example:"France"` // Name of the destination country.
	Requirements       []DestinationVisaRequirement `json:"requirements"`                         // List of visa requirements for all passports visiting the destination.
}

// DestinationVisaRequirement represents the visa requirement for a specific passport to visit that destination.
// This struct represents a single visa requirement entry for a specific
// passport country visiting the destination country.
type DestinationVisaRequirement struct {
	PassportCountry string `json:"passport_country" example:"United States"`     // Name of the passport country.
	VisaRequirement string `json:"visa_requirement" example:"Visa not required"` // Visa requirement (e.g., "Visa required", "Visa not required", "e-Visa").
	AllowedStay     string `json:"allowed_stay" example:"90 days"`               // Allowed duration of stay.
	Notes           string `json:"notes" example:""`                             // Additional notes about the visa requirement.
	ISO2            string `json:"iso2" example:"US"`                            // ISO 3166-1 alpha-2 code of the passport country.
	ISO3            string `json:"iso3" example:"USA"`                           // ISO 3166-1 alpha-3 code of the passport country.
}

// GetSortedVisaRequirementsForDestination handles GET /v2/visas/destination/{destinationCode}/sorted
// @Summary Get sorted visa requirements for destination
// @Description Get sorted visa requirements for all passports visiting a specific destination
// @Tags Visas
// @Accept json
// @Produce json
// @Param destinationCode path string true "Destination country code"
// @Param sort_by query string true "Sort field (passport_country, visa_requirement, allowed_stay, iso2, iso3)"
// @Success 200 {object} SortedVisaDestinationInfo
// @Failure 400,404 {object} types.ErrorResponse
// @Router /visas/destination/{destinationCode}/sorted [get]
// GetSortedVisaRequirementsForDestination retrieves sorted visa requirements for a specific destination country.
// This endpoint returns visa requirements for all passports visiting the
// specified destination country, sorted by a specified field.
//
// Parameters:
//   - destinationCode: The ISO3 country code of the destination country (e.g., "FRA").
//   - sort_by: The field to sort by (e.g., "passport_country", "visa_requirement", "allowed_stay", "iso2", "iso3").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a flexible way to view visa requirements for a specific
// destination, sorted according to specific needs, enabling easier analysis
// and reporting.
func GetSortedVisaRequirementsForDestination(c *gin.Context) {
	destinationCode := c.Param("destinationCode")
	sortBy := c.Query("sort_by")

	if sortBy == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "sort_by parameter is required"})
		return
	}

	destinationInfo, destinationFound := getCountryVisaInfo(destinationCode)
	if !destinationFound {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirements not found for this destination country code"})
		return
	}

	destinationResult := SortedVisaDestinationInfo{
		DestinationCountry: destinationInfo.Name,
		Requirements:       []DestinationVisaRequirement{},
		SortedBy:           sortBy,
	}

	for _, sourceCountryInfo := range visaData {
		for _, req := range sourceCountryInfo.Requirements {
			if strings.EqualFold(req.ISO3, destinationInfo.Codes.ISO3) ||
				strings.EqualFold(req.ISO2, destinationInfo.Codes.ISO2) {
				destinationResult.Requirements = append(destinationResult.Requirements, DestinationVisaRequirement{
					PassportCountry: sourceCountryInfo.Name,
					VisaRequirement: req.VisaRequirement,
					AllowedStay:     req.AllowedStay,
					Notes:           req.Notes,
					ISO2:            sourceCountryInfo.Codes.ISO2,
					ISO3:            sourceCountryInfo.Codes.ISO3,
				})
				break
			}
		}
	}

	// Perform sorting based on the 'sortBy' field
	sort.Slice(destinationResult.Requirements, func(i, j int) bool {
		switch strings.ToLower(sortBy) {
		case "passport_country":
			return destinationResult.Requirements[i].PassportCountry < destinationResult.Requirements[j].PassportCountry
		case "visa_requirement":
			return destinationResult.Requirements[i].VisaRequirement < destinationResult.Requirements[j].VisaRequirement
		case "allowed_stay":
			return destinationResult.Requirements[i].AllowedStay < destinationResult.Requirements[j].AllowedStay
		case "iso2":
			return destinationResult.Requirements[i].ISO2 < destinationResult.Requirements[j].ISO2
		case "iso3":
			return destinationResult.Requirements[i].ISO3 < destinationResult.Requirements[j].ISO3
		default:
			// If invalid sort field, we do no special ordering (or you could fallback to .PassportCountry).
			return false
		}
	})

	c.JSON(http.StatusOK, destinationResult)
}

// SortedVisaDestinationInfo represents the sorted visa requirements for a destination.
// This struct holds the sorted visa requirements for all passports visiting a
// specific destination country, along with the field by which the data is
// sorted.
type SortedVisaDestinationInfo struct {
	DestinationCountry string                       `json:"destination_country" example:"France"` // Name of the destination country.
	Requirements       []DestinationVisaRequirement `json:"requirements"`                         // List of visa requirements for all passports visiting the destination, sorted.
	SortedBy           string                       `json:"sorted_by" example:"visa_requirement"` // Field by which the requirements are sorted.
}

// ----------------------------------------------------------------------------
// ADVANCED "SEARCH" ENDPOINT
// ----------------------------------------------------------------------------

// SearchVisaData handles GET /v2/visas/search
// SearchVisaData provides an advanced search capability for visa data.
// This endpoint allows for filtering, sorting, and paginating visa data based
// on various criteria.
//
// Parameters:
//   - Query parameters:
//   - name: Filter by country name (case-insensitive substring match).
//   - region: Filter by region (case-insensitive).
//   - subregion: Filter by subregion (case-insensitive).
//   - minVisaFree: Filter by minimum number of visa-free countries accessible.
//   - sortBy: Field to sort by ("name", "region", "visa_free_count").
//   - sortOrder: Sort order ("asc" or "desc", defaults to "asc").
//   - limit: Maximum number of results to return (for pagination).
//   - offset: Offset for pagination.
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a powerful and flexible way to search and filter visa data,
// enabling complex queries and analysis.
func SearchVisaData(c *gin.Context) {
	// Copy query params
	q := c.Request.URL.Query()
	nameFilter := strings.ToLower(q.Get("name")) // substring match
	regionFilter := strings.ToLower(q.Get("region"))
	subregionFilter := strings.ToLower(q.Get("subregion"))
	minVisaFreeStr := q.Get("minVisaFree")

	sortBy := strings.ToLower(q.Get("sortBy"))
	sortOrder := strings.ToLower(q.Get("sortOrder"))
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	// Convert minVisaFree to int
	minVisaFree := 0
	if minVisaFreeStr != "" {
		if val, err := strconv.Atoi(minVisaFreeStr); err == nil {
			minVisaFree = val
		}
	}

	// Filter
	var results []CountryVisaInfo
	for _, info := range visaData {
		if nameFilter != "" && !strings.Contains(strings.ToLower(info.Name), nameFilter) {
			continue
		}
		if regionFilter != "" && !strings.EqualFold(strings.ToLower(info.Codes.Region), regionFilter) {
			continue
		}
		if subregionFilter != "" && !strings.EqualFold(strings.ToLower(info.Codes.Subregion), subregionFilter) {
			continue
		}
		if info.PassportIndex.VisaFreeCount < minVisaFree {
			continue
		}
		results = append(results, info)
	}

	// Sort
	sort.Slice(results, func(i, j int) bool {
		switch sortBy {
		case "region":
			if sortOrder == "desc" {
				return results[i].Codes.Region > results[j].Codes.Region
			}
			return results[i].Codes.Region < results[j].Codes.Region
		case "visa_free_count":
			if sortOrder == "desc" {
				return results[i].PassportIndex.VisaFreeCount > results[j].PassportIndex.VisaFreeCount
			}
			return results[i].PassportIndex.VisaFreeCount < results[j].PassportIndex.VisaFreeCount
		default: // "name"
			if sortOrder == "desc" {
				return results[i].Name > results[j].Name
			}
			return results[i].Name < results[j].Name
		}
	})

	// Pagination
	total := len(results)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}
	results = results[offset:end]

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No visa data found for the given search criteria."})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetPassportData handles GET /v2/visas/passport/:passportCode
// @Summary Get passport visa requirements
// @Description Get visa requirements for a specific passport
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {object} PassportResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode} [get]
// GetPassportData retrieves visa requirements for a specific passport.
// This endpoint returns a simplified view of visa requirements for a given
// passport code (ISO2, ISO3, or numeric).
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides quick access to visa requirements for a specific passport,
// facilitating travel planning and compliance checks.
func GetPassportData(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}
	visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}
	passportData := PassportResponse{
		Passport: passportCodeInput,
		Visas:    visaRules,
	}
	c.JSON(http.StatusOK, passportData)
}

// GetVisaRequirementsForPassport handles GET /v2/visas/passport/:passportCode/all
// @Summary Get all visa requirements for a passport
// @Description Get all visa requirements for a specific passport (same as /visas/passport/:passportCode)
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {object} PassportResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode}/all [get]
// GetVisaRequirementsForPassport retrieves all visa requirements for a specific passport.
// This endpoint is functionally identical to `GetPassportData` and is
// included for API completeness and clarity.
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides the same functionality as `GetPassportData` with a different route
// for API consistency.
func GetVisaRequirementsForPassport(c *gin.Context) {
	// This is the *exact* same logic as the original GetPassportData,
	// just with a different route and description.
	GetPassportData(c) // Reuse the existing handler
}

// GetVisaFreeCountries handles GET /v2/visas/passport/{passportCode}/visa-free
// @Summary Get visa-free countries
// @Description Get list of countries that are visa-free for a specific passport
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {array} string
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode}/visa-free [get]
// GetVisaFreeCountries retrieves a list of countries that are visa-free for a specific passport.
// This endpoint returns a list of ISO3 country codes representing countries
// that a passport holder can visit without a visa.
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a quick way to determine visa-free travel options for a specific
// passport, aiding in travel planning and risk assessment.
func GetVisaFreeCountries(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}

	visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}

	visaFreeCountries := []string{}
	for countryCode, requirement := range visaRules {
		if isVisaFreeOrSimilar(requirement) { // Use the helper function!
			visaFreeCountries = append(visaFreeCountries, countryCode)
		}
	}

	c.JSON(http.StatusOK, visaFreeCountries)
}

// GetVisaOnArrivalCountries handles GET /v2/visas/passport/{passportCode}/visa-on-arrival
// @Summary Get visa-on-arrival countries
// @Description Get list of countries that offer visa on arrival for a specific passport
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {array} string
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode}/visa-on-arrival [get]
// GetVisaOnArrivalCountries retrieves a list of countries that offer visa on arrival for a specific passport.
// This endpoint returns a list of ISO3 country codes representing countries
// where a passport holder can obtain a visa upon arrival.
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// helps identify countries where visa on arrival is an option, facilitating
// travel planning and reducing the need for pre-arranged visas.
func GetVisaOnArrivalCountries(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}

	visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}

	visaOnArrivalCountries := []string{}
	for countryCode, requirement := range visaRules {
		if isVisaFreeOrSimilar(requirement) { // Use the helper function
			visaOnArrivalCountries = append(visaOnArrivalCountries, countryCode)
		}
	}

	c.JSON(http.StatusOK, visaOnArrivalCountries)
}

// GetEVisaCountries handles GET /v2/visas/passport/{passportCode}/e-visa
// @Summary Get e-visa countries
// @Description Get list of countries that offer e-visa for a specific passport
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {array} string
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode}/e-visa [get]
// GetEVisaCountries retrieves a list of countries that offer e-visas for a specific passport.
// This endpoint returns a list of ISO3 country codes representing countries
// where a passport holder can apply for an e-visa.
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// identifies countries where e-visa applications are possible, streamlining
// the visa application process and reducing administrative overhead.
func GetEVisaCountries(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}

	visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}

	eVisaCountries := []string{}
	for countryCode, requirement := range visaRules {
		if isVisaFreeOrSimilar(requirement) { // Use the helper function!
			eVisaCountries = append(eVisaCountries, countryCode)
		}
	}

	c.JSON(http.StatusOK, eVisaCountries)
}

// GetVisaRequiredCountries handles GET /v2/visas/passport/{passportCode}/visa-required
// @Summary Get visa-required countries
// @Description Get list of countries that require a visa for a specific passport
// @Tags Visas
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840)"
// @Success 200 {array} string
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/passport/{passportCode}/visa-required [get]
// GetVisaRequiredCountries retrieves a list of countries that require a visa for a specific passport.
// This endpoint returns a list of ISO3 country codes representing countries
// where a passport holder requires a visa for entry.
//
// Parameters:
//   - passportCode: The passport code (e.g., "USA", "US", "840").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// identifies countries where a visa is mandatory, enabling proactive visa
// application planning and ensuring compliance with entry requirements.
func GetVisaRequiredCountries(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}

	visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}

	visaRequiredCountries := []string{}
	for countryCode, requirement := range visaRules {
		if requirement == "visa required" {
			visaRequiredCountries = append(visaRequiredCountries, countryCode)
		}
	}

	c.JSON(http.StatusOK, visaRequiredCountries)
}

// GetCommonVisaFreeDestinations handles GET /v2/visas/common-visa-free
// @Summary Get common visa-free destinations
// @Description Get destinations that are visa-free for multiple passports
// @Tags Visas
// @Accept json
// @Produce json
// @Param passports query string true "Comma-separated list of passport codes"
// @Success 200 {array} string
// @Failure 400 {object} types.ErrorResponse
// @Router /visas/common-visa-free [get]
// GetCommonVisaFreeDestinations retrieves a list of countries that are visa-free for multiple specified passports.
// This endpoint returns a list of ISO3 country codes representing countries
// that are visa-free for all specified passports.
//
// Parameters:
//   - passports: A comma-separated list of passport codes (e.g., "USA,CAN,GBR").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// identifies common visa-free destinations for multiple nationalities,
// facilitating group travel planning and international collaborations.
func GetCommonVisaFreeDestinations(c *gin.Context) {
	passportCodesInput := c.Query("passports")
	if passportCodesInput == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "passports query parameter is required"})
		return
	}

	passportCodes := strings.Split(passportCodesInput, ",")
	commonVisaFree := make(map[string]int)
	passportCount := len(passportCodes)

	for _, passportCodeInput := range passportCodes {
		passportCodeInput = strings.ToUpper(strings.TrimSpace(passportCodeInput))
		passportCCA3, ok := codeToCCA3[passportCodeInput]
		if !ok {
			continue // Skip invalid passport codes
		}

		visaRules, ok := Passports[passportCCA3] // Uses Passports from handlers.go
		if !ok {
			continue // Skip if passport data is not found
		}

		for countryCode, requirement := range visaRules {
			if isVisaFreeOrSimilar(requirement) { // Use the helper function!
				commonVisaFree[countryCode]++
			}
		}
	}

	// Filter out countries that are not common to all passports
	var result []string
	for countryCode, count := range commonVisaFree {
		if count == passportCount {
			result = append(result, countryCode)
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetReciprocalVisaRequirements handles GET /v2/visas/reciprocal/{countryCode1}/{countryCode2}
// @Summary Get reciprocal visa requirements
// @Description Get mutual visa requirements between two countries
// @Tags Visas
// @Accept json
// @Produce json
// @Param countryCode1 path string true "First country code"
// @Param countryCode2 path string true "Second country code"
// @Success 200 {object} map[string]VisaRequirement
// @Failure 404 {object} types.ErrorResponse
// @Router /visas/reciprocal/{countryCode1}/{countryCode2} [get]
// GetReciprocalVisaRequirements retrieves the reciprocal visa requirements between two countries.
// This endpoint returns a map containing the visa requirements for travel
// from country1 to country2 and from country2 to country1.
//
// Parameters:
//   - countryCode1: The ISO3 country code of the first country (e.g., "USA").
//   - countryCode2: The ISO3 country code of the second country (e.g., "CAN").
//
// For enterprise, governmental, commercial, and military use, this endpoint
// provides a clear view of mutual visa requirements, facilitating diplomatic
// relations and bilateral travel agreements.
func GetReciprocalVisaRequirements(c *gin.Context) {
	countryCode1Input := strings.ToUpper(c.Param("countryCode1"))
	countryCode2Input := strings.ToUpper(c.Param("countryCode2"))

	countryCCA3_1, ok := codeToCCA3[countryCode1Input]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid country code for countryCode1"})
		return
	}

	countryCCA3_2, ok := codeToCCA3[countryCode2Input]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid country code for countryCode2"})
		return
	}

	visaRules1, ok := Passports[countryCCA3_1]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found for the first country"})
		return
	}

	visaRules2, ok := Passports[countryCCA3_2]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found for the second country"})
		return
	}

	requirement1to2, ok1 := visaRules1[countryCCA3_2]
	if !ok1 {
		requirement1to2 = "Data not available" // Or handle it as appropriate
	}

	requirement2to1, ok2 := visaRules2[countryCCA3_1]
	if !ok2 {
		requirement2to1 = "Data not available" // Or handle it as appropriate
	}

	reciprocalRequirements := map[string]VisaRequirement{
		fmt.Sprintf("%s_to_%s", countryCCA3_1, countryCCA3_2): {
			From:        countryCCA3_1,
			To:          countryCCA3_2,
			Requirement: requirement1to2,
		},
		fmt.Sprintf("%s_to_%s", countryCCA3_2, countryCCA3_1): {
			From:        countryCCA3_2,
			To:          countryCCA3_1,
			Requirement: requirement2to1,
		},
	}

	c.JSON(http.StatusOK, reciprocalRequirements)
}
