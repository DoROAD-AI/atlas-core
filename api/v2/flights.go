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

	"regexp"

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

// FlightData represents data about a single flight
// as returned by the OpenSky API.
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

// ------------------------------------------------------------------------------
// FlightDataResponse is an "enhanced" version of FlightData for output.
// It preserves the original Unix fields but also includes human-readable times.
// ------------------------------------------------------------------------------
type FlightDataResponse struct {
	ICAO24 string `json:"icao24"`

	// Original Unix timestamps (for backward compatibility).
	FirstSeenUnix int `json:"firstSeenUnix"`
	LastSeenUnix  int `json:"lastSeenUnix"`

	// Human-readable RFC3339 timestamps.
	FirstSeenUtc string `json:"firstSeenUtc,omitempty"`
	LastSeenUtc  string `json:"lastSeenUtc,omitempty"`

	EstDepartureAirport              *string `json:"estDepartureAirport,omitempty"`
	EstArrivalAirport                *string `json:"estArrivalAirport,omitempty"`
	Callsign                         *string `json:"callsign,omitempty"`
	EstDepartureAirportHorizDistance *int    `json:"estDepartureAirportHorizDistance,omitempty"`
	EstDepartureAirportVertDistance  *int    `json:"estDepartureAirportVertDistance,omitempty"`
	EstArrivalAirportHorizDistance   *int    `json:"estArrivalAirportHorizDistance,omitempty"`
	EstArrivalAirportVertDistance    *int    `json:"estArrivalAirportVertDistance,omitempty"`
	DepartureAirportCandidatesCount  *int    `json:"departureAirportCandidatesCount,omitempty"`
	ArrivalAirportCandidatesCount    *int    `json:"arrivalAirportCandidatesCount,omitempty"`
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
	HTTPClient *http.Client
	Username   string
	Password   string
	BaseURL    string
	mu         sync.Mutex
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
	}
}

//=====================================================
// 3) Low-Level HTTP (Removed local rate-limiting)
//=====================================================

// doRequest performs an HTTP GET with optional params.
// Basic Auth is applied if configured. We have removed
// the local rate-limit logic so that we rely on the
// server-side rate limitations.
func (c *OpenSkyClient) doRequest(endpoint string, params url.Values) ([]byte, int, error) {
	// c.mu.Lock() and c.mu.Unlock() can still be used if concurrency is a concern
	// but local rate-limiting logic has been removed.

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

	// If credentials are provided, use Basic Auth.
	if c.Username != "" && c.Password != "" {
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

	return body, resp.StatusCode, nil
}

//=====================================================
// 4) Core OpenSky Methods
//=====================================================

// parseStateVector converts an array of interface{} (from JSON) into a typed StateVector.
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
				if sensorFloat, ok2 := s.(float64); ok2 {
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

// parseFlightData converts a map (from JSON) into a typed FlightData.
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

	body, status, err := c.doRequest("/states/all", params)
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

	body, status, err := c.doRequest("/states/own", params)
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

	body, status, err := c.doRequest("/flights/all", params)
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

	body, status, err := c.doRequest("/flights/aircraft", params)
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

	body, status, err := c.doRequest("/flights/arrival", params)
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

	body, status, err := c.doRequest("/flights/departure", params)
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
	// The official OpenSky docs say you cannot go older than 30 days,
	// but the user can still request t=0 => "live" track.
	if t != 0 && (int(time.Now().Unix())-t) > 2592000 {
		return nil, errors.New("cannot access flight tracks older than 30 days in the past")
	}

	params := url.Values{}
	params.Add("icao24", icao24)
	params.Add("time", strconv.Itoa(t))

	body, status, err := c.doRequest("/tracks/all", params)
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
// 5) Helper Functions for Flexible Time Parsing & Output
//=====================================================

// parseFlexibleTime parses a string time input into a Unix timestamp (int seconds).
// It supports:
//  1. Positive Unix timestamps (e.g. 1735022400).
//  2. Negative integer => relative offset from "now" (e.g. -3600 => now - 1 hour).
//  3. ISO8601/RFC3339 absolute times (e.g. 2025-01-21T12:00:00Z).
//  4. Relative patterns with suffix like -24h, -7d, -30m, etc. (optional).
//
// Returns an error if it cannot parse.
func parseFlexibleTime(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		// If user didn't specify, treat as 0 => "now" for many endpoints, or no filter.
		return 0, nil
	}

	// 1) Try integer parse directly (covers positive or negative).
	if val, err := strconv.Atoi(raw); err == nil {
		// If negative => offset from now.
		if val < 0 {
			now := time.Now().Unix()
			return int(now) + val, nil
		}
		// If positive => treat as a raw Unix timestamp
		return val, nil
	}

	// 2) Try scanning for suffix-based patterns (e.g. -24h, -7d, etc.).
	// Simple approach: use a regex or manual check.
	// Example pattern: ^(-?\d+)([smhd])$
	//   If the user typed "-24h", that means now - 24 hours.
	rx := regexp.MustCompile(`^(-?\d+)([smhd])$`)
	if match := rx.FindStringSubmatch(raw); match != nil {
		numStr := match[1] // e.g. "-24"
		unit := match[2]   // e.g. "h"
		amount, _ := strconv.Atoi(numStr)

		mult := 1
		switch unit {
		case "s":
			mult = 1
		case "m":
			mult = 60
		case "h":
			mult = 3600
		case "d":
			mult = 86400
		}
		offset := amount * mult
		now := time.Now().Unix()
		return int(now) + offset, nil
	}

	// 3) Attempt RFC3339 parse (ISO8601).
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return int(t.Unix()), nil
	}

	return 0, fmt.Errorf("could not parse time: %s", raw)
}

