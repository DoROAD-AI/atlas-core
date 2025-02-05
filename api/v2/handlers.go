// handlers.go - handlers for the v2 API.
package v2

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/DoROAD-AI/atlas/types"
	"github.com/gin-gonic/gin"

	v1 "github.com/DoROAD-AI/atlas/api/v1" // Import v1 to access Countries data
)

// PassportData defines the structure for passports data
type PassportData map[string]map[string]string

// Passports holds the passports data
var Passports PassportData

// codeToCCA3 maps various country codes to their CCA3 code
var codeToCCA3 map[string]string

// AirportData holds the airport data keyed by alpha-2 country code
// (e.g., "VC" -> { ... }).
var AirportData map[string]CountryAirports

// VisaRequirement represents the visa requirement between two countries.
// @Description VisaRequirement represents the visa requirement between two countries.
type VisaRequirement struct {
	From        string `json:"from" example:"USA"`
	To          string `json:"to" example:"DEU"`
	Requirement string `json:"requirement" example:"90"`
}

// AirportFrequency represents the frequency data for an airport.
// @Description AirportFrequency represents the frequency data for an airport.
type AirportFrequency struct {
	ID           string `json:"id" example:"322388"`
	AirportRef   string `json:"airport_ref" example:"322383"`
	AirportIdent string `json:"airport_ident" example:"TVSA"`
	Type         string `json:"type" example:"APP"`
	Description  string `json:"description" example:"Argyle Approach"`
	FrequencyMHz string `json:"frequency_mhz" example:"120.8"`
}

// AirportRunway represents the runway data for an airport.
// @Description AirportRunway represents the runway data for an airport.
type AirportRunway struct {
	ID                     string `json:"id" example:"322384"`
	AirportRef             string `json:"airport_ref" example:"322383"`
	AirportIdent           string `json:"airport_ident" example:"TVSA"`
	LengthFt               string `json:"length_ft" example:"9000"`
	WidthFt                string `json:"width_ft" example:"148"`
	Surface                string `json:"surface" example:"ASP"`
	Lighted                string `json:"lighted" example:"1"`
	Closed                 string `json:"closed" example:"0"`
	LEIdent                string `json:"le_ident" example:"04"`
	LELatitudeDeg          string `json:"le_latitude_deg" example:""`
	LELongitudeDeg         string `json:"le_longitude_deg" example:""`
	LEElevationFt          string `json:"le_elevation_ft" example:"136"`
	LEHeadingDegT          string `json:"le_heading_degT" example:""`
	LEDisplacedThresholdFt string `json:"le_displaced_threshold_ft" example:""`
	HEIdent                string `json:"he_ident" example:"22"`
	HELatitudeDeg          string `json:"he_latitude_deg" example:""`
	HELongitudeDeg         string `json:"he_longitude_deg" example:""`
	HEElevationFt          string `json:"he_elevation_ft" example:"52"`
	HEHeadingDegT          string `json:"he_heading_degT" example:""`
	HEDisplacedThresholdFt string `json:"he_displaced_threshold_ft" example:"985"`
}

// Airport represents the airport data.
// @Description Airport represents the airport data.
type Airport struct {
	ID               string             `json:"id" example:"322383"`
	Ident            string             `json:"ident" example:"TVSA"`
	Type             string             `json:"type" example:"medium_airport"`
	Name             string             `json:"name" example:"Argyle International Airport"`
	LatitudeDeg      string             `json:"latitude_deg" example:"13.156695"`
	LongitudeDeg     string             `json:"longitude_deg" example:"-61.149945"`
	ElevationFt      string             `json:"elevation_ft" example:"136"`
	Continent        string             `json:"continent" example:"NA"`
	ISOCountry       string             `json:"iso_country" example:"VC"`
	ISORegion        string             `json:"iso_region" example:"VC-04"`
	Municipality     string             `json:"municipality" example:"Kingstown"`
	ScheduledService string             `json:"scheduled_service" example:"yes"`
	GPSCode          string             `json:"gps_code" example:"TVSA"`
	IATACode         string             `json:"iata_code" example:"SVD"`
	LocalCode        string             `json:"local_code" example:""`
	HomeLink         string             `json:"home_link" example:"http://www.svgiadc.com"`
	WikipediaLink    string             `json:"wikipedia_link" example:"https://en.m.wikipedia.org/wiki/Argyle_International_Airport"`
	Keywords         string             `json:"keywords" example:""`
	Comments         []string           `json:"comments" example:""`
	Frequencies      []AirportFrequency `json:"frequencies"`
	Runways          []AirportRunway    `json:"runways"`
}

