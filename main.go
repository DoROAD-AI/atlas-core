// main.go - Atlas API entry point
package main

import (
	"log"
	"net/http"
	"os"

	v1 "github.com/DoROAD-AI/atlas/api/v1"
	v2 "github.com/DoROAD-AI/atlas/api/v2"
	"github.com/DoROAD-AI/atlas/docs" // Swagger docs
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files" // Swagger files
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title       Atlas - Global Travel and Aviation Intelligence Data API by DoROAD
// @version     2.0
// @description Atlas is DoROAD's flagship Global Travel and Aviation Intelligence Data API. Version 2.0 represents a significant leap forward, providing a comprehensive, high-performance RESTful API for accessing detailed country information, extensive airport data, and up-to-date passport visa requirements worldwide. This service offers extensive data about countries (demographics, geography, international codes, etc.), airports, and visa regulations for various passports.
// @termsOfService http://atlas.doroad.io/terms/
// @contact.name  Atlas API Support
// @contact.url   https://github.com/DoROAD-AI/atlas/issues
// @contact.email support@doroad.ai
// @license.name  MIT / Proprietary
// @license.url   https://github.com/DoROAD-AI/atlas/blob/main/LICENSE
// @BasePath      /v2
// @schemes       https http
func getHost() string {
	env := os.Getenv("ATLAS_ENV")
	switch env {
	case "production":
		return "atlas.doroad.io"
	case "test":
		return "atlas.doroad.dev"
	case "dev":
		return "atlas-guauaxfgd2enghft.francecentral-01.azurewebsites.net"
	default:
		return "localhost:3101"
	}
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file in main.go, relying on environment variables.")
	} else {
		log.Println(".env file loaded successfully in main.go")
	}

	// Initialize OpenSkyClient with credentials (or leave empty for anonymous access)
	v2.InitializeOpenSkyClient(os.Getenv("OPENSKY_USERNAME"), os.Getenv("OPENSKY_PASSWORD"))

	// Set Gin mode based on environment
	env := os.Getenv("ATLAS_ENV")
	if env == "development" {
		gin.SetMode(gin.DebugMode) // Use DebugMode for development
	} else {
		gin.SetMode(gin.ReleaseMode) // Use ReleaseMode for other environments
	}

	// Load country data from JSON
	if err := v1.LoadCountriesSafe("data/countries.json"); err != nil {
		log.Fatalf("Failed to initialize country data: %v", err)
	}

	// Load passport data from JSON
	if err := v2.LoadPassportData("data/passports.json"); err != nil {
		log.Fatalf("Failed to initialize passport data: %v", err)
	}

	// Load airport data from JSON
	if err := v2.LoadAirportsData("data/airports.json"); err != nil {
		log.Fatalf("Failed to initialize airport data: %v", err)
	}

	// Create Gin router with default middleware
	router := gin.Default()

	// Enable CORS - Configure to be more restrictive in production
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // Be more specific in production
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Dynamically set Swagger host
	docs.SwaggerInfo.Host = getHost()

	//------------------------------------------------
	// v1 routes
	//------------------------------------------------
	v1Group := router.Group("/v1")
	{
		// restcountries.com v3.1 compatible routes
		v1Group.GET("/all", v1.GetCountries)
		v1Group.GET("/countries", v1.GetCountries)
		v1Group.GET("/countries/:code", v1.GetCountryByCode)
		v1Group.GET("/name/:name", v1.GetCountriesByName)
		v1Group.GET("/alpha", v1.GetCountriesByCodes)
		v1Group.GET("/currency/:currency", v1.GetCountriesByCurrency)
		v1Group.GET("/demonym/:demonym", v1.GetCountriesByDemonym)
		v1Group.GET("/lang/:language", v1.GetCountriesByLanguage)
		v1Group.GET("/capital/:capital", v1.GetCountriesByCapital)
		v1Group.GET("/region/:region", v1.GetCountriesByRegion)
		v1Group.GET("/subregion/:subregion", v1.GetCountriesBySubregion)
		v1Group.GET("/translation/:translation", v1.GetCountriesByTranslation)
		v1Group.GET("/independent", v1.GetCountriesByIndependence)
		v1Group.GET("/alpha/:code", v1.GetCountryByAlphaCode)
		v1Group.GET("/ccn3/:code", v1.GetCountryByCCN3)
		v1Group.GET("/callingcode/:callingcode", v1.GetCountriesByCallingCode)
	}

	//------------------------------------------------
	// v2 routes
	//------------------------------------------------
	v2Group := router.Group("/v2")
	{
		// Replicate all v1 routes under v2
		v2Group.GET("/all", v1.GetCountries)
		v2Group.GET("/countries", v1.GetCountries)
		v2Group.GET("/countries/:code", v1.GetCountryByCode)
		v2Group.GET("/name/:name", v1.GetCountriesByName)
		v2Group.GET("/alpha", v1.GetCountriesByCodes)
		v2Group.GET("/currency/:currency", v1.GetCountriesByCurrency)
		v2Group.GET("/demonym/:demonym", v1.GetCountriesByDemonym)
		v2Group.GET("/lang/:language", v1.GetCountriesByLanguage)
		v2Group.GET("/capital/:capital", v1.GetCountriesByCapital)
		v2Group.GET("/region/:region", v1.GetCountriesByRegion)
		v2Group.GET("/subregion/:subregion", v1.GetCountriesBySubregion)
		v2Group.GET("/translation/:translation", v1.GetCountriesByTranslation)
		v2Group.GET("/independent", v1.GetCountriesByIndependence)
		v2Group.GET("/alpha/:code", v1.GetCountryByAlphaCode)
		v2Group.GET("/ccn3/:code", v1.GetCountryByCCN3)
		v2Group.GET("/callingcode/:callingcode", v1.GetCountriesByCallingCode)

		// v2 passport routes
		v2Group.GET("/passports/:passportCode", v2.GetPassportData)
		v2Group.GET("/passports/:passportCode/visas", v2.GetVisaRequirementsForPassport)
		v2Group.GET("/passports/visa", v2.GetVisaRequirements)
		v2Group.GET("/passports/:passportCode/visa-free", v2.GetVisaFreeCountries)
		v2Group.GET("/passports/:passportCode/visa-on-arrival", v2.GetVisaOnArrivalCountries)
		v2Group.GET("/passports/:passportCode/e-visa", v2.GetEVisaCountries)
		v2Group.GET("/passports/:passportCode/visa-required", v2.GetVisaRequiredCountries)
		v2Group.GET("/passports/:passportCode/visa-details/:destinationCode", v2.GetVisaDetails)
		v2Group.GET("/passports/reciprocal/:countryCode1/:countryCode2", v2.GetReciprocalVisaRequirements)
		v2Group.GET("/passports/compare", v2.CompareVisaRequirements)
		v2Group.GET("/passports/ranking", v2.GetPassportRanking)
		v2Group.GET("/passports/common-visa-free", v2.GetCommonVisaFreeDestinations)

		// v2 airport routes
		v2Group.GET("/search", v2.SuperTypeQuery)
		v2Group.GET("/airports", v2.GetAllAirports)
		v2Group.GET("/airports/:countryCode", v2.GetAirportsByCountry)
		v2Group.GET("/airports/:countryCode/:airportIdent", v2.GetAirportByIdent)
		v2Group.GET("/airports/by-code/:airportCode", v2.GetAirportByCode)
		v2Group.GET("/airports/region/:isoRegion", v2.GetAirportsByRegion)
		v2Group.GET("/airports/municipality/:municipalityName", v2.GetAirportsByMunicipality)
		v2Group.GET("/airports/type/:airportType", v2.GetAirportsByType)
		v2Group.GET("/airports/scheduled", v2.GetAirportsWithScheduledService)
		v2Group.GET("/airports/:countryCode/:airportIdent/runways", v2.GetAirportRunways)
		v2Group.GET("/airports/:countryCode/:airportIdent/frequencies", v2.GetAirportFrequencies)
		v2Group.GET("/airports/search", v2.SearchAirports)
		v2Group.GET("/airports/radius", v2.GetAirportsWithinRadius)
		v2Group.GET("/airports/distance", v2.CalculateDistanceBetweenAirports)
		v2Group.GET("/airports/keyword/:keyword", v2.GetAirportsByKeyword)

		// Flights routes (OpenSky API integration)
		flightsGroup := v2Group.Group("/flights")
		{
			flightsGroup.GET("/states/all", v2.GetStatesAllHandler)
			flightsGroup.GET("/my-states", v2.GetMyStatesHandler)
			flightsGroup.GET("/interval", v2.GetFlightsIntervalHandler)
			flightsGroup.GET("/aircraft/:icao24", v2.GetFlightsByAircraftHandlerV2)
			flightsGroup.GET("/arrivals/:airport", v2.GetArrivalsByAirportHandlerV2)
			flightsGroup.GET("/departures/:airport", v2.GetDeparturesByAirportHandlerV2)
			flightsGroup.GET("/track", v2.GetTrackByAircraftHandler)
		}
	}

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Determine port, default to 3101
	port := os.Getenv("PORT")
	if port == "" {
		port = "3101"
	}

	// Start server
	router.Run(":" + port)
}
