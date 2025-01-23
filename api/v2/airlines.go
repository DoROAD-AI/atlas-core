// airlines.go
package v2

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

const (
	airframesBaseURL = "http://www.airframes.org"
)

// Airline represents airline data retrieved from airframes.org
type Airline struct {
	ICAO        string `json:"icao"`
	IATA        string `json:"iata"`
	IATANum     string `json:"iata_num"`
	Name        string `json:"name"`
	Callsign    string `json:"callsign"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Website     string `json:"website"`
	Status      string `json:"status"`
	From        string `json:"from"`
	Until       string `json:"until"`
	Address     string `json:"address"`
	Remarks     string `json:"remarks"`
	Cert        string `json:"cert"`
}

// AirlineDetails represents detailed information about an airline
type AirlineDetails struct {
	ICAO         string            `json:"icao"`
	IATA         string            `json:"iata"`
	Name         string            `json:"name"`
	Callsign     string            `json:"callsign"`
	Country      string            `json:"country"`
	CountryCode  string            `json:"country_code"`
	Website      string            `json:"website"`
	Status       string            `json:"status"`
	From         string            `json:"from"`
	Until        string            `json:"until"`
	Address      string            `json:"address"`
	Remarks      string            `json:"remarks"`
	Cert         string            `json:"cert"`
	Fleet        []FleetEntry      `json:"fleet"`
	History      []HistoryEntry    `json:"history"`
	Accidents    []AccidentEntry   `json:"accidents"`
	OtherDetails map[string]string `json:"other_details"`
}

// FleetEntry represents an aircraft in the airline's fleet
type FleetEntry struct {
	AircraftType string `json:"aircraft_type"`
	Count        int    `json:"count"`
	Details      string `json:"details"`
}

// HistoryEntry represents a historical event for the airline
type HistoryEntry struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

// AccidentEntry represents an accident involving the airline
type AccidentEntry struct {
	Date     string `json:"date"`
	Aircraft string `json:"aircraft"`
	Location string `json:"location"`
	Details  string `json:"details"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Global variables for the authenticated client
var (
	airframesClient *http.Client
)

// init initializes the airframesClient with a cookie jar
func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to create cookie jar: %v", err))
	}
	airframesClient = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
}

// loginToAirframes logs into airframes.org to establish an authenticated session
func loginToAirframes() error {
	loginURL := airframesBaseURL + "/login"

	// Get credentials from environment variables
	username := os.Getenv("AIRFRAMES_USERNAME")
	password := os.Getenv("AIRFRAMES_PASSWORD")

	// Check if credentials are provided
	if username == "" || password == "" {
		return fmt.Errorf("AIRFRAMES_USERNAME and AIRFRAMES_PASSWORD must be set in the environment")
	}

	// Sanitize username and password
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	// Step 1: Perform a GET request to the login page to retrieve any necessary cookies or tokens
	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return fmt.Errorf("error creating GET request to login page: %w", err)
	}
	req.Header.Set("User-Agent", "AtlasAPI/1.0")
	resp, err := airframesClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing GET request to login page: %w", err)
	}
	defer resp.Body.Close()

	// Step 2: Prepare login form data using the field names from the login page
	formData := url.Values{
		"user1":   {username},
		"passwd1": {password},
		"submit":  {"Log in"},
	}

	// Create a POST request for login
	req, err = http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("error creating login request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)
	req.Header.Set("User-Agent", "AtlasAPI/1.0")

	// Perform the login request
	resp, err = airframesClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing login request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusFound &&
		resp.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// Verify login by accessing a page that requires authentication
	testURL := airframesBaseURL + "/airlines/"
	req, err = http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request to test login: %w", err)
	}
	req.Header.Set("User-Agent", "AtlasAPI/1.0")
	resp, err = airframesClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing test login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("test login page returned non-200 status code: %d", resp.StatusCode)
	}

	// Read the body to check for indicators of successful login
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing response to verify login: %w", err)
	}

	loginSuccessful := false
	if doc.Find("a[href='/logout']").Length() > 0 {
		loginSuccessful = true
	} else if doc.Find("small:contains('Logged in as')").Length() > 0 {
		loginSuccessful = true
	}

	if !loginSuccessful {
		return fmt.Errorf("login failed: unable to verify login success")
	}

	return nil
}

