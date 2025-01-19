package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CurrencyInfo defines the structure for each currency in the currencies map.
type CurrencyInfo struct {
	// Name of the currency
	Name string `json:"name" example:"US Dollar"`
	// Symbol of the currency
	Symbol string `json:"symbol" example:"$"`
}

// Currencies is a custom map type that can handle both map and array JSON representations.
type Currencies map[string]CurrencyInfo

// UnmarshalJSON implements custom handling for the "currencies" field.
// It first tries to unmarshal into a map. If that fails, it checks if the JSON is an empty array ([]).
func (c *Currencies) UnmarshalJSON(data []byte) error {
	// If the field is absent, empty, or null, just return with no error.
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// 1) Try unmarshaling as a map (the normal case).
	var m map[string]CurrencyInfo
	if err := json.Unmarshal(data, &m); err == nil {
		*c = m
		return nil
	}

	// 2) If that fails, try unmarshaling as an array. We only accept an empty array here.
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) == 0 {
			// If it's [], treat that as “no currencies.”
			*c = nil
			return nil
		}
		// If the array is not empty, return an error or handle as needed.
		return fmt.Errorf("unexpected non-empty array for currencies: %s", string(data))
	}

	// 3) If neither step worked, return an error about the unexpected format.
	return fmt.Errorf("cannot unmarshal currencies: %s", string(data))
}

// Name represents the name of a country.
type Name struct {
	// Common name of the country
	Common string `json:"common" example:"United States"`
	// Official name of the country
	Official string `json:"official" example:"United States of America"`
}

// IDD represents the International Direct Dialing info for a country.
type IDD struct {
	// Root for the international dialing code.
	Root string `json:"root" example:"+1"`
	// Suffixes for the international dialing code.
	Suffixes []string `json:"suffixes" example:"201,202"`
}

// Maps represents the map URLs for a country.
type Maps struct {
	// Google Maps URL
	GoogleMaps string `json:"googleMaps" example:"https://goo.gl/maps/..."`
	// OpenStreetMaps URL
	OpenStreetMaps string `json:"openStreetMaps" example:"https://www.openstreetmap.org/..."`
}

// Gini represents the Gini coefficient of a country.
type Gini struct {
	// Year the Gini coefficient is from
	Year int `json:"year" example:"2018"`
	// Gini coefficient value
	Coefficient float64 `json:"coefficient" example:"41.4"`
}

// Car represents the car information for a country.
type Car struct {
	// Signs used on license plates
	Signs []string `json:"signs" example:"USA"`
	// Side of the road cars drive on
	Side string `json:"side" example:"right"`
}

// Flags represents the flag URLs for a country.
type Flags struct {
	// SVG format URL of the flag
	Svg string `json:"svg" example:"https://restcountries.eu/data/usa.svg"`
	// PNG format URL of the flag
	Png string `json:"png" example:"https://restcountries.eu/data/usa.png"`
	// Alternative text description of the flag
	Alt string `json:"alt,omitempty" example:"The flag of the United States of America is a..."`
}

// CoatOfArms represents the coat of arms URLs for a country.
type CoatOfArms struct {
	// SVG format URL of the coat of arms
	Svg string `json:"svg" example:"https://mainfacts.com/media/images/coats_of_arms/us.svg"`
	// PNG format URL of the coat of arms
	Png string `json:"png" example:"https://mainfacts.com/media/images/coats_of_arms/us.png"`
}

// CapitalInfo represents the capital information for a country.
type CapitalInfo struct {
	// Latitude and longitude of the capital
	Latlng []float64 `json:"latlng" example:"38.8951,77.0364"`
}

// PostalCode represents the postal code information for a country.
type PostalCode struct {
	// Format of the postal code
	Format string `json:"format" example:"#####-####"`
	// Regex pattern for validating the postal code
	Regex string `json:"regex" example:"^\\d{5}(-\\d{4})?$"`
}

