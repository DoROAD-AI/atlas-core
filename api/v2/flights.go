package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DoROAD-AI/atlas/types"

	"github.com/gin-gonic/gin"
)

//=====================================================
// 1) Data Structures
//=====================================================

// StateVector represents the state of a vehicle at a particular time.
type StateVector struct {
	ICAO24         string   `json:"icao24" example:"48585773"`
	Callsign       string   `json:"callsign,omitempty" example:"SVA35"`
	OriginCountry  string   `json:"origin_country,omitempty" example:"Saudi Arabia"`
	TimePosition   *int     `json:"time_position,omitempty" example:"1674345600"`
	LastContact    *int     `json:"last_contact,omitempty" example:"1674345600"`
	Longitude      *float64 `json:"longitude,omitempty" example:"-73.778925"`
	Latitude       *float64 `json:"latitude,omitempty" example:"40.641766"`
	BaroAltitude   *float64 `json:"baro_altitude,omitempty" example:"11277.6"`
	OnGround       bool     `json:"on_ground,omitempty" example:"false"`
	Velocity       *float64 `json:"velocity,omitempty" example:"245.34"`
	TrueTrack      *float64 `json:"true_track,omitempty" example:"285.7"`
	VerticalRate   *float64 `json:"vertical_rate,omitempty" example:"0"`
	Sensors        []int    `json:"sensors,omitempty"`
	GeoAltitude    *float64 `json:"geo_altitude,omitempty" example:"11887.2"`
	Squawk         string   `json:"squawk,omitempty" example:"2200"`
	SPI            bool     `json:"spi,omitempty" example:"false"`
	PositionSource int      `json:"position_source,omitempty" example:"0"`
	Category       *int     `json:"category,omitempty" example:"1"`
}

// OpenSkyStates represents the state of the airspace at a given time.
type OpenSkyStates struct {
	Time   int           `json:"time" example:"1674345600"`
	States []StateVector `json:"states"`
}

// FlightData represents data about a single flight.
type FlightData struct {
	ICAO24                           string  `json:"icao24" example:"48585773"`
	FirstSeen                        int     `json:"firstSeen" example:"1674345600"`
	EstDepartureAirport              *string `json:"estDepartureAirport,omitempty" example:"RUH"`
	LastSeen                         int     `json:"lastSeen" example:"1674345600"`
	EstArrivalAirport                *string `json:"estArrivalAirport,omitempty" example:"JFK"`
	Callsign                         *string `json:"callsign,omitempty" example:"SVA35"`
	EstDepartureAirportHorizDistance *int    `json:"estDepartureAirportHorizDistance,omitempty" example:"1000"`
	EstDepartureAirportVertDistance  *int    `json:"estDepartureAirportVertDistance,omitempty" example:"500"`
	EstArrivalAirportHorizDistance   *int    `json:"estArrivalAirportHorizDistance,omitempty" example:"2000"`
	EstArrivalAirportVertDistance    *int    `json:"estArrivalAirportVertDistance,omitempty" example:"1000"`
	DepartureAirportCandidatesCount  *int    `json:"departureAirportCandidatesCount,omitempty" example:"5"`
	ArrivalAirportCandidatesCount    *int    `json:"arrivalAirportCandidatesCount,omitempty" example:"3"`
}

// Waypoint represents a single waypoint in a flight trajectory.
type Waypoint struct {
	Time         int      `json:"time" example:"1674345600"`
	Latitude     *float64 `json:"latitude,omitempty" example:"40.7789"`
	Longitude    *float64 `json:"longitude,omitempty" example:"-73.9692"`
	BaroAltitude *float64 `json:"baro_altitude,omitempty" example:"10000"`
	TrueTrack    *float64 `json:"true_track,omitempty" example:"270.0"`
	OnGround     bool     `json:"on_ground" example:"false"`
}

// FlightTrack represents the trajectory for a certain aircraft.
type FlightTrack struct {
	Icao24    string     `json:"icao24" example:"48585773"`
	StartTime int        `json:"startTime" example:"1674345600"`
	EndTime   int        `json:"endTime" example:"1674349200"`
	Callsign  *string    `json:"callsign,omitempty" example:"SVA35"`
	Path      []Waypoint `json:"path"`
}

//=====================================================
// 2) Client Config + Global Client
//=====================================================

type Config struct {
	Username    string
	Password    string
	BaseURL     string
	HTTPTimeout time.Duration
}