// ensureLoggedIn checks if we have a valid session and logs in if necessary
func ensureLoggedIn() error {
	// Attempt to access a protected page
	testURL := airframesBaseURL + "/airlines/"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request to test login: %w", err)
	}
	req.Header.Set("User-Agent", "AtlasAPI/1.0")
	resp, err := airframesClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing test login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Check if we are logged in
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err == nil {
			if doc.Find("a[href='/logout']").Length() > 0 ||
				doc.Find("small:contains('Logged in as')").Length() > 0 {
				// Already logged in
				return nil
			}
		}
	}

	// Not logged in, attempt to login
	if err := loginToAirframes(); err != nil {
		return fmt.Errorf("failed to log in to airframes.org: %w", err)
	}
	return nil
}

// searchAirframes performs a search on airframes.org based on the provided parameters
func searchAirframes(params url.Values) ([]Airline, error) {
	// Ensure we are logged in before making the request
	if err := ensureLoggedIn(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", airframesBaseURL+"/airlines/", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set the Content-Type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request using the dedicated client
	resp, err := airframesClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to airframes.org: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If unauthorized, try logging in again and retrying once
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if err := loginToAirframes(); err != nil {
				return nil, fmt.Errorf("authentication failed: %w", err)
			}
			// Retry the request
			resp, err = airframesClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error retrying request to airframes.org: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("airframes.org returned non-200 status code after retry: %d", resp.StatusCode)
			}
		} else {
			return nil, fmt.Errorf("airframes.org returned non-200 status code: %d", resp.StatusCode)
		}
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing airframes.org response: %w", err)
	}

	var airlines []Airline
	// Find the main results table
	table := doc.Find("table").First()

	// Parse each row in the table
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		// Skip the header row
		if row.Find("th").Length() > 0 {
			return
		}

		// Check if this is a detail row (address, remarks, cert)
		if row.Find("td[colspan='3']").Length() > 0 || row.Find("td[colspan='7']").Length() > 0 || row.Find("td[colspan='10']").Length() > 0 {
			// For safety, we’ll capture them all in case the site changes colspan in the future
			if len(airlines) > 0 {
				lastAirline := &airlines[len(airlines)-1]
				label := strings.TrimSpace(row.Find("td[align='right']").Text())
				value := ""

				// Depending on how many columns are spanned, check them
				if row.Find("td[colspan='7']").Length() > 0 {
					value = strings.TrimSpace(row.Find("td[colspan='7']").Text())
				} else if row.Find("td[colspan='3']").Length() > 0 {
					value = strings.TrimSpace(row.Find("td[colspan='3']").Text())
				} else if row.Find("td[colspan='10']").Length() > 0 {
					value = strings.TrimSpace(row.Find("td[colspan='10']").Text())
				}

				switch label {
				case "Address:":
					lastAirline.Address = value
				case "Remarks:":
					lastAirline.Remarks = value
				case "Cert.:":
					lastAirline.Cert = value
				}
			}
			return
		}

		// This is a regular airline data row
		var airline Airline
		cells := row.Find("td")

		// If it doesn't have at least 10 columns, skip
		if cells.Length() < 10 {
			return
		}

		cells.Each(func(j int, cell *goquery.Selection) {
			switch j {
			case 0:
				airline.ICAO = strings.TrimSpace(cell.Find("a").Text())
			case 1:
				airline.IATA = strings.TrimSpace(cell.Text())
			case 2:
				airline.IATANum = strings.TrimSpace(cell.Text())
			case 3:
				airline.Name = strings.TrimSpace(cell.Find("b").Text())
			case 4:
				airline.Callsign = strings.TrimSpace(cell.Text())
			case 5:
				// Sometimes the country code is inside a separate text node, and the flag alt is the real code
				airline.CountryCode = strings.TrimSpace(cell.Text())
				flagAlt := cell.Find("img").AttrOr("alt", "")
				if flagAlt != "" && flagAlt != airline.CountryCode {
					// Some rows might have alt=TT and text=TT, or alt=TT and text="TT " with spaces
					airline.Country = flagAlt
				} else {
					airline.Country = airline.CountryCode
				}
			case 6:
				airline.Website = strings.TrimSpace(cell.Find("a").AttrOr("href", ""))
			case 7:
				airline.Status = strings.TrimSpace(cell.Text())
			case 8:
				airline.From = strings.TrimSpace(cell.Text())
			case 9:
				airline.Until = strings.TrimSpace(cell.Text())
			}
		})

		// Only append if we got an ICAO or name
		if airline.ICAO != "" || airline.Name != "" {
			airlines = append(airlines, airline)
		}
	})

	return airlines, nil
}

