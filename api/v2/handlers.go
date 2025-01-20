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
