// handlers.go - handlers for the v2 API.
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// PassportData defines the structure for passports data
type PassportData map[string]map[string]string

// Passports holds the passports data
var Passports PassportData

// LoadPassportData reads local JSON data into the global Passports variable.
func LoadPassportData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read passports file: %w", err)
	}
	if err := json.Unmarshal(data, &Passports); err != nil {
		return fmt.Errorf("failed to parse passports data: %w", err)
	}
	return nil
}

// VisaRequirement represents the visa requirement between two countries.
type VisaRequirement struct {
	From        string `json:"from" example:"USA"`
	To          string `json:"to" example:"DEU"`
	Requirement string `json:"requirement" example:"90"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message" example:"Bad request"`
}

// GetPassportData handles GET /passports/:passportCode
// @Summary     Get passport data
// @Description Get visa requirement data for a specific passport.
// @Tags        Passports
// @Accept      json
// @Produce     json
// @Param       passportCode path string true "Passport code (e.g., USA)"
// @Success     200 {object} map[string]interface{}
// @Failure     404 {object} ErrorResponse
// @Router      /passports/{passportCode} [get]
func GetPassportData(c *gin.Context) {
	passportCode := strings.ToUpper(c.Param("passportCode"))
	visaRules, ok := Passports[passportCode]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found"})
		return
	}
	passportData := map[string]interface{}{
		"passport": passportCode,
		"visas":    visaRules,
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
// @Success     200 {object} map[string]interface{}
// @Failure     404 {object} ErrorResponse
// @Router      /passports/{passportCode}/visas [get]
func GetVisaRequirementsForPassport(c *gin.Context) {
	passportCode := strings.ToUpper(c.Param("passportCode"))
	visaRules, ok := Passports[passportCode]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found"})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"passport": passportCode,
		"visas":    visaRules,
	})
}

// GetVisaRequirements godoc
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
	fromCountry := strings.ToUpper(c.Query("fromCountry"))
	toCountry := strings.ToUpper(c.Query("toCountry"))

	if fromCountry == "" || toCountry == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "fromCountry and toCountry query parameters are required"})
		return
	}

	visaRules, ok := Passports[fromCountry]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Passport data not found for origin country"})
		return
	}
	requirement, ok := visaRules[toCountry]
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Visa requirement data not found for this country pair"})
		return
	}
	c.JSON(http.StatusOK, VisaRequirement{
		From:        fromCountry,
		To:          toCountry,
		Requirement: requirement,
	})
}