// CountryAirports represents the airport data for a country.
// @Description CountryAirports represents the airport data for a country.
type CountryAirports struct {
	ID            string    `json:"id" example:"302756"`
	Code          string    `json:"code" example:"VC"`
	Name          string    `json:"name" example:"Saint Vincent and the Grenadines"`
	Continent     string    `json:"continent" example:"NA"`
	WikipediaLink string    `json:"wikipedia_link" example:"https://en.wikipedia.org/wiki/Saint_Vincent_and_the_Grenadines"`
	Keywords      string    `json:"keywords" example:"Airports in Saint Vincent and the Grenadines"`
	Airports      []Airport `json:"airports"`
}

// PassportResponse represents the passport data response.
// @Description PassportResponse represents the passport data response.
type PassportResponse struct {
	Passport string            `json:"passport" example:"USA"`
	Visas    map[string]string `json:"visas"`
}

// init initializes the global codeToCCA3 map (used for passports and for
// mapping any recognized country code to its CCA3 form).
func init() {
	codeToCCA3 = make(map[string]string)
}

// LoadPassportData reads local JSON data into the global Passports variable.
func LoadPassportData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read passports file: %w", err)
	}
	if err := json.Unmarshal(data, &Passports); err != nil {
		return fmt.Errorf("failed to parse passports data: %w", err)
	}
	// Initialize code mapping after loading passports
	initCodeMapping()
	return nil
}

// LoadAirportsData loads airport data from a JSON file into AirportData.
func LoadAirportsData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read airports file: %w", err)
	}
	if err := json.Unmarshal(data, &AirportData); err != nil {
		return fmt.Errorf("failed to parse airports data: %w", err)
	}
	return nil
}

// initCodeMapping builds a mapping from various country codes to CCA3 codes.
// This mapping is used both for passport data and to route "country codes"
// to a single standard (CCA3).
func initCodeMapping() {
	for _, country := range v1.Countries {
		cca3 := country.CCA3
		codes := []string{
			country.CCA2,
			country.CCA3,
			country.CCN3,
			country.CIOC,
			country.FIFA,
		}
		// Include alternative spellings
		for _, alt := range country.AltSpellings {
			codes = append(codes, strings.ToUpper(alt))
		}
		for _, code := range codes {
			if code != "" {
				codeToCCA3[strings.ToUpper(code)] = cca3
			}
		}
	}
}

// toAlpha2 attempts to convert an arbitrary country code (CCA2, CCA3, CCN3, etc.)
// to its ISO alpha-2 equivalent. It returns the alpha-2 code if found.
func toAlpha2(code string) (string, bool) {
	upper := strings.ToUpper(code)
	cca3, ok := codeToCCA3[upper]
	if !ok {
		return "", false
	}
	// Find the country in the v1.Countries slice by matching alpha-3
	for _, c := range v1.Countries {
		if c.CCA3 == cca3 {
			return c.CCA2, true
		}
	}
	return "", false
}

// ----------------------------------------------------------------------------
// Passport Handlers
// ----------------------------------------------------------------------------

// GetPassportData handles GET /passports/:passportCode
// @Summary     Get passport data
// @Description Get visa requirement data for a specific passport.
// @Tags        Passports
// @Accept      json
// @Produce     json
// @Param       passportCode path string true "Passport code (e.g., USA, US, 840, etc.)"
// @Success     200 {object} PassportResponse
// @Failure     404 {object} ErrorResponse
// @Router      /passports/{passportCode} [get]
func GetBasicPassportData(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}
	visaRules, ok := Passports[passportCCA3]
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

