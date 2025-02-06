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
// Keyed by an upper-case ISO3 country code.
type VisaData map[string]CountryVisaInfo

// OuterVisaJSON is the JSON structure you actually have at the top level.
type OuterVisaJSON struct {
	LastUpdated    string                     `json:"last_updated"`
	TotalCountries int                        `json:"total_countries"`
	Countries      map[string]CountryVisaInfo `json:"countries"`
}

// CountryVisaInfo holds visa information for a specific country.
type CountryVisaInfo struct {
	Name          string                 `json:"name" example:"Saint Vincent and the Grenadines"`
	WikiURL       string                 `json:"wiki_url" example:"https://en.wikipedia.org/wiki/Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens"`
	Codes         CountryCodes           `json:"codes"`
	PassportIndex PassportIndex          `json:"passport_index"`
	VisaMap       VisaMap                `json:"visa_map"`
	Requirements  []VisaRequirementEntry `json:"requirements"`
}

// CountryCodes represents various country codes.
type CountryCodes struct {
	ISO2      string `json:"iso2" example:"VC"`
	ISO3      string `json:"iso3" example:"VCT"`
	Region    string `json:"region" example:"Americas"`
	Subregion string `json:"subregion" example:"Caribbean"`
}

// PassportIndex represents passport ranking information.
type PassportIndex struct {
	VisaFreeCount int    `json:"visa_free_count" example:"157"`
	Ranking       *int   `json:"ranking" example:"25"` // Use pointer for nullable int
	RankingSource string `json:"ranking_source" example:"Henley Passport Index"`
}

// VisaMap represents the visa map information.
type VisaMap struct {
	MapURL string            `json:"map_url" example:"https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens.png/800px-Visa_requirements_for_Saint_Vincent_and_the_Grenadines_citizens.png"`
	Legend map[string]string `json:"legend"`
}

// VisaRequirementEntry represents a single visa requirement entry.
type VisaRequirementEntry struct {
	Country         string `json:"country" example:"Afghanistan"`
	VisaRequirement string `json:"visa_requirement" example:"Visa required"`
	AllowedStay     string `json:"allowed_stay" example:""`
	Notes           string `json:"notes" example:""`
	ISO2            string `json:"iso2" example:"AF"`
	ISO3            string `json:"iso3" example:"AFG"`
	Region          string `json:"region" example:"Asia"`
	Subregion       string `json:"subregion" example:"Southern Asia"`
}

// visaData holds the loaded visa data.
var visaData VisaData

// ----------------------------------------------------------------------------
// LOADING / INITIAL SETUP
// ----------------------------------------------------------------------------

// LoadVisaData loads and parses the visa data from the JSON file
func LoadVisaData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read visa data file: %w", err)
	}
	var outer OuterVisaJSON
	if err := json.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("failed to parse visa data: %w", err)
	}
	if outer.Countries == nil { // Add this check
		return fmt.Errorf("visa data file is missing 'countries' field")
	}
	visaData = outer.Countries // Use ISO3 keys directly

	// Add ISO2 and ISO3 to the codeToCCA3 map in handlers.go.  This is crucial.
	for _, info := range visaData {
		AddCodesToCCA3Map(info.Codes.ISO2, info.Codes.ISO3)
	}
	return nil
}

// RegisterVisaRoutes sets up all the visa-related routes under /v2/visas/*.
func RegisterVisaRoutes(r *gin.RouterGroup) {
	visas := r.Group("/visas")
	{
		// Base endpoints
		visas.GET("", GetAllVisaData)
		visas.GET("/search", SearchVisaData)

		// Country-specific endpoints
		visas.GET("/:countryCode", GetVisaRequirementsByCountry)
		visas.GET("/:countryCode/filtered", GetFilteredVisaRequirements)

		// Destination endpoints
		dest := visas.Group("/destination")
		{
			dest.GET("/:destinationCode", GetVisaRequirementsForDestination)
			dest.GET("/:destinationCode/sorted", GetSortedVisaRequirementsForDestination)
		}

		// Comparison and analysis endpoints
		visas.GET("/compare", CompareVisaRequirementsCountries)

		// --- Routes moved from /v2/passports ---
		visas.GET("/passport/:passportCode/visa-free", GetVisaFreeCountries)
		visas.GET("/passport/:passportCode/visa-on-arrival", GetVisaOnArrivalCountries)
		visas.GET("/passport/:passportCode/e-visa", GetEVisaCountries)
		visas.GET("/passport/:passportCode/visa-required", GetVisaRequiredCountries)
		visas.GET("/passport/:passportCode", GetPassportData) // Renamed route
		visas.GET("/requirements", GetVisaRequirements)       // Renamed and moved
		visas.GET("/ranking", GetPassportRanking)
		visas.GET("/common-visa-free", GetCommonVisaFreeDestinations)
		visas.GET("/reciprocal/:countryCode1/:countryCode2", GetReciprocalVisaRequirements)
		visas.GET("/passport/:passportCode/all", GetVisaRequirementsForPassport) // Renamed
	}
}

