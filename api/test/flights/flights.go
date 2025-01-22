// flights.go - Flight tracking functionality for the Atlas API.

package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// =========================================================
// 1) Client, Config, and Types
// =========================================================

// FlightRadarClient holds the logic for interacting with Flightradar24.
type FlightRadarClient struct {
	HTTPClient  *http.Client
	User        string
	Password    string
	loggedIn    bool
	accessToken string
	cookies     []*http.Cookie
	mu          sync.Mutex // Ensure thread-safe access
}

// Flight represents a structure for a flight.
type Flight struct {
	ID           string `json:"id"`
	Callsign     string `json:"callsign,omitempty"`
	Registration string `json:"registration,omitempty"`
	AircraftType string `json:"aircraft_type,omitempty"`
	Origin       string `json:"origin,omitempty"`
	Destination  string `json:"destination,omitempty"`
	// Add more fields as needed
}

// AirportInfo represents basic information about an airport.
type AirportInfo struct {
	Code    string `json:"code"`
	Name    string `json:"name,omitempty"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
	// Add more fields as needed
}

// AirportDetails represents detailed information about an airport.
type AirportDetails struct {
	AirportInfo
	Runways   []string `json:"runways,omitempty"`
	Schedules []string `json:"schedules,omitempty"`
	// Add more fields as needed
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// =========================================================
// 2) Instantiate a global FlightRadarClient and Auto-Login
// =========================================================

var flightRadar *FlightRadarClient

func init() {
	// Read environment variables
	user := os.Getenv("FLIGHTRADAR_USER")
	pass := os.Getenv("FLIGHTRADAR_PASS")
	flightRadar = &FlightRadarClient{
		User:     user,
		Password: pass,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	// Optionally: Auto-login if credentials are present
	if user != "" && pass != "" {
		if err := flightRadar.Login(); err != nil {
			fmt.Printf("[FlightRadar] Auto-login failed: %v\n", err)
		} else {
			fmt.Println("[FlightRadar] Auto-login successful")
		}
	}
}

// =========================================================
// 3) Implement Core Client Methods (Login, Logout, Request)
// =========================================================

// Login attempts to authenticate with Flightradar24 using the client’s credentials.
func (c *FlightRadarClient) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loggedIn {
		return nil // Already logged in
	}
	if c.User == "" || c.Password == "" {
		return errors.New("flightradar user/password environment variables not set")
	}
	loginURL := "https://www.flightradar24.com/user/login"
	data := url.Values{
		"email":    {c.User},
		"password": {c.Password},
		"remember": {"true"},
		"type":     {"web"},
	}
	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-FlightRadar24-Client/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Expecting 200 OK
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login request failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	// Save cookies
	c.cookies = resp.Cookies()

	// Parse JSON response for an access token
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var dataResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dataResp); err != nil {
		return err
	}
	if success, ok := dataResp["success"].(bool); ok && success {
		userData, ok := dataResp["userData"].(map[string]interface{})
		if ok {
			if accessToken, ok := userData["token"].(string); ok {
				c.accessToken = accessToken
			}
		}
		c.loggedIn = true
		return nil
	}
	return errors.New("login failed: unexpected response")
}

// Logout logs out of the FlightRadar24 account.
func (c *FlightRadarClient) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.loggedIn {
		return nil
	}
	logoutURL := "https://www.flightradar24.com/user/logout"
	req, err := http.NewRequest(http.MethodGet, logoutURL, nil)
	if err != nil {
		return err
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout request failed: status=%d", resp.StatusCode)
	}
	c.loggedIn = false
	c.cookies = nil
	c.accessToken = ""
	return nil
}

// doRequest is a helper that attaches cookies, default headers, and handles requests.
func (c *FlightRadarClient) doRequest(method, urlStr string, params map[string]string) ([]byte, int, error) {
	reqBody := io.Reader(nil)
	if method == http.MethodPost || method == http.MethodPut {
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}
		reqBody = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, 0, err
	}

	// Add cookies if we have them
	c.mu.Lock()
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	c.mu.Unlock()

	// Set headers
	req.Header.Set("User-Agent", "Go-FlightRadar24-Client/1.0")
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Add query parameters
	if method == http.MethodGet && params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

// =========================================================
// 4) Implement FlightRadar24 Methods
// =========================================================

// GetFlights retrieves flights based on provided filters.
func (c *FlightRadarClient) GetFlights(airline, registration, aircraftType string) ([]Flight, error) {
	urlStr := "https://data-cloud.flightradar24.com/zones/fcgi/feed.js"
	params := map[string]string{
		"faa":       "1",
		"satellite": "1",
		"mlat":      "1",
		"flarm":     "1",
		"adsb":      "1",
		"gnd":       "1",
		"air":       "1",
		"vehicles":  "1",
		"estimated": "1",
		"maxage":    "14400",
		"gliders":   "1",
		"stats":     "1",
		"limit":     "5000",
	}
	if airline != "" {
		params["airline"] = airline
	}
	if registration != "" {
		params["reg"] = registration
	}
	if aircraftType != "" {
		params["type"] = aircraftType
	}

	body, status, err := c.doRequest(http.MethodGet, urlStr, params)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to get flights (status %d): %s", status, string(body))
	}

	// Parse the response
	var flightsMap map[string]interface{}
	if err := json.Unmarshal(body, &flightsMap); err != nil {
		return nil, err
	}

	var flights []Flight
	for key, val := range flightsMap {
		// Skip non-flight entries
		if len(key) == 0 || key[0] < '0' || key[0] > '9' {
			continue
		}
		flightData, ok := val.([]interface{})
		if !ok || len(flightData) < 15 {
			continue
		}

		flight := Flight{
			ID:           key,
			Callsign:     flightData[16].(string),
			Registration: flightData[9].(string),
			AircraftType: flightData[8].(string),
			Origin:       flightData[11].(string),
			Destination:  flightData[12].(string),
		}
		flights = append(flights, flight)
	}
	return flights, nil
}

// GetFlightDetails retrieves detailed information about a flight.
func (c *FlightRadarClient) GetFlightDetails(flightID string) (map[string]interface{}, error) {
	urlStr := fmt.Sprintf("https://data-live.flightradar24.com/clickhandler/?version=1.5&flight=%s", flightID)
	body, status, err := c.doRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to get flight details (status %d): %s", status, string(body))
	}

	var details map[string]interface{}
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}
	return details, nil
}

// GetAirport retrieves basic information about an airport.
func (c *FlightRadarClient) GetAirport(code string) (*AirportInfo, error) {
	urlStr := fmt.Sprintf("https://www.flightradar24.com/airports/traffic-stats/?airport=%s", code)
	body, status, err := c.doRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to get airport info (status %d): %s", status, string(body))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	airportData, ok := data["details"].(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected airport data format")
	}

	airport := &AirportInfo{
		Code:    code,
		Name:    airportData["name"].(string),
		City:    airportData["city"].(string),
		Country: airportData["country"].(string),
	}
	return airport, nil
}

// GetAirportDetails retrieves detailed information about an airport.
func (c *FlightRadarClient) GetAirportDetails(code string) (*AirportDetails, error) {
	urlStr := "https://api.flightradar24.com/common/v1/airport.json"
	params := map[string]string{
		"code":   code,
		"plugin": "sched_prev,sched_next,details",
	}

	body, status, err := c.doRequest(http.MethodGet, urlStr, params)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to get airport details (status %d): %s", status, string(body))
	}

	var result struct {
		Result struct {
			Response struct {
				Airport struct {
					PluginData struct {
						Details map[string]interface{}   `json:"details"`
						Runways []map[string]interface{} `json:"runways"`
					} `json:"pluginData"`
				} `json:"airport"`
			} `json:"response"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	details := result.Result.Response.Airport.PluginData.Details
	name, _ := details["name"].(string)
	city, _ := details["city"].(string)
	country, _ := details["country"].(string)

	var runways []string
	for _, rw := range result.Result.Response.Airport.PluginData.Runways {
		runwayName, _ := rw["name"].(string)
		runways = append(runways, runwayName)
	}

	airportDetails := &AirportDetails{
		AirportInfo: AirportInfo{
			Code:    code,
			Name:    name,
			City:    city,
			Country: country,
		},
		Runways: runways,
	}

	return airportDetails, nil
}