type OpenSkyClient struct {
	HTTPClient   *http.Client
	Username     string
	Password     string
	BaseURL      string
	mu           sync.Mutex
	lastRequests map[string]time.Time // track last request times for rate limiting
}

// Global instance used by handlers
var openSkyApi *OpenSkyClient

// InitializeOpenSkyClient sets up the global instance (called from main.go).
func InitializeOpenSkyClient(username, password string) {
	config := Config{
		Username:    username,
		Password:    password,
		BaseURL:     "https://opensky-network.org/api",
		HTTPTimeout: 15 * time.Second,
	}
	openSkyApi = NewOpenSkyClient(config)
}

// NewOpenSkyClient creates a new OpenSkyClient instance.
func NewOpenSkyClient(config Config) *OpenSkyClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://opensky-network.org/api"
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 15 * time.Second
	}
	return &OpenSkyClient{
		Username: config.Username,
		Password: config.Password,
		BaseURL:  config.BaseURL,
		HTTPClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		lastRequests: make(map[string]time.Time),
	}
}

//=====================================================
// 3) Low-Level HTTP + Rate Limiting
//=====================================================

func (c *OpenSkyClient) doRequest(endpoint string, params url.Values, caller string) ([]byte, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Basic local rate-limiting, adapted from the Python client's approach:
	var minInterval time.Duration
	isAuth := (c.Username != "" && c.Password != "")

	switch caller {
	case "GetStates":
		if isAuth {
			minInterval = 5 * time.Second
		} else {
			minInterval = 10 * time.Second
		}
	case "GetMyStates":
		// must be authenticated
		minInterval = 1 * time.Second
	case "GetFlightsFromInterval", "GetFlightsByAircraft":
		if isAuth {
			minInterval = 5 * time.Second
		} else {
			minInterval = 10 * time.Second
		}
	case "GetArrivalsByAirport", "GetDeparturesByAirport", "GetTrackByAircraft":
		if isAuth {
			minInterval = 5 * time.Second
		} else {
			minInterval = 10 * time.Second
		}
	default:
		minInterval = 5 * time.Second
	}

	// Key by method+endpoint to track last request
	rateLimitKey := fmt.Sprintf("GET %s", endpoint)
	lastTime, ok := c.lastRequests[rateLimitKey]
	if ok {
		elapsed := time.Since(lastTime)
		if elapsed < minInterval {
			time.Sleep(minInterval - elapsed)
		}
	}

	// Build the request
	reqURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}

	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	req.Header.Set("User-Agent", "Go-OpenSky-Client/2.0")
	req.Header.Set("Accept", "application/json")

	// Basic Auth if credentials present
	if isAuth {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	// Update last request time
	c.lastRequests[rateLimitKey] = time.Now()

	return body, resp.StatusCode, nil
}

//=====================================================
// 4) Core OpenSky Methods
//=====================================================

// parseStateVector helper
func parseStateVector(arr []interface{}) StateVector {
	sv := StateVector{}
	if len(arr) > 0 {
		if v, ok := arr[0].(string); ok {
			sv.ICAO24 = v
		}
	}
	if len(arr) > 1 {
		if v, ok := arr[1].(string); ok {
			sv.Callsign = strings.TrimSpace(v)
		}
	}
	if len(arr) > 2 {
		if v, ok := arr[2].(string); ok {
			sv.OriginCountry = v
		}
	}
	if len(arr) > 3 {
		if v, ok := arr[3].(float64); ok {
			tmp := int(v)
			sv.TimePosition = &tmp
		}
	}
	if len(arr) > 4 {
		if v, ok := arr[4].(float64); ok {
			tmp := int(v)
			sv.LastContact = &tmp
		}
	}
	if len(arr) > 5 {
		if v, ok := arr[5].(float64); ok {
			sv.Longitude = &v
		}
	}
	if len(arr) > 6 {
		if v, ok := arr[6].(float64); ok {
			sv.Latitude = &v
		}
	}
	if len(arr) > 7 {
		if v, ok := arr[7].(float64); ok {
			sv.BaroAltitude = &v
		}
	}
	if len(arr) > 8 {
		if v, ok := arr[8].(bool); ok {
			sv.OnGround = v
		}
	}
	if len(arr) > 9 {
		if v, ok := arr[9].(float64); ok {
			sv.Velocity = &v
		}
	}
	if len(arr) > 10 {
		if v, ok := arr[10].(float64); ok {
			sv.TrueTrack = &v
		}
	}
	if len(arr) > 11 {
		if v, ok := arr[11].(float64); ok {
			sv.VerticalRate = &v
		}
	}
	if len(arr) > 12 {
		if vs, ok := arr[12].([]interface{}); ok {
			sensorArr := make([]int, 0, len(vs))
			for _, s := range vs {
				if sensorFloat, ok := s.(float64); ok {
					sensorArr = append(sensorArr, int(sensorFloat))
				}
			}
			sv.Sensors = sensorArr
		}
	}
	if len(arr) > 13 {
		if v, ok := arr[13].(float64); ok {
			sv.GeoAltitude = &v
		}
	}
	if len(arr) > 14 {
		if v, ok := arr[14].(string); ok {
			sv.Squawk = v
		}
	}
	if len(arr) > 15 {
		if v, ok := arr[15].(bool); ok {
			sv.SPI = v
		}
	}
	if len(arr) > 16 {
		if v, ok := arr[16].(float64); ok {
			sv.PositionSource = int(v)
		}
	}
	if len(arr) > 17 {
		if v, ok := arr[17].(float64); ok {
			tmp := int(v)
			sv.Category = &tmp
		}
	}
	return sv
}