// GetVisaDetails handles GET /v2/passports/{passportCode}/visa-details/{destinationCode}
// @Summary Get detailed visa requirements for a passport and destination
// @Description Provides specific visa requirement details (duration, type, etc.) for a given passport and destination.
// @Tags Passports
// @Accept json
// @Produce json
// @Param passportCode path string true "Passport code (e.g., USA, US, 840, etc.)"
// @Param destinationCode path string true "Destination country code (e.g., DEU, DE, 276, etc.)"
// @Success 200 {object} VisaRequirement
// @Failure 404 {object} ErrorResponse
// @Router /passports/{passportCode}/visa-details/{destinationCode} [get]
func GetVisaDetails(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	destinationCodeInput := strings.ToUpper(c.Param("destinationCode"))

	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid passport country code"})
		return
	}

	destinationCCA3, ok := codeToCCA3[destinationCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid destination country code"})
		return
	}

	visaRules, ok := Passports[passportCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Passport data not found"})
		return
	}

	requirement, ok := visaRules[destinationCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Visa requirement data not found for this destination"})
		return
	}

	visaDetails := VisaRequirement{
		From:        passportCCA3,
		To:          destinationCCA3,
		Requirement: requirement,
	}

	c.JSON(http.StatusOK, visaDetails)
}

// CompareVisaRequirements handles GET /v2/passports/compare
// @Summary Compare visa requirements for multiple passports to a single destination
// @Description Compares visa requirements for a list of passports to a single destination.
// @Tags Passports
// @Accept json
// @Produce json
// @Param passports query []string true "Comma-separated list of passport codes (e.g., USA,DEU,JPN)"
// @Param destination query string true "Destination country code (e.g., FRA)"
// @Success 200 {object} map[string]VisaRequirement
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /passports/compare [get]
func CompareVisaRequirements(c *gin.Context) {
	passportCodesInput := c.Query("passports")
	destinationCodeInput := strings.ToUpper(c.Query("destination"))

	if passportCodesInput == "" || destinationCodeInput == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "passports and destination query parameters are required"})
		return
	}

	passportCodes := strings.Split(passportCodesInput, ",")
	destinationCCA3, ok := codeToCCA3[destinationCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid destination country code"})
		return
	}

	comparisonResults := make(map[string]VisaRequirement)
	for _, passportCodeInput := range passportCodes {
		passportCodeInput = strings.ToUpper(strings.TrimSpace(passportCodeInput))
		passportCCA3, ok := codeToCCA3[passportCodeInput]
		if !ok {
			comparisonResults[passportCodeInput] = VisaRequirement{
				From:        passportCodeInput,
				To:          destinationCodeInput,
				Requirement: "Invalid passport code",
			}
			continue
		}

		visaRules, ok := Passports[passportCCA3]
		if !ok {
			comparisonResults[passportCodeInput] = VisaRequirement{
				From:        passportCCA3,
				To:          destinationCodeInput,
				Requirement: "Passport data not found",
			}
			continue
		}

		requirement, ok := visaRules[destinationCCA3]
		if !ok {
			comparisonResults[passportCodeInput] = VisaRequirement{
				From:        passportCCA3,
				To:          destinationCodeInput,
				Requirement: "Data not available",
			}
		} else {
			comparisonResults[passportCodeInput] = VisaRequirement{
				From:        passportCCA3,
				To:          destinationCodeInput,
				Requirement: requirement,
			}
		}
	}

	c.JSON(http.StatusOK, comparisonResults)
}

// ----------------------------------------------------------------------------
// Airports Handlers
// ----------------------------------------------------------------------------

// GetAllAirports handles GET /airports
// @Summary     Get all airports
// @Description Retrieves a list of all airports for all countries (keyed by each country's alpha-2 code).
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Success     200 {object} map[string]CountryAirports
// @Failure     500 {object} ErrorResponse
// @Router      /airports [get]
func GetAllAirports(c *gin.Context) {
	c.JSON(http.StatusOK, AirportData)
}