// fetchAirlineDetails fetches detailed information about a specific airline from its individual page
func fetchAirlineDetails(icao string) (*AirlineDetails, error) {
	// Ensure we are logged in before making the request
	if err := ensureLoggedIn(); err != nil {
		return nil, err
	}

	airlineURL := fmt.Sprintf("%s/fleet/%s", airframesBaseURL, strings.ToLower(icao))

	req, err := http.NewRequest("GET", airlineURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Perform the request using the dedicated client
	resp, err := airframesClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to airframes.org: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If unauthorized, try logging in again and retrying once
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if err := loginToAirframes(); err != nil {
				return nil, fmt.Errorf("authentication failed: %w", err)
			}
			// Retry the request
			resp, err = airframesClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error retrying request to airframes.org: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("airframes.org returned non-200 status code after retry: %d", resp.StatusCode)
			}
		} else {
			return nil, fmt.Errorf("airframes.org returned non-200 status code: %d", resp.StatusCode)
		}
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing airframes.org response: %w", err)
	}

	details := &AirlineDetails{
		ICAO:         icao,
		Fleet:        []FleetEntry{},
		History:      []HistoryEntry{},
		Accidents:    []AccidentEntry{},
		OtherDetails: make(map[string]string),
	}

	// Extract basic information from the top table
	topTable := doc.Find("#content table").First()
	if topTable.Length() == 0 {
		// Possibly no data found
		return details, nil
	}

	topTable.Find("tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			// Header row, extract IATA and Name
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				switch j {
				case 0:
					details.IATA = strings.TrimSpace(cell.Text())
				case 1:
					details.Name = strings.TrimSpace(cell.Find("b").Text())
				}
			})
		} else {
			label := strings.TrimSpace(row.Find("td").First().Text())
			valueCell := row.Find("td").Eq(1)
			value := strings.TrimSpace(valueCell.Text())

			switch label {
			case "Callsign:":
				details.Callsign = value
			case "Country:":
				// country code is sometimes text, or alt from an <img>
				details.CountryCode = value
				flagAlt := valueCell.Find("img").AttrOr("alt", "")
				if flagAlt != "" && flagAlt != details.CountryCode {
					details.Country = flagAlt
				} else {
					details.Country = details.CountryCode
				}
			case "Website:":
				details.Website = valueCell.Find("a").AttrOr("href", "")
			case "Status:":
				details.Status = value
			case "From:":
				details.From = value
			case "Until:":
				details.Until = value
			case "Address:":
				details.Address = value
			case "Remarks:":
				details.Remarks = value
			case "Cert.:":
				details.Cert = value
			default:
				if label != "" {
					details.OtherDetails[label] = value
				}
			}
		}
	})

	// There may be subsequent tables for fleet, history, accidents, etc.
	tables := doc.Find("#content table")
	if tables.Length() > 1 {
		// Extract fleet (2nd table)
		fleetTable := tables.Eq(1)
		fleetTable.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // Skip header row
			}
			var fleetEntry FleetEntry
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				switch j {
				case 0:
					fleetEntry.AircraftType = strings.TrimSpace(cell.Text())
				case 1:
					fmt.Sscan(strings.TrimSpace(cell.Text()), &fleetEntry.Count)
				case 2:
					fleetEntry.Details = strings.TrimSpace(cell.Text())
				}
			})
			if fleetEntry.AircraftType != "" {
				details.Fleet = append(details.Fleet, fleetEntry)
			}
		})
	}

	if tables.Length() > 2 {
		// Extract history (3rd table)
		historyTable := tables.Eq(2)
		historyTable.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // Skip header row
			}
			var historyEntry HistoryEntry
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				switch j {
				case 0:
					historyEntry.Date = strings.TrimSpace(cell.Text())
				case 1:
					historyEntry.Description = strings.TrimSpace(cell.Text())
				}
			})
			if historyEntry.Date != "" {
				details.History = append(details.History, historyEntry)
			}
		})
	}

	if tables.Length() > 3 {
		// Extract accidents (4th table)
		accidentTable := tables.Eq(3)
		accidentTable.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // Skip header row
			}
			var accidentEntry AccidentEntry
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				switch j {
				case 0:
					accidentEntry.Date = strings.TrimSpace(cell.Text())
				case 1:
					accidentEntry.Aircraft = strings.TrimSpace(cell.Text())
				case 2:
					accidentEntry.Location = strings.TrimSpace(cell.Text())
				case 3:
					accidentEntry.Details = strings.TrimSpace(cell.Text())
				}
			})
			if accidentEntry.Date != "" {
				details.Accidents = append(details.Accidents, accidentEntry)
			}
		})
	}

	return details, nil
}