// parseFlightData helper
func parseFlightData(entry map[string]interface{}) FlightData {
	fd := FlightData{}
	if v, ok := entry["icao24"].(string); ok {
		fd.ICAO24 = v
	}
	if v, ok := entry["firstSeen"].(float64); ok {
		fd.FirstSeen = int(v)
	}
	if v, ok := entry["estDepartureAirport"].(string); ok {
		fd.EstDepartureAirport = &v
	}
	if v, ok := entry["lastSeen"].(float64); ok {
		fd.LastSeen = int(v)
	}
	if v, ok := entry["estArrivalAirport"].(string); ok {
		fd.EstArrivalAirport = &v
	}
	if v, ok := entry["callsign"].(string); ok {
		trimmed := strings.TrimSpace(v)
		fd.Callsign = &trimmed
	}
	if v, ok := entry["estDepartureAirportHorizDistance"].(float64); ok {
		d := int(v)
		fd.EstDepartureAirportHorizDistance = &d
	}
	if v, ok := entry["estDepartureAirportVertDistance"].(float64); ok {
		d := int(v)
		fd.EstDepartureAirportVertDistance = &d
	}
	if v, ok := entry["estArrivalAirportHorizDistance"].(float64); ok {
		d := int(v)
		fd.EstArrivalAirportHorizDistance = &d
	}
	if v, ok := entry["estArrivalAirportVertDistance"].(float64); ok {
		d := int(v)
		fd.EstArrivalAirportVertDistance = &d
	}
	if v, ok := entry["departureAirportCandidatesCount"].(float64); ok {
		tmp := int(v)
		fd.DepartureAirportCandidatesCount = &tmp
	}
	if v, ok := entry["arrivalAirportCandidatesCount"].(float64); ok {
		tmp := int(v)
		fd.ArrivalAirportCandidatesCount = &tmp
	}
	return fd
}

// GetStates retrieves state vectors for a given time. 0 => most recent.
func (c *OpenSkyClient) GetStates(timeSecs int, icao24 string, bbox []float64) (*OpenSkyStates, error) {
	params := url.Values{}
	if timeSecs != 0 {
		params.Add("time", strconv.Itoa(timeSecs))
	}
	if icao24 != "" {
		params.Add("icao24", icao24)
	}
	params.Add("extended", "true")

	if len(bbox) == 4 {
		params.Add("lamin", fmt.Sprintf("%f", bbox[0]))
		params.Add("lamax", fmt.Sprintf("%f", bbox[1]))
		params.Add("lomin", fmt.Sprintf("%f", bbox[2]))
		params.Add("lomax", fmt.Sprintf("%f", bbox[3]))
	} else if len(bbox) != 0 {
		return nil, errors.New("invalid bounding box, must be exactly 4 floats")
	}

	body, status, err := c.doRequest("/states/all", params, "GetStates")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_states failed: %d => %s", status, string(body))
	}

	var raw struct {
		Time   int             `json:"time"`
		States [][]interface{} `json:"states"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	states := make([]StateVector, 0, len(raw.States))
	for _, arr := range raw.States {
		sv := parseStateVector(arr)
		states = append(states, sv)
	}

	return &OpenSkyStates{Time: raw.Time, States: states}, nil
}

// GetMyStates requires authentication.
func (c *OpenSkyClient) GetMyStates(timeSecs int, icao24 string, serials string) (*OpenSkyStates, error) {
	if c.Username == "" || c.Password == "" {
		return nil, errors.New("getMyStates requires username/password")
	}

	params := url.Values{}
	if timeSecs != 0 {
		params.Add("time", strconv.Itoa(timeSecs))
	}
	if icao24 != "" {
		params.Add("icao24", icao24)
	}
	if serials != "" {
		params.Add("serials", serials)
	}
	params.Add("extended", "true")

	body, status, err := c.doRequest("/states/own", params, "GetMyStates")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_my_states failed: %d => %s", status, string(body))
	}

	var raw struct {
		Time   int             `json:"time"`
		States [][]interface{} `json:"states"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	states := make([]StateVector, 0, len(raw.States))
	for _, arr := range raw.States {
		sv := parseStateVector(arr)
		states = append(states, sv)
	}

	return &OpenSkyStates{Time: raw.Time, States: states}, nil
}