// GetAirportsByCountry handles GET /airports/:countryCode
// @Summary     Get airports by country
// @Description Retrieves all airports in a specific country. The country code can be in any recognized format (CCA2, CCA3, CCN3, CIOC, FIFA, or alt spelling).
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Param       countryCode path string true "Country code (e.g., VC, VCT, 670, etc.)"
// @Success     200 {object} CountryAirports
// @Failure     404 {object} ErrorResponse
// @Router      /airports/{countryCode} [get]
func GetAirportsByCountry(c *gin.Context) {
	countryParam := c.Param("countryCode")

	// Convert any recognized code to alpha-2
	alpha2, ok := toAlpha2(countryParam)
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid or unrecognized country code"})
		return
	}

	// Retrieve airport data by alpha-2 code
	countryAirports, found := AirportData[strings.ToUpper(alpha2)]
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airport data found for this country"})
		return
	}

	c.JSON(http.StatusOK, countryAirports)
}

// GetAirportByIdent handles GET /airports/:countryCode/:airportIdent
// @Summary     Get a single airport by identifier
// @Description Retrieves a specific airport within a country by matching the airport's ICAO or IATA code. The country code can be in any recognized format (CCA2, CCA3, CCN3, CIOC, etc.).
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Param       countryCode   path string true "Country code (e.g., VC, VCT, 670, etc.)"
// @Param       airportIdent  path string true "Airport Ident (ICAO) or IATA code"
// @Success     200 {object} Airport
// @Failure     404 {object} ErrorResponse
// @Router      /airports/{countryCode}/{airportIdent} [get]
func GetAirportByIdent(c *gin.Context) {
	countryParam := c.Param("countryCode")
	airportIdent := strings.ToUpper(c.Param("airportIdent"))

	// Convert any recognized code to alpha-2
	alpha2, ok := toAlpha2(countryParam)
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid or unrecognized country code"})
		return
	}

	countryAirports, found := AirportData[strings.ToUpper(alpha2)]
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airport data found for this country"})
		return
	}

	// Search airports array by matching ident or IATA code
	for _, airport := range countryAirports.Airports {
		if strings.EqualFold(airport.Ident, airportIdent) || strings.EqualFold(airport.IATACode, airportIdent) {
			c.JSON(http.StatusOK, airport)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airport not found in this country"})
}

// ----------------------------------------------------------------------------
// New Airport Handlers (Enhanced and Enterprise)
// ----------------------------------------------------------------------------

// AirportDistance represents the distance between two airports.
// @Description AirportDistance represents the distance between two airports.
type AirportDistance struct {
	Airport1      string  `json:"airport1" example:"TVSA"`
	Airport2      string  `json:"airport2" example:"TVSB"`
	DistanceKM    float64 `json:"distance_km" example:"1234.5"`
	DistanceMiles float64 `json:"distance_miles" example:"767.1"`
}

// GetAirportByCode handles GET /v2/airports/by-code/{airportCode}
// @Summary Get airport by ICAO or IATA code
// @Description Retrieves a specific airport by its ICAO or IATA code.
// @Tags Airports
// @Accept json
// @Produce json
// @Param airportCode path string true "Airport ICAO or IATA code"
// @Success 200 {object} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/by-code/{airportCode} [get]
func GetAirportByCode(c *gin.Context) {
	airportCode := strings.ToUpper(c.Param("airportCode"))

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.Ident == airportCode || airport.IATACode == airportCode {
				c.JSON(http.StatusOK, airport)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airport not found"})
}

// GetAirportsByRegion handles GET /v2/airports/region/{isoRegion}
// @Summary Get airports by ISO region
// @Description Retrieves all airports within a specific ISO region.
// @Tags Airports
// @Accept json
// @Produce json
// @Param isoRegion path string true "ISO region code (e.g., VC-04)"
// @Success 200 {array} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/region/{isoRegion} [get]
func GetAirportsByRegion(c *gin.Context) {
	isoRegion := strings.ToUpper(c.Param("isoRegion"))
	var airportsInRegion []Airport

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.ISORegion == isoRegion {
				airportsInRegion = append(airportsInRegion, airport)
			}
		}
	}

	if len(airportsInRegion) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found for this ISO region"})
		return
	}

	c.JSON(http.StatusOK, airportsInRegion)
}

