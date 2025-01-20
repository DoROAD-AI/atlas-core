// handlers.go contains the HTTP handlers for the API endpoints. The handlers are implemented using the Gin framework, which provides a high-performance HTTP router and middleware framework for Go.
package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// CurrencyInfo defines the structure for each currency in the currencies map.
type CurrencyInfo struct {
	Name   string `json:"name" example:"US Dollar"`
	Symbol string `json:"symbol" example:"$"`
}

// Currencies is a custom map type that can handle both map and array JSON representations.
type Currencies map[string]CurrencyInfo

// UnmarshalJSON implements custom handling for the "currencies" field.
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

	// 3) If neither step worked, return an error.
	return fmt.Errorf("cannot unmarshal currencies: %s", string(data))
}

// Name represents the name of a country.
type Name struct {
	Common   string `json:"common" example:"United States"`
	Official string `json:"official" example:"United States of America"`
}

// IDD represents the International Direct Dialing info for a country.
type IDD struct {
	Root     string   `json:"root" example:"+1"`
	Suffixes []string `json:"suffixes" example:"201,202"`
}

// Maps represents the map URLs for a country.
type Maps struct {
	GoogleMaps     string `json:"googleMaps" example:"https://goo.gl/maps/..."`
	OpenStreetMaps string `json:"openStreetMaps" example:"https://www.openstreetmap.org/..."`
}

// Car represents the car information for a country.
type Car struct {
	Signs []string `json:"signs" example:"USA"`
	Side  string   `json:"side" example:"right"`
}

// Flags represents the flag URLs for a country.
type Flags struct {
	Svg string `json:"svg" example:"https://restcountries.eu/data/usa.svg"`
	Png string `json:"png" example:"https://restcountries.eu/data/usa.png"`
	Alt string `json:"alt,omitempty" example:"Flag of the United States"`
}

// CoatOfArms represents the coat of arms URLs for a country.
type CoatOfArms struct {
	Svg string `json:"svg" example:"https://mainfacts.com/media/images/coats_of_arms/us.svg"`
	Png string `json:"png" example:"https://mainfacts.com/media/images/coats_of_arms/us.png"`
}

// CapitalInfo represents the capital information for a country.
type CapitalInfo struct {
	Latlng []float64 `json:"latlng" example:"38.8951,77.0364"`
}

// PostalCode represents the postal code information for a country.
type PostalCode struct {
	Format string `json:"format" example:"#####-####"`
	Regex  string `json:"regex" example:"^\\d{5}(-\\d{4})?$"`
}

// Demonyms represents the demonyms for a country.
type Demonyms struct {
	Eng DemonymInfo  `json:"eng"`
	Fra *DemonymInfo `json:"fra,omitempty"`
}

// DemonymInfo represents the demonym information with gender-specific forms.
type DemonymInfo struct {
	F string `json:"f" example:"American"`
	M string `json:"m" example:"American"`
}