// GetFlightsFromInterval gets flights for [begin, end], up to 2 hours.
func (c *OpenSkyClient) GetFlightsFromInterval(begin, end int) ([]FlightData, error) {
	if begin >= end {
		return nil, errors.New("end must be greater than begin")
	}
	if (end - begin) > 7200 {
		return nil, errors.New("interval must be < 2 hours")
	}

	params := url.Values{}
	params.Add("begin", strconv.Itoa(begin))
	params.Add("end", strconv.Itoa(end))

	body, status, err := c.doRequest("/flights/all", params, "GetFlightsFromInterval")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_flights_from_interval failed: %d => %s", status, string(body))
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	flights := make([]FlightData, 0, len(raw))
	for _, entry := range raw {
		fd := parseFlightData(entry)
		flights = append(flights, fd)
	}

	return flights, nil
}

// GetFlightsByAircraft gets flights for [icao24] in [begin, end] up to 30 days.
func (c *OpenSkyClient) GetFlightsByAircraft(icao24 string, begin, end int) ([]FlightData, error) {
	if begin >= end {
		return nil, errors.New("end must be greater than begin")
	}
	if (end - begin) > 2592000 {
		return nil, errors.New("interval must be < 30 days")
	}

	params := url.Values{}
	params.Add("icao24", icao24)
	params.Add("begin", strconv.Itoa(begin))
	params.Add("end", strconv.Itoa(end))

	body, status, err := c.doRequest("/flights/aircraft", params, "GetFlightsByAircraft")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_flights_by_aircraft failed: %d => %s", status, string(body))
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	flights := make([]FlightData, 0, len(raw))
	for _, entry := range raw {
		fd := parseFlightData(entry)
		flights = append(flights, fd)
	}

	return flights, nil
}

// GetArrivalsByAirport gets arrivals at [airport] in [begin, end] up to 7 days.
func (c *OpenSkyClient) GetArrivalsByAirport(airport string, begin, end int) ([]FlightData, error) {
	if begin >= end {
		return nil, errors.New("end must be greater than begin")
	}
	if (end - begin) > 604800 {
		return nil, errors.New("interval must be < 7 days")
	}

	params := url.Values{}
	params.Add("airport", airport)
	params.Add("begin", strconv.Itoa(begin))
	params.Add("end", strconv.Itoa(end))

	body, status, err := c.doRequest("/flights/arrival", params, "GetArrivalsByAirport")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_arrivals_by_airport failed: %d => %s", status, string(body))
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	flights := make([]FlightData, 0, len(raw))
	for _, entry := range raw {
		fd := parseFlightData(entry)
		flights = append(flights, fd)
	}

	return flights, nil
}

// GetDeparturesByAirport gets departures from [airport] in [begin, end] up to 7 days.
func (c *OpenSkyClient) GetDeparturesByAirport(airport string, begin, end int) ([]FlightData, error) {
	if begin >= end {
		return nil, errors.New("end must be greater than begin")
	}
	if (end - begin) > 604800 {
		return nil, errors.New("interval must be < 7 days")
	}

	params := url.Values{}
	params.Add("airport", airport)
	params.Add("begin", strconv.Itoa(begin))
	params.Add("end", strconv.Itoa(end))

	body, status, err := c.doRequest("/flights/departure", params, "GetDeparturesByAirport")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_departures_by_airport failed: %d => %s", status, string(body))
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	flights := make([]FlightData, 0, len(raw))
	for _, entry := range raw {
		fd := parseFlightData(entry)
		flights = append(flights, fd)
	}

	return flights, nil
}