// GetAirportsByMunicipality handles GET /v2/airports/municipality/{municipalityName}
// @Summary Get airports by municipality
// @Description Retrieves all airports within a specific municipality.
// @Tags Airports
// @Accept json
// @Produce json
// @Param municipalityName path string true "Municipality name"
// @Success 200 {array} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/municipality/{municipalityName} [get]
func GetAirportsByMunicipality(c *gin.Context) {
	municipalityName := strings.ToUpper(c.Param("municipalityName"))
	var airportsInMunicipality []Airport

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if strings.EqualFold(airport.Municipality, municipalityName) {
				airportsInMunicipality = append(airportsInMunicipality, airport)
			}
		}
	}

	if len(airportsInMunicipality) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found in this municipality"})
		return
	}

	c.JSON(http.StatusOK, airportsInMunicipality)
}

// GetAirportsByType handles GET /v2/airports/type/{airportType}
// @Summary Get airports by type
// @Description Retrieves all airports of a specific type.
// @Tags Airports
// @Accept json
// @Produce json
// @Param airportType path string true "Airport type (e.g., medium_airport, closed)"
// @Success 200 {array} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/type/{airportType} [get]
func GetAirportsByType(c *gin.Context) {
	airportType := strings.ToLower(c.Param("airportType"))
	var matchingAirports []Airport

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.Type == airportType {
				matchingAirports = append(matchingAirports, airport)
			}
		}
	}

	if len(matchingAirports) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found for this type"})
		return
	}

	c.JSON(http.StatusOK, matchingAirports)
}

// GetAirportsWithScheduledService handles GET /v2/airports/scheduled
// @Summary Get airports with scheduled service
// @Description Retrieves all airports that have scheduled airline service.
// @Tags Airports
// @Accept json
// @Produce json
// @Success 200 {array} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/scheduled [get]
func GetAirportsWithScheduledService(c *gin.Context) {
	var scheduledAirports []Airport

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.ScheduledService == "yes" {
				scheduledAirports = append(scheduledAirports, airport)
			}
		}
	}

	if len(scheduledAirports) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports with scheduled service found"})
		return
	}

	c.JSON(http.StatusOK, scheduledAirports)
}

// GetAirportRunways handles GET /v2/airports/{countryCode}/{airportIdent}/runways
// @Summary Get airport runways
// @Description Retrieves detailed runway information for a specific airport.
// @Tags Airports
// @Accept json
// @Produce json
// @Param countryCode path string true "Country code (e.g., VC, VCT, 670, etc.)"
// @Param airportIdent path string true "Airport Ident (ICAO) or IATA code"
// @Success 200 {array} AirportRunway
// @Failure 404 {object} ErrorResponse
// @Router /airports/{countryCode}/{airportIdent}/runways [get]
func GetAirportRunways(c *gin.Context) {
	countryParam := c.Param("countryCode")
	airportIdent := strings.ToUpper(c.Param("airportIdent"))

	alpha2, ok := toAlpha2(countryParam)
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid or unrecognized country code"})
		return
	}

	countryAirports, found := AirportData[strings.ToUpper(alpha2)]
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airport data found for this country"})
		return
	}

	for _, airport := range countryAirports.Airports {
		if airport.Ident == airportIdent || airport.IATACode == airportIdent {
			c.JSON(http.StatusOK, airport.Runways)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airport not found in this country"})
}

// GetAirportFrequencies handles GET /v2/airports/{countryCode}/{airportIdent}/frequencies
// @Summary Get airport frequencies
// @Description Retrieves communication frequencies used at a specific airport.
// @Tags Airports
// @Accept json
// @Produce json
// @Param countryCode path string true "Country code (e.g., VC, VCT, 670, etc.)"
// @Param airportIdent path string true "Airport Ident (ICAO) or IATA code"
// @Success 200 {array} AirportFrequency
// @Failure 404 {object} ErrorResponse
// @Router /airports/{countryCode}/{airportIdent}/frequencies [get]
func GetAirportFrequencies(c *gin.Context) {
	countryParam := c.Param("countryCode")
	airportIdent := strings.ToUpper(c.Param("airportIdent"))

	alpha2, ok := toAlpha2(countryParam)
	if !ok {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Invalid or unrecognized country code"})
		return
	}

	countryAirports, found := AirportData[strings.ToUpper(alpha2)]
	if !found {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airport data found for this country"})
		return
	}

	for _, airport := range countryAirports.Airports {
		if airport.Ident == airportIdent || airport.IATACode == airportIdent {
			c.JSON(http.StatusOK, airport.Frequencies)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airport not found in this country"})
}

// SearchAirports handles GET /v2/airports/search?query={searchString}
// @Summary Search airports
// @Description Performs a flexible search for airports based on a query string.
// @Tags Airports
// @Accept json
// @Produce json
// @Param query query string true "Search string (can match airport name, city, ICAO/IATA code, etc.)"
// @Success 200 {array} Airport
// @Failure 400 {object} ErrorResponse
// @Router /airports/search [get]
func SearchAirports(c *gin.Context) {
	searchString := strings.ToUpper(c.Query("query"))
	if searchString == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Query parameter 'query' is required"})
		return
	}

	var matchingAirports []Airport
	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if strings.Contains(strings.ToUpper(airport.Name), searchString) ||
				strings.Contains(strings.ToUpper(airport.Municipality), searchString) ||
				strings.Contains(strings.ToUpper(airport.Ident), searchString) ||
				strings.Contains(strings.ToUpper(airport.IATACode), searchString) ||
				strings.Contains(strings.ToUpper(airport.GPSCode), searchString) {
				matchingAirports = append(matchingAirports, airport)
			}
		}
	}

	if len(matchingAirports) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found matching the search criteria"})
		return
	}

	c.JSON(http.StatusOK, matchingAirports)
}