// GetAirlineDetails handles GET requests to /airlines/{icao}/details
// NOTE: for your main.go route usage "/:icao/details", the final path is /v2/airlines/BWA/details
// @Summary Get detailed airline information
// @Description Retrieves detailed information about a specific airline, including fleet, history, and accidents.
// @Tags Airlines
// @Accept json
// @Produce json
// @Param icao path string true "ICAO code of the airline (e.g. 'BAW')"
// @Success 200 {object} AirlineDetails
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /airlines/{icao}/details [get]
func GetAirlineDetails(c *gin.Context) {
	icao := strings.ToUpper(c.Param("icao"))
	if icao == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ICAO code is required"})
		return
	}

	airlineDetails, err := fetchAirlineDetails(icao)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, airlineDetails)
}

// GetAirlinesByICAO handles GET requests to /airlines/icao/:icao
// @Summary Get airline by ICAO code
// @Description Retrieves airline information based on the ICAO code.
// @Tags Airlines
// @Accept json
// @Produce json
// @Param icao path string true "ICAO code of the airline"
// @Success 200 {array} Airline
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /airlines/icao/{icao} [get]
func GetAirlinesByICAO(c *gin.Context) {
	icao := strings.ToUpper(c.Param("icao"))
	if icao == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ICAO code is required"})
		return
	}

	params := url.Values{}
	params.Set("icao", icao)
	params.Set("submit", "submit")

	airlines, err := searchAirframes(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, airlines)
}

// GetAirlinesByIATA handles GET requests to /airlines/iata/:iata
// @Summary Get airline by IATA code
// @Description Retrieves airline information based on the IATA code.
// @Tags Airlines
// @Accept json
// @Produce json
// @Param iata path string true "IATA code of the airline"
// @Success 200 {array} Airline
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /airlines/iata/{iata} [get]
func GetAirlinesByIATA(c *gin.Context) {
	iata := strings.ToUpper(c.Param("iata"))
	if iata == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "IATA code is required"})
		return
	}

	params := url.Values{}
	params.Set("iata", iata)
	params.Set("submit", "submit")

	airlines, err := searchAirframes(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, airlines)
}

// GetAirlinesByName handles GET requests to /airlines/name/:name
// @Summary Get airlines by name
// @Description Retrieves airline information based on the name or callsign.
// @Tags Airlines
// @Accept json
// @Produce json
// @Param name path string true "Name or callsign of the airline"
// @Success 200 {array} Airline
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /airlines/name/{name} [get]
func GetAirlinesByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Airline name is required"})
		return
	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("submit", "submit")

	airlines, err := searchAirframes(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, airlines)
}