// GetTrackByAircraft retrieves the flight track for [icao24] at time [t]. 0 => live track.
func (c *OpenSkyClient) GetTrackByAircraft(icao24 string, t int) (*FlightTrack, error) {
	if t != 0 && (int(time.Now().Unix())-t) > 2592000 {
		return nil, errors.New("cannot access flight tracks older than 30 days in the past")
	}

	params := url.Values{}
	params.Add("icao24", icao24)
	params.Add("time", strconv.Itoa(t))

	body, status, err := c.doRequest("/tracks/all", params, "GetTrackByAircraft")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("opensky get_track_by_aircraft failed: %d => %s", status, string(body))
	}

	var raw struct {
		Icao24    string          `json:"icao24"`
		StartTime int             `json:"startTime"`
		EndTime   int             `json:"endTime"`
		Callsign  *string         `json:"callsign"`
		Path      [][]interface{} `json:"path"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	waypoints := make([]Waypoint, 0, len(raw.Path))
	for _, p := range raw.Path {
		wp := Waypoint{}
		// [ time, lat, lon, baro_alt, track, on_ground ]
		if len(p) > 0 {
			if v, ok := p[0].(float64); ok {
				wp.Time = int(v)
			}
		}
		if len(p) > 1 {
			if v, ok := p[1].(float64); ok {
				wp.Latitude = &v
			}
		}
		if len(p) > 2 {
			if v, ok := p[2].(float64); ok {
				wp.Longitude = &v
			}
		}
		if len(p) > 3 {
			if v, ok := p[3].(float64); ok {
				wp.BaroAltitude = &v
			}
		}
		if len(p) > 4 {
			if v, ok := p[4].(float64); ok {
				wp.TrueTrack = &v
			}
		}
		if len(p) > 5 {
			if v, ok := p[5].(bool); ok {
				wp.OnGround = v
			}
		}
		waypoints = append(waypoints, wp)
	}

	ft := &FlightTrack{
		Icao24:    raw.Icao24,
		StartTime: raw.StartTime,
		EndTime:   raw.EndTime,
		Callsign:  raw.Callsign,
		Path:      waypoints,
	}

	return ft, nil
}

//=====================================================
// 5) Gin Handlers (Swagger-friendly)
//=====================================================

// GetStatesAllHandler
// @Summary Get aircraft states (all) [like Python get_states]
// @Description Retrieves the state vectors for aircraft at a given time (or 0 for "now"). Optional: filter by icao24 or bounding box.
// @Tags Flights
// @Param time query int false "Time in seconds since epoch (default=0 => now)"
// @Param icao24 query string false "Single or comma-separated ICAO24 address(es)"
// @Param bbox query string false "min_lat,max_lat,min_lon,max_lon [4 floats]"
// @Produce json
// @Success 200 {object} OpenSkyStates
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /flights/states/all [get]
func GetStatesAllHandler(c *gin.Context) {
	timeParam := c.Query("time")
	icao24Param := c.Query("icao24")
	bboxStr := c.Query("bbox")

	timeSecs := 0
	if timeParam != "" {
		t, err := strconv.Atoi(timeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: "Invalid 'time' param",
			})
			return
		}
		timeSecs = t
	}

	var bbox []float64
	if bboxStr != "" {
		parts := strings.Split(bboxStr, ",")
		if len(parts) != 4 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: "bbox must have exactly 4 floats",
			})
			return
		}
		for _, p := range parts {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, types.ErrorResponse{
					Error: fmt.Sprintf("invalid float in bbox: %v", err),
				})
				return
			}
			bbox = append(bbox, f)
		}
	}

	states, err := openSkyApi.GetStates(timeSecs, icao24Param, bbox)
	if err != nil {
		log.Println("GetStatesAllHandler Error:", err)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, states)
}

// GetMyStatesHandler
// @Summary Get states for your own sensors [like Python get_my_states]
// @Description Requires Basic Auth. Retrieves the state vectors from your own sensors only.
// @Tags Flights
// @Param time query int false "Time in seconds since epoch (0 => now)"
// @Param icao24 query string false "ICAO24 filter"
// @Param serials query string false "Sensor serial(s)"
// @Produce json
// @Success 200 {object} OpenSkyStates
// @Failure 401 {object} ErrorResponse "Unauthorized if no username/password configured"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/my-states [get]
func GetMyStatesHandler(c *gin.Context) {
	timeParam := c.Query("time")
	icao24Param := c.Query("icao24")
	serialsParam := c.Query("serials")

	timeSecs := 0
	if timeParam != "" {
		t, err := strconv.Atoi(timeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid 'time' parameter"})
			return
		}
		timeSecs = t
	}

	result, err := openSkyApi.GetMyStates(timeSecs, icao24Param, serialsParam)
	if err != nil {
		if strings.Contains(err.Error(), "requires username/password") {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetFlightsIntervalHandler
// @Summary Get flights from interval [like Python get_flights_from_interval]
// @Description Retrieves flights for a short interval [begin, end], max 2 hours.
// @Tags Flights
// @Param begin query int true "Start time in seconds"
// @Param end query int true "End time in seconds"
// @Produce json
// @Success 200 {array} FlightData
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/interval [get]
func GetFlightsIntervalHandler(c *gin.Context) {
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	if beginStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "both 'begin' and 'end' are required"})
		return
	}

	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'begin' param"})
		return
	}

	end, err2 := strconv.Atoi(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'end' param"})
		return
	}

	flights, err := openSkyApi.GetFlightsFromInterval(begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, flights)
}

// GetFlightsByAircraftHandlerV2
// @Summary Get flights by aircraft [like Python get_flights_by_aircraft]
// @Description Retrieves flights for [icao24] in [begin, end], up to 30 days.
// @Tags Flights
// @Param icao24 path string true "ICAO24 address (hex)"
// @Param begin query int true "Start time in seconds"
// @Param end query int true "End time in seconds"
// @Produce json
// @Success 200 {array} FlightData
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/aircraft/{icao24} [get]
func GetFlightsByAircraftHandlerV2(c *gin.Context) {
	icao24 := c.Param("icao24")
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	if icao24 == "" || beginStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "icao24, begin, and end are required"})
		return
	}

	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'begin' param"})
		return
	}

	end, err2 := strconv.Atoi(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'end' param"})
		return
	}

	flights, err := openSkyApi.GetFlightsByAircraft(icao24, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, flights)
}

// GetArrivalsByAirportHandlerV2
// @Summary Get arrivals by airport [like Python get_arrivals_by_airport]
// @Description Retrieves flights that arrived at [airport] in [begin, end], up to 7 days.
// @Tags Flights
// @Param airport path string true "ICAO code of airport"
// @Param begin query int true "Start time in seconds"
// @Param end query int true "End time in seconds"
// @Produce json
// @Success 200 {array} FlightData
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/arrivals/{airport} [get]
func GetArrivalsByAirportHandlerV2(c *gin.Context) {
	airport := c.Param("airport")
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	if airport == "" || beginStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "airport, begin, and end are required"})
		return
	}

	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'begin' param"})
		return
	}

	end, err2 := strconv.Atoi(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'end' param"})
		return
	}

	arrivals, err := openSkyApi.GetArrivalsByAirport(airport, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, arrivals)
}

// GetDeparturesByAirportHandlerV2
// @Summary Get departures by airport [like Python get_departures_by_airport]
// @Description Retrieves flights that departed [airport] in [begin, end], up to 7 days.
// @Tags Flights
// @Param airport path string true "ICAO code of airport"
// @Param begin query int true "Start time in seconds"
// @Param end query int true "End time in seconds"
// @Produce json
// @Success 200 {array} FlightData
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/departures/{airport} [get]
func GetDeparturesByAirportHandlerV2(c *gin.Context) {
	airport := c.Param("airport")
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	if airport == "" || beginStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "airport, begin, and end are required"})
		return
	}

	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'begin' param"})
		return
	}

	end, err2 := strconv.Atoi(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'end' param"})
		return
	}

	departures, err := openSkyApi.GetDeparturesByAirport(airport, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, departures)
}

// GetTrackByAircraftHandler
// @Summary Get flight track by aircraft [like Python get_track_by_aircraft]
// @Description Retrieves the trajectory for an aircraft [icao24] at time t. If t=0 => live track.
// @Tags Flights
// @Param icao24 query string true "ICAO24 address"
// @Param time query int false "Unix time (0 => live track)"
// @Produce json
// @Success 200 {object} FlightTrack
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /flights/track [get]
func GetTrackByAircraftHandler(c *gin.Context) {
	icao24 := c.Query("icao24")
	timeStr := c.Query("time")

	if icao24 == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "icao24 is required"})
		return
	}

	t := 0
	if timeStr != "" {
		n, err := strconv.Atoi(timeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid 'time' param"})
			return
		}
		t = n
	}

	track, err := openSkyApi.GetTrackByAircraft(icao24, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, track)
}