// Demonyms represents the demonyms for a country.
type Demonyms struct {
	// English demonyms
	Eng DemonymInfo `json:"eng"`
	// French demonyms (if available)
	Fra *DemonymInfo `json:"fra,omitempty"`
}

// DemonymInfo represents the demonym information with gender-specific forms.
type DemonymInfo struct {
	// Feminine form of the demonym
	F string `json:"f" example:"American"`
	// Masculine form of the demonym
	M string `json:"m" example:"American"`
}

// Country is the main data structure to match your RestCountries-like JSON.
type Country struct {
	Name         Name               `json:"name"`
	TLD          []string           `json:"tld,omitempty"`
	CCA2         string             `json:"cca2" example:"US"`
	CCN3         string             `json:"ccn3,omitempty" example:"840"`
	CCA3         string             `json:"cca3" example:"USA"`
	CIOC         string             `json:"cioc,omitempty" example:"USA"`
	FIFA         string             `json:"fifa,omitempty" example:"USA"`
	Independent  bool               `json:"independent" example:"true"`
	Status       string             `json:"status,omitempty" example:"officially-assigned"`
	UNMember     bool               `json:"unMember" example:"true"`
	Currencies   Currencies         `json:"currencies,omitempty"`
	IDD          IDD                `json:"idd"`
	Capital      []string           `json:"capital,omitempty" example:"Washington, D.C."`
	AltSpellings []string           `json:"altSpellings,omitempty" example:"US,USA,United States of America"`
	Latlng       []float64          `json:"latlng,omitempty" example:"38,97"`
	Landlocked   bool               `json:"landlocked" example:"false"`
	Borders      []string           `json:"borders,omitempty" example:"CAN,MEX"`
	Area         float64            `json:"area" example:"9372610"`
	Flag         string             `json:"flag,omitempty" example:"🇺🇸"`
	Region       string             `json:"region" example:"Americas"`
	Subregion    string             `json:"subregion,omitempty" example:"North America"`
	Maps         Maps               `json:"maps"`
	Population   int                `json:"population" example:"334805269"`
	Gini         map[string]float64 `json:"gini,omitempty" example:"2019:39.7"`
	Car          Car                `json:"car"`
	Timezones    []string           `json:"timezones" example:"UTC-12:00,UTC-11:00,UTC-10:00,UTC-09:00,UTC-08:00,UTC-07:00,UTC-06:00,UTC-05:00,UTC-04:00,UTC+10:00,UTC+12:00"`
	Continents   []string           `json:"continents" example:"North America"`
	Flags        Flags              `json:"flags"`
	CoatOfArms   CoatOfArms         `json:"coatOfArms"`
	StartOfWeek  string             `json:"startOfWeek" example:"sunday"`
	CapitalInfo  CapitalInfo        `json:"capitalInfo"`
	PostalCode   PostalCode         `json:"postalCode,omitempty"`
	Demonyms     Demonyms           `json:"demonyms"`

	Languages    map[string]string `json:"languages,omitempty" example:"eng:English"`
	Translations map[string]struct {
		Official string `json:"official" example:"Vereinigte Staaten von Amerika"`
		Common   string `json:"common" example:"Vereinigte Staaten"`
	} `json:"translations,omitempty"`
}

// ErrorResponse represents an error response.
// @Description Error response model
type ErrorResponse struct {
	// The error message
	// Required: true
	Message string `json:"message" example:"Bad request"`
}

// Countries holds the data once loaded.
var Countries []Country

// LoadCountriesSafe reads local JSON data into the global Countries variable.
func LoadCountriesSafe(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read countries file: %w", err)
	}

	if err := json.Unmarshal(data, &Countries); err != nil {
		return fmt.Errorf("failed to parse countries data: %w", err)
	}

	return nil
}