// =========================================================
// 5) Handlers (Gin) for the new routes
// =========================================================

// GetFlightsHandler handles GET /v2/flights
func GetFlightsHandler(c *gin.Context) {
	airline := c.Query("airline")
	registration := c.Query("registration")
	aircraftType := c.Query("aircraftType")
	flights, err := flightRadar.GetFlights(airline, registration, aircraftType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, flights)
}

// GetFlightDetailsHandler handles GET /v2/flights/:flightID
func GetFlightDetailsHandler(c *gin.Context) {
	flightID := c.Param("flightID")
	details, err := flightRadar.GetFlightDetails(flightID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// GetAirportHandler handles GET /v2/flights/airports/:code
func GetAirportHandler(c *gin.Context) {
	code := c.Param("code")
	info, err := flightRadar.GetAirport(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetAirportDetailsHandler handles GET /v2/flights/airports/:code/details
func GetAirportDetailsHandler(c *gin.Context) {
	code := c.Param("code")
	details, err := flightRadar.GetAirportDetails(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// LoginHandler handles GET /v2/flights/login
func LoginHandler(c *gin.Context) {
	err := flightRadar.Login()
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
}

// LogoutHandler handles GET /v2/flights/logout
func LogoutHandler(c *gin.Context) {
	err := flightRadar.Logout()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