// GetAirportsWithinRadius handles GET /v2/airports/radius?latitude={latitude}&longitude={longitude}&radius={radiusInKm}
// @Summary Get airports within a radius
// @Description Retrieves all airports within a specified radius of a given latitude/longitude coordinate.
// @Tags Airports
// @Accept json
// @Produce json
// @Param latitude query number true "Latitude of the center point"
// @Param longitude query number true "Longitude of the center point"
// @Param radius query number true "Radius in kilometers"
// @Success 200 {array} Airport
// @Failure 400 {object} ErrorResponse
// @Router /airports/radius [get]
func GetAirportsWithinRadius(c *gin.Context) {
	latitude, err := parseFloatQueryParam(c, "latitude")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid latitude"})
		return
	}
	longitude, err := parseFloatQueryParam(c, "longitude")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid longitude"})
		return
	}
	radius, err := parseFloatQueryParam(c, "radius")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid radius"})
		return
	}

	var airportsWithinRadius []Airport
	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			airportLat, _ := parseFloat(airport.LatitudeDeg)
			airportLon, _ := parseFloat(airport.LongitudeDeg)
			distance := calculateHaversineDistance(latitude, longitude, airportLat, airportLon)
			if distance <= radius {
				airportsWithinRadius = append(airportsWithinRadius, airport)
			}
		}
	}

	if len(airportsWithinRadius) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found within the specified radius"})
		return
	}

	c.JSON(http.StatusOK, airportsWithinRadius)
}

// parseFloatQueryParam is a helper function to parse a float64 query parameter.
func parseFloatQueryParam(c *gin.Context, paramName string) (float64, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return 0, fmt.Errorf("parameter %s is required", paramName)
	}
	return parseFloat(valueStr)
}

// parseFloat is a helper function to parse a float64 string.
func parseFloat(valueStr string) (float64, error) {
	return strconv.ParseFloat(valueStr, 64)
}

// calculateHaversineDistance calculates the distance between two points on the Earth.
func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadiusKm * c

	return distance
}

