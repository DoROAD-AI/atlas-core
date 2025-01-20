// handlers.go - handlers for the v2 API.
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	v1 "github.com/DoROAD-AI/atlas/api/v1" // Import v1 to access Countries data
)

// PassportData defines the structure for passports data
type PassportData map[string]map[string]string

// Passports holds the passports data
var Passports PassportData

// codeToCCA3 maps various country codes to their CCA3 code
var codeToCCA3 map[string]string

// VisaRequirement represents the visa requirement between two countries.
// @Description VisaRequirement represents the visa requirement between two countries.
type VisaRequirement struct {
	From        string `json:"from" example:"USA"`
	To          string `json:"to" example:"DEU"`
	Requirement string `json:"requirement" example:"90"`
}

// ErrorResponse represents an error response.
// @Description ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message" example:"Bad request"`
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

// AirportData holds the airport data
var AirportData map[string]CountryAirports

// Initialize code mapping after loading passports data
func init() {
	// Ensure that codeToCCA3 is initialized
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

// LoadAirportData loads airport data from a JSON file.
func LoadAirportData(filename string) error {
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

// GetPassportData handles GET /passports/:passportCode
// @Summary     Get passport data
// @Description Get visa requirement data for a specific passport.
// @Tags        Passports
// @Accept      json
// @Produce     json
// @Param       passportCode path string true "Passport code (e.g., USA)"
// @Success     200 {object} PassportResponse
// @Failure     404 {object} ErrorResponse
// @Router      /passports/{passportCode} [get]
func GetPassportData(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Invalid passport country code"})
		return
	}
	visaRules, ok := Passports[passportCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found"})
		return
	}
	passportData := PassportResponse{
		Passport: passportCodeInput,
		Visas:    visaRules,
	}
	c.JSON(http.StatusOK, passportData)
}

// GetVisaRequirementsForPassport handles GET /passports/:passportCode/visas
// @Summary     Get visa requirements for a passport
// @Description Get visa requirements for all destinations for a specific passport.
// @Tags        Passports
// @Accept      json
// @Produce     json
// @Param       passportCode path string true "Passport code (e.g., USA)"
// @Success     200 {object} PassportResponse
// @Failure     404 {object} ErrorResponse
// @Router      /passports/{passportCode}/visas [get]
func GetVisaRequirementsForPassport(c *gin.Context) {
	passportCodeInput := strings.ToUpper(c.Param("passportCode"))
	passportCCA3, ok := codeToCCA3[passportCodeInput]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Invalid passport country code"})
		return
	}
	visaRules, ok := Passports[passportCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found"})
		return
	}
	passportData := PassportResponse{
		Passport: passportCodeInput,
		Visas:    visaRules,
	}
	c.JSON(http.StatusOK, passportData)
}

// GetVisaRequirements handles GET /passports/visa
// @Summary     Get visa requirements between two countries
// @Description Get visa requirements for a passport holder from one country traveling to another.
// @Tags        Passports
// @Accept      json
// @Produce     json
// @Param       fromCountry query string true "Origin country code (e.g., USA)"
// @Param       toCountry   query string true "Destination country code (e.g., DEU)"
// @Success     200 {object} VisaRequirement
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /passports/visa [get]
func GetVisaRequirements(c *gin.Context) {
	fromCountryInput := strings.ToUpper(c.Query("fromCountry"))
	toCountryInput := strings.ToUpper(c.Query("toCountry"))

	if fromCountryInput == "" || toCountryInput == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "fromCountry and toCountry query parameters are required"})
		return
	}

	fromCountryCCA3, ok := codeToCCA3[fromCountryInput]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Invalid fromCountry code"})
		return
	}
	toCountryCCA3, ok := codeToCCA3[toCountryInput]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Invalid toCountry code"})
		return
	}

	visaRules, ok := Passports[fromCountryCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found for origin country"})
		return
	}
	requirement, ok := visaRules[toCountryCCA3]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Visa requirement data not found for this country pair"})
		return
	}
	c.JSON(http.StatusOK, VisaRequirement{
		From:        fromCountryInput,
		To:          toCountryInput,
		Requirement: requirement,
	})
}

// PassportResponse represents the passport data response.
// @Description PassportResponse represents the passport data response.
type PassportResponse struct {
	Passport string            `json:"passport" example:"USA"`
	Visas    map[string]string `json:"visas"`
}

// GetAirports handles GET /airports
// @Summary     Get all airports
// @Description Retrieves a list of all airports.
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Success     200 {array} CountryAirports
// @Failure     500 {object} ErrorResponse
// @Router      /airports [get]
func GetAirports(c *gin.Context) {
	c.JSON(http.StatusOK, AirportData)
}

// GetAirportByCode handles GET /airports/:code
// @Summary     Get airport by code
// @Description Retrieves an airport by its IATA or ICAO code.
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Param       code path string true "Airport code (IATA or ICAO)"
// @Success     200 {object} Airport
// @Failure     404 {object} ErrorResponse
// @Router      /airports/{code} [get]
func GetAirportByCode(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))

	for _, countryAirports := range AirportData {
		for _, airport := range countryAirports.Airports {
			if airport.IATACode == code || airport.Ident == code {
				c.JSON(http.StatusOK, airport)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, ErrorResponse{Message: "Airport not found"})
}

// GetAirportsByCountry handles GET /airports/country/:countryCode
// @Summary     Get airports by country
// @Description Retrieves all airports in a specific country.
// @Tags        Airports
// @Accept      json
// @Produce     json
// @Param       countryCode path string true "Country code (e.g., VC)"
// @Success     200 {object} CountryAirports
// @Failure     404 {object} ErrorResponse
// @Router      /airports/country/{countryCode} [get]
func GetAirportsByCountry(c *gin.Context) {
	countryCode := strings.ToUpper(c.Param("countryCode"))

	if countryAirports, ok := AirportData[countryCode]; ok {
		c.JSON(http.StatusOK, countryAirports)
	} else {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Country not found or no airports available"})
	}
}
