// airlines.go - handlers for airline data in the v2 API.
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/DoROAD-AI/atlas/types"
	"github.com/gin-gonic/gin"
)

// Airline represents data for a single airline.
// @Description Airline represents data for a single airline.
type Airline struct {
	AirlineID int    `json:"airline_id" example:"28"`
	Name      string `json:"name" example:"Asiana Airlines"`
	Alias     string `json:"alias" example:"\\N"`
	IATA      string `json:"iata" example:"OZ"`
	ICAO      string `json:"icao" example:"AAR"`
	Callsign  string `json:"callsign" example:"ASIANA"`
	Country   string `json:"country" example:"Republic of Korea"`
	Active    string `json:"active" example:"Y"`
}

// Airlines holds the airline data.
var Airlines []Airline

// LoadAirlinesData loads airline data from a JSON file.
func LoadAirlinesData(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read airlines file: %w", err)
	}
	if err := json.Unmarshal(data, &Airlines); err != nil {
		return fmt.Errorf("failed to parse airlines data: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Airline Handlers
// ----------------------------------------------------------------------------

// GetAllAirlines handles GET /v2/airlines
// @Summary     Get all airlines
// @Description Retrieves a list of all airlines.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Success     200 {array} Airline
// @Failure     500 {object} ErrorResponse
// @Router      /airlines [get]
func GetAllAirlines(c *gin.Context) {
	c.JSON(http.StatusOK, Airlines)
}

// GetAirlineByID handles GET /v2/airlines/id/{airlineID}
// @Summary     Get airline by ID
// @Description Retrieves a specific airline by its ID.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Param       airlineID path int true "Airline ID"
// @Success     200 {object} Airline
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/id/{airlineID} [get]
func GetAirlineByID(c *gin.Context) {
	airlineID, err := parseIntParam(c, "airlineID")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid Airline ID"})
		return
	}

	for _, airline := range Airlines {
		if airline.AirlineID == airlineID {
			c.JSON(http.StatusOK, airline)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airline not found"})
}

func parseIntParam(c *gin.Context, s string) (any, any) {
	panic("unimplemented")
}

// GetAirlinesByCountry handles GET /v2/airlines/country/{countryName}
// @Summary     Get airlines by country
// @Description Retrieves all airlines based in a specific country.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Param       countryName path string true "Country name"
// @Success     200 {array} Airline
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/country/{countryName} [get]
func GetAirlinesByCountry(c *gin.Context) {
	countryName := c.Param("countryName")
	var matchingAirlines []Airline

	for _, airline := range Airlines {
		if strings.EqualFold(airline.Country, countryName) {
			matchingAirlines = append(matchingAirlines, airline)
		}
	}

	if len(matchingAirlines) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airlines found for this country"})
		return
	}

	c.JSON(http.StatusOK, matchingAirlines)
}

// GetAirlineByICAO handles GET /v2/airlines/icao/{icaoCode}
// @Summary     Get airline by ICAO code
// @Description Retrieves an airline by its ICAO code.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Param       icaoCode path string true "ICAO code"
// @Success     200 {object} Airline
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/icao/{icaoCode} [get]
func GetAirlineByICAO(c *gin.Context) {
	icaoCode := strings.ToUpper(c.Param("icaoCode"))

	for _, airline := range Airlines {
		if airline.ICAO == icaoCode {
			c.JSON(http.StatusOK, airline)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airline not found for this ICAO code"})
}

// GetAirlineByIATA handles GET /v2/airlines/iata/{iataCode}
// @Summary     Get airline by IATA code
// @Description Retrieves an airline by its IATA code.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Param       iataCode path string true "IATA code"
// @Success     200 {object} Airline
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/iata/{iataCode} [get]
func GetAirlineByIATA(c *gin.Context) {
	iataCode := strings.ToUpper(c.Param("iataCode"))

	for _, airline := range Airlines {
		if airline.IATA == iataCode {
			c.JSON(http.StatusOK, airline)
			return
		}
	}

	c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Airline not found for this IATA code"})
}

// GetActiveAirlines handles GET /v2/airlines/active
// @Summary     Get active airlines
// @Description Retrieves all airlines that are currently active.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Success     200 {array} Airline
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/active [get]
func GetActiveAirlines(c *gin.Context) {
	var activeAirlines []Airline

	for _, airline := range Airlines {
		if airline.Active == "Y" {
			activeAirlines = append(activeAirlines, airline)
		}
	}

	if len(activeAirlines) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No active airlines found"})
		return
	}

	c.JSON(http.StatusOK, activeAirlines)
}

// SearchAirlines handles GET /v2/airlines/search?query={searchString}
// @Summary     Search airlines
// @Description Performs a flexible search for airlines based on a query string.
// @Tags        Airlines
// @Accept      json
// @Produce     json
// @Param       query query string true "Search string (can match airline name, country, ICAO/IATA code, etc.)"
// @Success     200 {array} Airline
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /airlines/search [get]
func SearchAirlines(c *gin.Context) {
	searchString := strings.ToLower(c.Query("query"))
	if searchString == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Query parameter 'query' is required"})
		return
	}

	var matchingAirlines []Airline
	for _, airline := range Airlines {
		if strings.Contains(strings.ToLower(airline.Name), searchString) ||
			strings.Contains(strings.ToLower(airline.Alias), searchString) ||
			strings.Contains(strings.ToLower(airline.Country), searchString) ||
			strings.Contains(strings.ToLower(airline.ICAO), searchString) ||
			strings.Contains(strings.ToLower(airline.IATA), searchString) ||
			strings.Contains(strings.ToLower(airline.Callsign), searchString) {
			matchingAirlines = append(matchingAirlines, airline)
		}
	}

	if len(matchingAirlines) == 0 {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No airlines found matching the search criteria"})
		return
	}

	c.JSON(http.StatusOK, matchingAirlines)
}