// CalculateDistanceBetweenAirports handles GET /v2/airports/distance?airport1={airportCode1}&airport2={airportCode2}
// @Summary Calculate distance between two airports
// @Description Calculates the distance (in kilometers and miles) between two airports.
// @Tags Airports
// @Accept json
// @Produce json
// @Param airport1 query string true "ICAO or IATA code of the first airport"
// @Param airport2 query string true "ICAO or IATA code of the second airport"
// @Success 200 {object} AirportDistance
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /airports/distance [get]
func CalculateDistanceBetweenAirports(c *gin.Context) {
	airportCode1 := strings.ToUpper(c.Query("airport1"))
	airportCode2 := strings.ToUpper(c.Query("airport2"))

	if airportCode1 == "" || airportCode2 == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Both airport1 and airport2 query parameters are required"})
		return
	}

	airport1, found1 := findAirportByCode(airportCode1)
	airport2, found2 := findAirportByCode(airportCode2)

	if !found1 || !found2 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "One or both airports not found"})
		return
	}

	lat1, _ := strconv.ParseFloat(airport1.LatitudeDeg, 64)
	lon1, _ := strconv.ParseFloat(airport1.LongitudeDeg, 64)
	lat2, _ := strconv.ParseFloat(airport2.LatitudeDeg, 64)
	lon2, _ := strconv.ParseFloat(airport2.LongitudeDeg, 64)

	distanceKm := calculateHaversineDistance(lat1, lon1, lat2, lon2)
	distanceMiles := distanceKm * 0.621371 // Convert kilometers to miles

	c.JSON(http.StatusOK, AirportDistance{
		Airport1:      airportCode1,
		Airport2:      airportCode2,
		DistanceKM:    distanceKm,
		DistanceMiles: distanceMiles,
	})
}

// findAirportByCode is a helper function to find an airport by its ICAO or IATA code.
func findAirportByCode(airportCode string) (*Airport, bool) {
	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.Ident == airportCode || airport.IATACode == airportCode {
				return &airport, true
			}
		}
	}
	return nil, false
}

// GetAirportsByKeyword handles GET /v2/airports/keyword/{keyword}
// @Summary Get airports by keyword
// @Description Retrieves all airports associated with a specific keyword.
// @Tags Airports
// @Accept json
// @Produce json
// @Param keyword path string true "Keyword to search for"
// @Success 200 {array} Airport
// @Failure 404 {object} ErrorResponse
// @Router /airports/keyword/{keyword} [get]
func GetAirportsByKeyword(c *gin.Context) {
	keyword := strings.ToLower(c.Param("keyword"))
	var matchingAirports []Airport

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if strings.Contains(strings.ToLower(airport.Keywords), keyword) {
				matchingAirports = append(matchingAirports, airport)
			}
		}
	}

	if len(matchingAirports) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found matching the keyword"})
		return
	}

	c.JSON(http.StatusOK, matchingAirports)
}

// SuperTypeQuery handles GET /v2/search
// @Summary Super Type Query
// @Description Performs a comprehensive search across all data types (countries, airports) based on query parameters.
// @Tags Search
// @Accept json
// @Produce json
// @Param type query string false "Type of data to search for (country, airport). If omitted or set to 'all', searches across all data types."
// @Param name query string false "Name of the country or airport"
// @Param region query string false "Region of the country"
// @Param subregion query string false "Subregion of the country"
// @Param cca2 query string false "Country code Alpha-2"
// @Param cca3 query string false "Country code Alpha-3"
// @Param ccn3 query string false "Country code Numeric"
// @Param capital query string false "Capital city of the country"
// @Param ident query string false "Airport Ident code (e.g., ICAO code)"
// @Param iata_code query string false "Airport IATA code"
// @Param iso_country query string false "ISO country code for airports"
// @Param iso_region query string false "ISO region code for airports"
// @Param municipality query string false "Municipality of the airport"
// @Param airport_type query string false "Type of the airport"
// @Success 200 {object} interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /search [get]
func SuperTypeQuery(c *gin.Context) {
	// Get the 'type' query parameter
	dataType := c.Query("type")

	// Copy all query parameters
	queryParams := c.Request.URL.Query()

	// Remove 'type' from queryParams
	delete(queryParams, "type")

	// If queryParams is empty, return error
	if len(queryParams) == 0 {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "At least one query parameter is required"})
		return
	}

	switch strings.ToLower(dataType) {
	case "country":
		results := searchCountries(queryParams)
		if len(results) == 0 {
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No countries found matching the criteria"})
			return
		}
		c.JSON(http.StatusOK, results)

	case "airport":
		results := searchAirports(queryParams)
		if len(results) == 0 {
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airports found matching the criteria"})
			return
		}
		c.JSON(http.StatusOK, results)

	case "", "all":
		// Search both
		var combinedResults struct {
			Countries []v1.Country `json:"countries"`
			Airports  []Airport    `json:"airports"`
		}

		countries := searchCountries(queryParams)
		airports := searchAirports(queryParams)
		combinedResults.Countries = countries
		combinedResults.Airports = airports

		if len(countries) == 0 && len(airports) == 0 {
			c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No results found matching the criteria"})
			return
		}

		c.JSON(http.StatusOK, combinedResults)

	default:
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid type parameter. Allowed values are 'country', 'airport', or 'all'"})
		return
	}
}

