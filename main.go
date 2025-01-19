package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	v1 "github.com/DoROAD-AI/atlas/api/v1"
	"github.com/DoROAD-AI/atlas/docs"
	"github.com/gin-contrib/cors"
)

// @title       Atlas - Geographic Data API by DoROAD
// @version     1.0
// @description A comprehensive REST API providing detailed country information worldwide. This modern, high-performance service offers extensive data about countries, including demographics, geography, and international codes.
// @termsOfService http://atlas.doroad.io/terms/
// @contact.name  Atlas API Support
// @contact.url   https://github.com/DoROAD-AI/atlas/issues
// @contact.email support@doroad.ai
// @license.name  MIT
// @license.url   https://github.com/DoROAD-AI/atlas/blob/main/LICENSE
// @BasePath      /v1
// @schemes       https http

func getHost() string {
	env := os.Getenv("ATLAS_ENV")
	switch env {
	case "production":
		return "atlas.doroad.io"
	case "test":
		return "atlas.doroad.dev"
	case "dev":
		return "localhost:3101"
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
		log.Fatalf("Failed to initialize API: %v", err)
	}

	// Create Gin router with default middleware
	router := gin.Default()

	// Enable CORS
	router.Use(cors.Default())

	// Dynamically set Swagger host
	docs.SwaggerInfo.Host = getHost()

	// v1 routes
	v1Group := router.Group("/v1")
	{
		// Existing routes
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
		// New route for numeric ISO code
		v1Group.GET("/ccn3/:code", v1.GetCountryByCCN3)
		// Alias for "/v1/all" -> same as "/v1/countries"
		v1Group.GET("/all", v1.GetCountries)
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