// ----------------------------------------------------------------------------
// HELPER FUNCTIONS
// ----------------------------------------------------------------------------

// getCountryVisaInfo is a helper function to retrieve visa info by country code.
func getCountryVisaInfo(countryCode string) (*CountryVisaInfo, bool) {
	countryCode = strings.ToUpper(countryCode)
	info, ok := visaData[countryCode] // Direct lookup!
	return &info, ok
}

// parseInt is a simple helper to parse a query param as an integer.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// isVisaFreeOrOnArrival is moved to handlers.go

// ----------------------------------------------------------------------------
// ROUTE HANDLERS
// ----------------------------------------------------------------------------

// GetAllVisaData handles GET /v2/visas
func GetAllVisaData(c *gin.Context) {
	if len(visaData) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No visa data found or not loaded."})
		return
	}
	c.JSON(http.StatusOK, visaData)
}

// GetVisaRequirementsByCountry handles GET /v2/visas/{countryCode}
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

// CompareVisaRequirementsCountries handles GET /v2/visas/compare
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
type VisaComparisonResult struct {
	Country1              string               `json:"country1" example:"USA"`
	Country2              string               `json:"country2" example:"CAN"`
	Requirements          map[string]string    `json:"requirements"` // e.g. "USA_to_CAN" : "Visa required"
	CommonAccess          []CommonAccessResult `json:"common_access"`
	Country1PassportIndex PassportIndex        `json:"country1_passport_index"`
	Country2PassportIndex PassportIndex        `json:"country2_passport_index"`
}

// CommonAccessResult represents a country accessible by both compared countries.
type CommonAccessResult struct {
	CountryCode  string `json:"country_code" example:"MEX"`
	CountryName  string `json:"country_name" example:"Mexico"`
	Requirement1 string `json:"requirement_1" example:"Visa not required"`
	Requirement2 string `json:"requirement_2" example:"Visa not required"`
}

// GetVisaRequirementsForDestination handles GET /v2/visas/destination/{destinationCode}
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
type VisaDestinationInfo struct {
	DestinationCountry string                       `json:"destination_country" example:"France"`
	Requirements       []DestinationVisaRequirement `json:"requirements"`
}

// DestinationVisaRequirement represents the visa requirement for a specific passport to visit that destination.
type DestinationVisaRequirement struct {
	PassportCountry string `json:"passport_country" example:"United States"`
	VisaRequirement string `json:"visa_requirement" example:"Visa not required"`
	AllowedStay     string `json:"allowed_stay" example:"90 days"`
	Notes           string `json:"notes" example:""`
	ISO2            string `json:"iso2" example:"US"`
	ISO3            string `json:"iso3" example:"USA"`
}

// GetSortedVisaRequirementsForDestination handles GET /v2/visas/destination/{destinationCode}/sorted
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
type SortedVisaDestinationInfo struct {
	DestinationCountry string                       `json:"destination_country" example:"France"`
	Requirements       []DestinationVisaRequirement `json:"requirements"`
	SortedBy           string                       `json:"sorted_by" example:"visa_requirement"`
}

// ----------------------------------------------------------------------------
// ADVANCED "SEARCH" ENDPOINT
// ----------------------------------------------------------------------------

// SearchVisaData handles GET /v2/visas/search
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
func GetVisaRequirementsForPassport(c *gin.Context) {
	// This is the *exact* same logic as the original GetPassportData,
	// just with a different route and description.
	GetPassportData(c) // Reuse the existing handler
}

// GetVisaRequirements handles GET /v2/visas/requirements
func GetVisaRequirements(c *gin.Context) {
	fromCountryInput := strings.ToUpper(c.Query("fromCountry"))
	toCountryInput := strings.ToUpper(c.Query("toCountry"))

	if fromCountryInput == "" || toCountryInput == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "fromCountry and toCountry query parameters are required"})
		return
	}

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
	c.JSON(http.StatusOK, VisaRequirement{
		From:        fromCountryInput,
		To:          toCountryInput,
		Requirement: requirement,
	})
}

// GetVisaFreeCountries handles GET /v2/visas/passport/{passportCode}/visa-free
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

// GetPassportRanking handles GET /v2/visas/ranking
func GetPassportRanking(c *gin.Context) {
	type PassportRank struct {
		PassportCode  string `json:"passportCode"`
		Rank          int    `json:"rank"`
		VisaFreeCount int    `json:"visaFreeCount"`
	}

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

// GetCommonVisaFreeDestinations handles GET /v2/visas/common-visa-free
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