// filterCountries applies field-based filtering logic.
func filterCountries(filters map[string]string) []Country {
	filteredCountries := []Country{}

	for _, country := range Countries {
		match := true

		for key, value := range filters {
			switch key {
			case "independent":
				// e.g. ?independent=true or ?independent=false
				wantBool := (value == "true")
				if country.Independent != wantBool {
					match = false
				}

			case "name":
				// partial match on Name.Common or Name.Official
				lowVal := strings.ToLower(value)
				if !strings.Contains(strings.ToLower(country.Name.Common), lowVal) &&
					!strings.Contains(strings.ToLower(country.Name.Official), lowVal) {
					match = false
				}

			case "fullName":
				// exact match on Name.Common or Name.Official
				// value is the original name param
				if !strings.EqualFold(country.Name.Common, value) &&
					!strings.EqualFold(country.Name.Official, value) {
					match = false
				}

			case "currency":
				// e.g. currency=USD or currency="United States dollar"
				found := false
				for code, cinfo := range country.Currencies {
					if strings.EqualFold(code, value) ||
						strings.EqualFold(cinfo.Name, value) {
						found = true
						break
					}
				}
				if !found {
					match = false
				}

			case "demonym":
				// e.g. /v1/demonym/Aruban => check country.Demonyms.Eng or .Fra
				d := country.Demonyms
				if !strings.EqualFold(d.Eng.M, value) &&
					!strings.EqualFold(d.Eng.F, value) &&
					d.Fra != nil && !strings.EqualFold(d.Fra.M, value) &&
					d.Fra != nil && !strings.EqualFold(d.Fra.F, value) {
					match = false
				}

			case "language":
				// e.g. /v1/lang/spanish or /v1/lang/spa
				found := false
				for code, lang := range country.Languages {
					if strings.EqualFold(code, value) ||
						strings.EqualFold(lang, value) {
						found = true
						break
					}
				}
				if !found {
					match = false
				}

			case "capital":
				// e.g. /v1/capital/Tallinn
				found := false
				for _, capVal := range country.Capital {
					if strings.EqualFold(capVal, value) {
						found = true
						break
					}
				}
				if !found {
					match = false
				}

			case "region":
				// e.g. /v1/region/Europe
				if !strings.EqualFold(country.Region, value) {
					match = false
				}

			case "subregion":
				// e.g. /v1/subregion/Northern Europe
				if !strings.EqualFold(country.Subregion, value) {
					match = false
				}

			case "translation":
				// e.g. /v1/translation/Saksamaa
				found := false
				lowVal := strings.ToLower(value)
				for _, tr := range country.Translations {
					if strings.Contains(strings.ToLower(tr.Common), lowVal) ||
						strings.Contains(strings.ToLower(tr.Official), lowVal) {
						found = true
						break
					}
				}
				if !found {
					match = false
				}
			}

			if !match {
				break
			}
		}

		if match {
			filteredCountries = append(filteredCountries, country)
		}
	}

	return filteredCountries
}

// selectFields returns only the requested fields for a single country.
func selectFields(country Country, fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range fields {
		switch field {
		case "name":
			result["name"] = country.Name
		case "capital":
			result["capital"] = country.Capital
		case "currencies":
			result["currencies"] = country.Currencies
		case "languages":
			result["languages"] = country.Languages
		case "demonyms":
			result["demonyms"] = country.Demonyms
		case "region":
			result["region"] = country.Region
		case "subregion":
			result["subregion"] = country.Subregion
		case "translations":
			result["translations"] = country.Translations
		case "cca2":
			result["cca2"] = country.CCA2
		case "cca3":
			result["cca3"] = country.CCA3
		case "ccn3":
			result["ccn3"] = country.CCN3
		case "cioc":
			result["cioc"] = country.CIOC
		case "fifa":
			result["fifa"] = country.FIFA
		case "independent":
			result["independent"] = country.Independent
		case "status":
			result["status"] = country.Status
		case "unMember":
			result["unMember"] = country.UNMember
		case "idd":
			result["idd"] = country.IDD
		case "altSpellings":
			result["altSpellings"] = country.AltSpellings
		case "latlng":
			result["latlng"] = country.Latlng
		case "landlocked":
			result["landlocked"] = country.Landlocked
		case "borders":
			result["borders"] = country.Borders
		case "area":
			result["area"] = country.Area
		case "flag":
			result["flag"] = country.Flag
		case "maps":
			result["maps"] = country.Maps
		case "population":
			result["population"] = country.Population
		case "gini":
			result["gini"] = country.Gini
		case "car":
			result["car"] = country.Car
		case "timezones":
			result["timezones"] = country.Timezones
		case "continents":
			result["continents"] = country.Continents
		case "flags":
			result["flags"] = country.Flags
		case "coatOfArms":
			result["coatOfArms"] = country.CoatOfArms
		case "startOfWeek":
			result["startOfWeek"] = country.StartOfWeek
		case "capitalInfo":
			result["capitalInfo"] = country.CapitalInfo
		case "postalCode":
			result["postalCode"] = country.PostalCode
		}
	}

	return result
}