// transformFlightData converts FlightData into FlightDataResponse,
// preserving Unix times but also adding ISO8601 versions.
func transformFlightData(f FlightData) FlightDataResponse {
	fdResp := FlightDataResponse{
		ICAO24:                           f.ICAO24,
		FirstSeenUnix:                    f.FirstSeen,
		LastSeenUnix:                     f.LastSeen,
		EstDepartureAirport:              f.EstDepartureAirport,
		EstArrivalAirport:                f.EstArrivalAirport,
		Callsign:                         f.Callsign,
		EstDepartureAirportHorizDistance: f.EstDepartureAirportHorizDistance,
		EstDepartureAirportVertDistance:  f.EstDepartureAirportVertDistance,
		EstArrivalAirportHorizDistance:   f.EstArrivalAirportHorizDistance,
		EstArrivalAirportVertDistance:    f.EstArrivalAirportVertDistance,
		DepartureAirportCandidatesCount:  f.DepartureAirportCandidatesCount,
		ArrivalAirportCandidatesCount:    f.ArrivalAirportCandidatesCount,
	}

	// Convert Unix => RFC3339 for the "Utc" fields, if not zero.
	if f.FirstSeen > 0 {
		fdResp.FirstSeenUtc = time.Unix(int64(f.FirstSeen), 0).UTC().Format(time.RFC3339)
	}
	if f.LastSeen > 0 {
		fdResp.LastSeenUtc = time.Unix(int64(f.LastSeen), 0).UTC().Format(time.RFC3339)
	}

	return fdResp
}

// enhanceFlightsResponse wraps the array of flight data with additional
// "beginTimeUnix", "beginTimeUtc", "endTimeUnix", and "endTimeUtc" fields.
func enhanceFlightsResponse(
	c *gin.Context,
	flights []FlightData,
	begin, end int,
) {
	// Convert flight slice to response slice
	results := make([]FlightDataResponse, 0, len(flights))
	for _, f := range flights {
		results = append(results, transformFlightData(f))
	}

	beginTimeUtc := ""
	endTimeUtc := ""
	if begin != 0 {
		beginTimeUtc = time.Unix(int64(begin), 0).UTC().Format(time.RFC3339)
	}
	if end != 0 {
		endTimeUtc = time.Unix(int64(end), 0).UTC().Format(time.RFC3339)
	}

	// Return a JSON response wrapping the flights plus the time info
	c.JSON(http.StatusOK, gin.H{
		"beginTimeUnix": begin,
		"beginTimeUtc":  beginTimeUtc,
		"endTimeUnix":   end,
		"endTimeUtc":    endTimeUtc,
		"flights":       results,
	})
}

//=====================================================
// 6) Gin Handlers (Swagger-friendly)
//=====================================================