// searchCountries searches countries based on query parameters (partial match, case-insensitive, OR match for multi-values).
func searchCountries(queryParams url.Values) []v1.Country {
	var results []v1.Country

OuterLoop:
	for _, country := range v1.Countries {
		// For each country, check all query parameters
		for key, values := range queryParams {
			// If multiple values for the same key: treat them as OR
			// If at least one value matches, the parameter is satisfied
			matchedParam := false

			for _, val := range values {
				val = strings.TrimSpace(strings.ToLower(val))

				switch strings.ToLower(key) {
				case "name":
					// Check partial in Common or Official
					if strings.Contains(strings.ToLower(country.Name.Common), val) ||
						strings.Contains(strings.ToLower(country.Name.Official), val) {
						matchedParam = true
						break
					}

				case "region":
					if strings.Contains(strings.ToLower(country.Region), val) {
						matchedParam = true
						break
					}

				case "subregion":
					if strings.Contains(strings.ToLower(country.Subregion), val) {
						matchedParam = true
						break
					}

				case "cca2":
					if strings.Contains(strings.ToLower(country.CCA2), val) {
						matchedParam = true
						break
					}

				case "cca3":
					if strings.Contains(strings.ToLower(country.CCA3), val) {
						matchedParam = true
						break
					}

				case "ccn3":
					if strings.Contains(strings.ToLower(country.CCN3), val) {
						matchedParam = true
						break
					}

				case "capital":
					// country.Capital is a slice, so check partial match in any capital city
					for _, cap := range country.Capital {
						if strings.Contains(strings.ToLower(cap), val) {
							matchedParam = true
							break
						}
					}

				default:
					// Skip unrecognized keys instead of forcing a mismatch
					// matchedParam remains false for this query key
					// This means we IGNORE unrecognized keys, so the user can pass them
					// and they simply won't filter anything.
					continue
				}
			}

			// If we never found a match for this parameter, the country does not match
			if !matchedParam {
				continue OuterLoop
			}
		}

		// If all query parameters matched (some with OR logic among multi-values),
		// then this country is a result
		results = append(results, country)
	}
	return results
}

// searchAirports searches airports based on query parameters (partial match, case-insensitive, OR match for multi-values).
func searchAirports(queryParams url.Values) []Airport {
	var results []Airport

	for _, countryAirports := range AirportData {
	AirportLoop:
		for _, airport := range countryAirports.Airports {
			// For each airport, check all query parameters
			for key, values := range queryParams {
				matchedParam := false

				for _, val := range values {
					val = strings.TrimSpace(strings.ToLower(val))

					switch strings.ToLower(key) {
					case "name":
						if strings.Contains(strings.ToLower(airport.Name), val) {
							matchedParam = true
							break
						}
					case "municipality":
						if strings.Contains(strings.ToLower(airport.Municipality), val) {
							matchedParam = true
							break
						}
					case "ident":
						if strings.Contains(strings.ToLower(airport.Ident), val) {
							matchedParam = true
							break
						}
					case "iata_code":
						if strings.Contains(strings.ToLower(airport.IATACode), val) {
							matchedParam = true
							break
						}
					case "iso_country":
						if strings.Contains(strings.ToLower(airport.ISOCountry), val) {
							matchedParam = true
							break
						}
					case "iso_region":
						if strings.Contains(strings.ToLower(airport.ISORegion), val) {
							matchedParam = true
							break
						}
					case "airport_type":
						if strings.Contains(strings.ToLower(airport.Type), val) {
							matchedParam = true
							break
						}
					default:
						// Skip unrecognized keys
						continue
					}
				}

				// If this particular query parameter never matched, skip this airport
				if !matchedParam {
					continue AirportLoop
				}
			}

			// If we satisfied all query parameters (some with OR logic),
			// add this airport to the results
			results = append(results, airport)
		}
	}
	return results
}
