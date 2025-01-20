// main.go - Atlas API entry point
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files" // Swagger files
	ginSwagger "github.com/swaggo/gin-swagger"

	v1 "github.com/DoROAD-AI/atlas/api/v1"
	v2 "github.com/DoROAD-AI/atlas/api/v2"

	"github.com/DoROAD-AI/atlas/docs" // Swagger docs
	"github.com/gin-contrib/cors"
)

// @title       Atlas - Geographic, Airport, and Passport Data API by DoROAD
// @version     2.0
// @description A comprehensive REST API providing detailed country information, airport data, and passport visa requirements worldwide. This service offers extensive data about countries (demographics, geography, international codes, etc.), airports, and visa regulations for various passports.
// @termsOfService http://atlas.doroad.io/terms/

// @contact.name  Atlas API Support
// @contact.url   https://github.com/DoROAD-AI/atlas/issues
// @contact.email support@doroad.ai

// @license.name  MIT
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
	// Set Gin mode based on environment
	env := os.Getenv("ATLAS_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load country data from JSON
	if err := v1.LoadCountriesSafe("countries.json"); err != nil {
		log.Fatalf("Failed to initialize country data: %v", err)
	}

	// Load passport data from JSON
	if err := v2.LoadPassportData("passports.json"); err != nil {
		log.Fatalf("Failed to initialize passport data: %v", err)
	}

	// Load airport data from JSON
	if err := v2.LoadAirportsData("airports.json"); err != nil {
		log.Fatalf("Failed to initialize airport data: %v", err)
	}

	// Create Gin router with default middleware
	router := gin.Default()

	// Enable CORS
	router.Use(cors.Default())

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

		// v2 airport routes
		v2Group.GET("/airports", v2.GetAllAirports)
		v2Group.GET("/airports/:countryCode", v2.GetAirportsByCountry)
		v2Group.GET("/airports/:countryCode/:airportIdent", v2.GetAirportByIdent)
	}

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Determine port, default to 3101
	port := os.Getenv("PORT")
	if port == "" {
		port = "3101"
	}

	// Start server
	router.Run(":" + port)
}