// GetStatesAllHandler
// @Summary Get aircraft states (all)
// @Description Retrieves the state vectors for aircraft at a given time (or 0 for "now"). Optional: filter by icao24 or bounding box.
// @Tags Flights
// @Param time query string false "Time can be Unix, RFC3339, or negative/relative (default=0 => now)"
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

	parsedTime, err := parseFlexibleTime(timeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: fmt.Sprintf("Invalid 'time' param: %v", err),
		})
		return
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

	states, err := openSkyApi.GetStates(parsedTime, icao24Param, bbox)
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
// @Summary Get states for your own sensors
// @Description Requires Basic Auth. Retrieves the state vectors from your own sensors only.
// @Tags Flights
// @Param time query string false "Time can be Unix, RFC3339, or negative/relative (default=0 => now)"
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

	parsedTime, err := parseFlexibleTime(timeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: fmt.Sprintf("Invalid 'time' parameter: %v", err),
		})
		return
	}

	result, err := openSkyApi.GetMyStates(parsedTime, icao24Param, serialsParam)
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
// @Summary Get flights from interval
// @Description Retrieves flights for a short interval [begin, end], max 2 hours.
// @Tags Flights
// @Param begin query string true "Start time (Unix, RFC3339, or relative)"
// @Param end query string true "End time (Unix, RFC3339, or relative)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Enhanced flight data + boundary times"
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

	begin, err := parseFlexibleTime(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'begin' param: %v", err)})
		return
	}

	end, err2 := parseFlexibleTime(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'end' param: %v", err2)})
		return
	}

	flights, err := openSkyApi.GetFlightsFromInterval(begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	// Return an enhanced response with times
	enhanceFlightsResponse(c, flights, begin, end)
}

// GetFlightsByAircraftHandlerV2
// @Summary Get flights by aircraft
// @Description Retrieves flights for [icao24] in [begin, end], up to 30 days.
// @Tags Flights
// @Param icao24 path string true "ICAO24 address (hex)"
// @Param begin query string true "Start time (Unix, RFC3339, or relative)"
// @Param end query string true "End time (Unix, RFC3339, or relative)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Enhanced flight data + boundary times"
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

	begin, err := parseFlexibleTime(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'begin' param: %v", err)})
		return
	}
	end, err2 := parseFlexibleTime(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'end' param: %v", err2)})
		return
	}

	flights, err := openSkyApi.GetFlightsByAircraft(icao24, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	enhanceFlightsResponse(c, flights, begin, end)
}

// GetArrivalsByAirportHandlerV2
// @Summary Get arrivals by airport
// @Description Retrieves flights that arrived at [airport] in [begin, end], up to 7 days.
// @Tags Flights
// @Param airport path string true "ICAO code of airport"
// @Param begin query string true "Start time (Unix, RFC3339, or relative)"
// @Param end query string true "End time (Unix, RFC3339, or relative)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Enhanced flight data + boundary times"
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

	begin, err := parseFlexibleTime(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'begin' param: %v", err)})
		return
	}
	end, err2 := parseFlexibleTime(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'end' param: %v", err2)})
		return
	}

	arrivals, err := openSkyApi.GetArrivalsByAirport(airport, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	enhanceFlightsResponse(c, arrivals, begin, end)
}

// GetDeparturesByAirportHandlerV2
// @Summary Get departures by airport
// @Description Retrieves flights that departed [airport] in [begin, end], up to 7 days.
// @Tags Flights
// @Param airport path string true "ICAO code of airport"
// @Param begin query string true "Start time (Unix, RFC3339, or relative)"
// @Param end query string true "End time (Unix, RFC3339, or relative)"
// @Produce json
// @Success 200 {object} map[string]interface{} "Enhanced flight data + boundary times"
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

	begin, err := parseFlexibleTime(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'begin' param: %v", err)})
		return
	}
	end, err2 := parseFlexibleTime(endStr)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'end' param: %v", err2)})
		return
	}

	departures, err := openSkyApi.GetDeparturesByAirport(airport, begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	enhanceFlightsResponse(c, departures, begin, end)
}

// GetTrackByAircraftHandler
// @Summary Get flight track by aircraft
// @Description Retrieves the trajectory for an aircraft [icao24] at time t. If t=0 => live track.
// @Tags Flights
// @Param icao24 query string true "ICAO24 address"
// @Param time query string false "Time can be Unix, RFC3339, or negative/relative (0 => live track)"
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

	t, err := parseFlexibleTime(timeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: fmt.Sprintf("invalid 'time' param: %v", err)})
		return
	}

	track, err := openSkyApi.GetTrackByAircraft(icao24, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, track)
}