// Country is the main data structure.
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
	AltSpellings []string           `json:"altSpellings,omitempty"`
	Latlng       []float64          `json:"latlng,omitempty"`
	Landlocked   bool               `json:"landlocked" example:"false"`
	Borders      []string           `json:"borders,omitempty"`
	Area         float64            `json:"area" example:"9372610"`
	Flag         string             `json:"flag,omitempty" example:"🇺🇸"`
	Region       string             `json:"region" example:"Americas"`
	Subregion    string             `json:"subregion,omitempty" example:"North America"`
	Maps         Maps               `json:"maps"`
	Population   int                `json:"population" example:"334805269"`
	Gini         map[string]float64 `json:"gini,omitempty"`
	Car          Car                `json:"car"`
	Timezones    []string           `json:"timezones"`
	Continents   []string           `json:"continents"`
	Flags        Flags              `json:"flags"`
	CoatOfArms   CoatOfArms         `json:"coatOfArms"`
	StartOfWeek  string             `json:"startOfWeek" example:"sunday"`
	CapitalInfo  CapitalInfo        `json:"capitalInfo"`
	PostalCode   PostalCode         `json:"postalCode,omitempty"`
	Demonyms     Demonyms           `json:"demonyms"`

	Languages    map[string]string `json:"languages,omitempty"`
	Translations map[string]struct {
		Official string `json:"official"`
		Common   string `json:"common"`
	} `json:"translations,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
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

// filterCountries applies field-based filtering logic (unchanged).
func filterCountries(filters map[string]string) []Country {
	filteredCountries := []Country{}

	for _, country := range Countries {
		match := true
		for key, value := range filters {
			switch key {
			case "independent":
				// ?independent=true or ?independent=false
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
				if !strings.EqualFold(country.Name.Common, value) &&
					!strings.EqualFold(country.Name.Official, value) {
					match = false
				}

			case "currency":
				// currency=USD or currency="United States dollar"
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
				d := country.Demonyms
				// Check both English and French demonyms if available
				if !strings.EqualFold(d.Eng.M, value) &&
					!strings.EqualFold(d.Eng.F, value) &&
					(d.Fra == nil ||
						(!strings.EqualFold(d.Fra.M, value) &&
							!strings.EqualFold(d.Fra.F, value))) {
					match = false
				}

			case "language":
				// language=Spanish or language=spa
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
				// capital=Tallinn
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
				if !strings.EqualFold(country.Region, value) {
					match = false
				}

			case "subregion":
				if !strings.EqualFold(country.Subregion, value) {
					match = false
				}

			case "translation":
				// translation=Saksamaa
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

// selectFields uses reflection to retrieve nested fields (e.g., "flags.svg").
func selectFields(country Country, fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range fields {
		fieldParts := strings.Split(field, ".")
		var value interface{} = country

		// Traverse nested fields
		for _, part := range fieldParts {
			value = getFieldValue(value, part)
			if value == nil {
				break
			}
		}

		// If we successfully found a value, store it under the top-level field name
		if value != nil {
			result[fieldParts[0]] = value
		}
	}

	return result
}

// getFieldValue dynamically gets the field (case-insensitive) from struct or map.
func getFieldValue(obj interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Case-insensitive match on struct field name
		f := v.FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(n, fieldName)
		})
		if f.IsValid() && f.CanInterface() {
			return f.Interface()
		}
	case reflect.Map:
		// Case-insensitive match on map key
		for _, key := range v.MapKeys() {
			if key.Kind() == reflect.String && strings.EqualFold(key.String(), fieldName) {
				value := v.MapIndex(key)
				if value.IsValid() && value.CanInterface() {
					return value.Interface()
				}
			}
		}
	}

	return nil
}

// validateBooleanQuery checks if the query parameter is "true", "false", or empty.
func validateBooleanQuery(paramValue string) (string, error) {
	if paramValue == "" {
		return "", nil
	}
	lowVal := strings.ToLower(paramValue)
	if lowVal == "true" || lowVal == "false" {
		return lowVal, nil
	}
	return "", fmt.Errorf("invalid boolean value: %s (must be 'true' or 'false')", paramValue)
}

// --------------------------------------------------------------------------
// HTTP Handlers
// --------------------------------------------------------------------------

// GetCountries godoc
// @Summary     Get all countries
// @Description Get details of all countries, with optional filters.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       independent query string false "Filter by independent status (true or false)"
// @Param       fields      query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array}  Country
// @Failure     400 {object} ErrorResponse
// @Router      /countries [get]
func GetCountries(c *gin.Context) {
	filters := make(map[string]string)

	indVal, err := validateBooleanQuery(c.Query("independent"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	if indVal != "" {
		// It's either "true" or "false"
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
// @Failure     404 {object} ErrorResponse
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
// @Success     200 {array}  Country
// @Failure     400 {object} ErrorResponse
// @Router      /name/{name} [get]
func GetCountriesByName(c *gin.Context) {
	name := c.Param("name")
	fullTextParam := c.Query("fullText")
	fields := c.Query("fields")

	boolVal, err := validateBooleanQuery(fullTextParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	filters := map[string]string{}
	if boolVal == "true" {
		filters["fullName"] = name
	} else {
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
// @Success     200 {array}  Country
// @Failure     400 {object} ErrorResponse
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
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /currency/{currency} [get]
func GetCountriesByCurrency(c *gin.Context) {
	currency := c.Param("currency")
	fields := c.Query("fields")

	filters := map[string]string{"currency": currency}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /demonym/{demonym} [get]
func GetCountriesByDemonym(c *gin.Context) {
	demonym := c.Param("demonym")
	fields := c.Query("fields")

	filters := map[string]string{"demonym": demonym}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /lang/{language} [get]
func GetCountriesByLanguage(c *gin.Context) {
	language := c.Param("language")
	fields := c.Query("fields")

	filters := map[string]string{"language": language}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /capital/{capital} [get]
func GetCountriesByCapital(c *gin.Context) {
	capital := c.Param("capital")
	fields := c.Query("fields")

	filters := map[string]string{"capital": capital}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /region/{region} [get]
func GetCountriesByRegion(c *gin.Context) {
	region := c.Param("region")
	fields := c.Query("fields")

	filters := map[string]string{"region": region}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /subregion/{subregion} [get]
func GetCountriesBySubregion(c *gin.Context) {
	subregion := c.Param("subregion")
	fields := c.Query("fields")

	filters := map[string]string{"subregion": subregion}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
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
// @Success     200 {array}  Country
// @Failure     404 {object} ErrorResponse
// @Router      /translation/{translation} [get]
func GetCountriesByTranslation(c *gin.Context) {
	translation := c.Param("translation")
	fields := c.Query("fields")

	filters := map[string]string{"translation": translation}
	filteredCountries := filterCountries(filters)

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		var result []map[string]interface{}
		for _, cty := range filteredCountries {
			result = append(result, selectFields(cty, fieldList))
		}
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, filteredCountries)
	}
}

// GetCountryByAlphaCode handles GET requests to /alpha/{code}.
func GetCountryByAlphaCode(c *gin.Context) {
	code := c.Param("code")
	fields := c.Query("fields")

	for _, country := range Countries {
		if strings.EqualFold(country.CCA2, code) ||
			strings.EqualFold(country.CCA3, code) ||
			strings.EqualFold(country.CCN3, code) ||
			strings.EqualFold(country.CIOC, code) {
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

// GetCountriesByIndependence godoc
// @Summary     Get countries by independence status
// @Description Get countries filtered by independence. Defaults to status=true if not specified.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       status query string false "true or false. Defaults to 'true'"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {array}  Country
// @Failure     400 {object} ErrorResponse
// @Router      /independent [get]
func GetCountriesByIndependence(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		status = "true"
	}

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

// GetCountryByCCN3 godoc
// @Summary     Get country by numeric ISO code (CCN3)
// @Description Get details of a specific country by its numeric ISO code.
// @Tags        Countries
// @Accept      json
// @Produce     json
// @Param       code   path  string true  "Numeric code (e.g., 840)"
// @Param       fields query string false "Comma-separated list of fields to include in the response"
// @Success     200 {object} Country
// @Failure     404 {object} ErrorResponse
// @Router      /ccn3/{code} [get]
func GetCountryByCCN3(c *gin.Context) {
	code := c.Param("code")
	fields := c.Query("fields")

	for _, country := range Countries {
		if strings.EqualFold(country.CCN3, code) {
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

// GetCountriesByCallingCode handles GET requests to /callingcode/{callingcode}.
func GetCountriesByCallingCode(c *gin.Context) {
	callingCode := c.Param("callingcode")
	fields := c.Query("fields")
	var filteredCountries []Country

	for _, country := range Countries {
		codeRoot := country.IDD.Root
		for _, suffix := range country.IDD.Suffixes {
			fullCode := strings.TrimSpace(codeRoot + suffix)
			// Remove '+' for comparison
			fullCode = strings.TrimPrefix(fullCode, "+")
			if fullCode == callingCode {
				filteredCountries = append(filteredCountries, country)
				break
			}
		}
	}

	if len(filteredCountries) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Country not found"})
		return
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