// validateBooleanQuery checks if the query parameter is "true", "false", or empty.
// Returns the lowercase string if valid, or an error otherwise.
func validateBooleanQuery(paramValue string) (string, error) {
	if paramValue == "" {
		return "", nil // no value was given
	}
	lowVal := strings.ToLower(paramValue)
	if lowVal == "true" || lowVal == "false" {
		return lowVal, nil
	}
	return "", fmt.Errorf("invalid boolean value: %s (must be 'true' or 'false')", paramValue)
}

// --------------------------------------------------------------------------
// Below are your actual HTTP handlers with Swag doc comments, referencing Country.
// --------------------------------------------------------------------------

// GetCountries godoc
// @Summary     Get all countries
// @Description Get details of all countries, with optional filters.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       independent query string false "Filter by independent status (true or false)"
// @Param       fields      query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     400 {object} ErrorResponse "Bad request"
// @Router      /countries [get]
func GetCountries(c *gin.Context) {
	filters := make(map[string]string)

	indVal, err := validateBooleanQuery(c.Query("independent"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	if indVal != "" {
		// It's either "true" or "false" from validateBooleanQuery
		filters["independent"] = indVal
	}

	filteredCountries := filterCountries(filters)
	fields := c.Query("fields")

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountryByCode godoc
// @Summary     Get country by code
// @Description Get details of a specific country by its code (CCA2 or CCA3).
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       code   path  string true  "Country code (CCA2 or CCA3)"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {object} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /countries/{code} [get]
func GetCountryByCode(c *gin.Context) {
	code := c.Param("code")
	fields := c.Query("fields")

	for _, country := range Countries {
		if strings.EqualFold(country.CCA2, code) || strings.EqualFold(country.CCA3, code) {
			if fields != "" {
				fieldList := strings.Split(fields, ",")
				c.JSON(http.StatusOK, selectFields(country, fieldList))
			} else {
				c.JSON(http.StatusOK, country)
			}
			return
		}
	}

	c.JSON(http.StatusNotFound, ErrorResponse{Message: "Country not found"})
}

// GetCountriesByName godoc
// @Summary     Get countries by name
// @Description Get countries matching a name query (common or official). Use fullText=true for exact name match.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       name     path string true  "Country name (common or official)"
// @Param       fullText query string false "Exact match for full name (true/false)"
// @Param       fields   query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     400 {object} ErrorResponse "Bad request"
// @Router      /name/{name} [get]
func GetCountriesByName(c *gin.Context) {
	name := c.Param("name")
	fullTextParam := c.Query("fullText")
	fields := c.Query("fields")

	// Validate the fullTextParam
	boolVal, err := validateBooleanQuery(fullTextParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	filters := map[string]string{}
	// If user sets fullText=true, we do exact match. Otherwise partial match
	if boolVal == "true" {
		filters["fullName"] = name
	} else {
		// if boolVal == "false" or not specified => do partial match
		filters["name"] = name
	}

	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByCodes godoc
// @Summary     Get countries by codes
// @Description Get countries matching a list of codes (CCA2, CCN3, CCA3, or CIOC).
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       codes  query string true  "Comma-separated list of country codes (CCA2, CCN3, CCA3, CIOC)"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     400 {object} ErrorResponse "Bad request"
// @Router      /alpha [get]
func GetCountriesByCodes(c *gin.Context) {
	codes := c.Query("codes")
	fields := c.Query("fields")

	if codes == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Query parameter 'codes' is required"})
		return
	}

	codeList := strings.Split(codes, ",")

	var filteredCountries []Country
	for _, country := range Countries {
		for _, code := range codeList {
			if strings.EqualFold(country.CCA2, code) ||
				strings.EqualFold(country.CCA3, code) ||
				strings.EqualFold(country.CCN3, code) ||
				strings.EqualFold(country.CIOC, code) {
				filteredCountries = append(filteredCountries, country)
				break
			}
		}
	}

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByCurrency godoc
// @Summary     Get countries by currency
// @Description Get countries matching a currency code or name.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       currency path string true  "Currency code or name"
// @Param       fields   query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /currency/{currency} [get]
func GetCountriesByCurrency(c *gin.Context) {
	currency := c.Param("currency")
	fields := c.Query("fields")

	filters := map[string]string{"currency": currency}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByDemonym godoc
// @Summary     Get countries by demonym
// @Description Get countries matching a demonym.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       demonym path string true  "Demonym"
// @Param       fields  query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /demonym/{demonym} [get]
func GetCountriesByDemonym(c *gin.Context) {
	demonym := c.Param("demonym")
	fields := c.Query("fields")

	filters := map[string]string{"demonym": demonym}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByLanguage godoc
// @Summary     Get countries by language
// @Description Get countries matching a language code or name.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       language path string true  "Language code or name"
// @Param       fields   query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /lang/{language} [get]
func GetCountriesByLanguage(c *gin.Context) {
	language := c.Param("language")
	fields := c.Query("fields")

	filters := map[string]string{"language": language}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByCapital godoc
// @Summary     Get countries by capital
// @Description Get countries matching a capital city name.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       capital path string true  "Capital city name"
// @Param       fields  query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /capital/{capital} [get]
func GetCountriesByCapital(c *gin.Context) {
	capital := c.Param("capital")
	fields := c.Query("fields")

	filters := map[string]string{"capital": capital}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByRegion godoc
// @Summary     Get countries by region
// @Description Get countries matching a region.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       region path string true  "Region name"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /region/{region} [get]
func GetCountriesByRegion(c *gin.Context) {
	region := c.Param("region")
	fields := c.Query("fields")

	filters := map[string]string{"region": region}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesBySubregion godoc
// @Summary     Get countries by subregion
// @Description Get countries matching a subregion.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       subregion path string true  "Subregion name"
// @Param       fields    query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /subregion/{subregion} [get]
func GetCountriesBySubregion(c *gin.Context) {
	subregion := c.Param("subregion")
	fields := c.Query("fields")

	filters := map[string]string{"subregion": subregion}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByTranslation godoc
// @Summary     Get countries by translation
// @Description Get countries matching a translation.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       translation path string true  "Translation"
// @Param       fields      query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     404 {object} ErrorResponse "Country not found"
// @Router      /translation/{translation} [get]
func GetCountriesByTranslation(c *gin.Context) {
	translation := c.Param("translation")
	fields := c.Query("fields")

	filters := map[string]string{"translation": translation}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountriesByIndependence godoc
// @Summary     Get countries by independence status
// @Description Get countries filtered by independence. If not specified, defaults to status=true.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       status query string false "true or false. Defaults to 'true' if not specified"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array} Country
// @Failure     400 {object} ErrorResponse "Bad request"
// @Router      /independent [get]
func GetCountriesByIndependence(c *gin.Context) {
	status := c.Query("status")

	// Default to "true" if not specified
	if status == "" {
		status = "true"
	}

	// Validate
	statusBool, err := validateBooleanQuery(status)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	filters := map[string]string{
		"independent": statusBool, // "true" or "false"
	}

	filteredCountries := filterCountries(filters)

	fields := c.Query("fields")
	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, country := range filteredCountries {
			result = append(result, selectFields(country, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}
