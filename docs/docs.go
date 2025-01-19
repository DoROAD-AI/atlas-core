// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://atlas.doroad.io/terms/",
        "contact": {
            "name": "Atlas API Support",
            "url": "https://github.com/DoROAD-AI/atlas/issues",
            "email": "support@doroad.ai"
        },
        "license": {
            "name": "MIT",
            "url": "https://github.com/DoROAD-AI/atlas/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/alpha": {
            "get": {
                "description": "Get countries matching a list of codes (CCA2, CCN3, CCA3, or CIOC).",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by codes",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Comma-separated list of country codes (CCA2, CCN3, CCA3, CIOC)",
                        "name": "codes",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/capital/{capital}": {
            "get": {
                "description": "Get countries matching a capital city name.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by capital",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Capital city name",
                        "name": "capital",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/ccn3/{code}": {
            "get": {
                "description": "Get details of a specific country by its numeric ISO code.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get country by numeric ISO code (CCN3)",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Numeric code (e.g., 840)",
                        "name": "code",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/v1.Country"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/countries": {
            "get": {
                "description": "Get details of all countries, with optional filters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get all countries",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Filter by independent status (true or false)",
                        "name": "independent",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/countries/{code}": {
            "get": {
                "description": "Get details of a specific country by its code (CCA2 or CCA3).",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get country by code",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Country code (CCA2 or CCA3)",
                        "name": "code",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/v1.Country"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/currency/{currency}": {
            "get": {
                "description": "Get countries matching a currency code or name.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by currency",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Currency code or name",
                        "name": "currency",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/demonym/{demonym}": {
            "get": {
                "description": "Get countries matching a demonym.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by demonym",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Demonym",
                        "name": "demonym",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/independent": {
            "get": {
                "description": "Get countries filtered by independence. Defaults to status=true if not specified.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by independence status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "true or false. Defaults to 'true'",
                        "name": "status",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/lang/{language}": {
            "get": {
                "description": "Get countries matching a language code or name.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by language",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Language code or name",
                        "name": "language",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/name/{name}": {
            "get": {
                "description": "Get countries matching a name query (common or official). Use fullText=true for exact name match.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by name",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Country name (common or official)",
                        "name": "name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Exact match for full name (true/false)",
                        "name": "fullText",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/region/{region}": {
            "get": {
                "description": "Get countries matching a region.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by region",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Region name",
                        "name": "region",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/subregion/{subregion}": {
            "get": {
                "description": "Get countries matching a subregion.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by subregion",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subregion name",
                        "name": "subregion",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/translation/{translation}": {
            "get": {
                "description": "Get countries matching a translation.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Countries"
                ],
                "summary": "Get countries by translation",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Translation",
                        "name": "translation",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Comma-separated list of fields to include in the response",
                        "name": "fields",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/v1.Country"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/v1.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "v1.CapitalInfo": {
            "type": "object",
            "properties": {
                "latlng": {
                    "type": "array",
                    "items": {
                        "type": "number"
                    },
                    "example": [
                        38.8951,
                        77.0364
                    ]
                }
            }
        },
        "v1.Car": {
            "type": "object",
            "properties": {
                "side": {
                    "type": "string",
                    "example": "right"
                },
                "signs": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "USA"
                    ]
                }
            }
        },
        "v1.CoatOfArms": {
            "type": "object",
            "properties": {
                "png": {
                    "type": "string",
                    "example": "https://mainfacts.com/media/images/coats_of_arms/us.png"
                },
                "svg": {
                    "type": "string",
                    "example": "https://mainfacts.com/media/images/coats_of_arms/us.svg"
                }
            }
        },
        "v1.Country": {
            "type": "object",
            "properties": {
                "altSpellings": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "area": {
                    "type": "number",
                    "example": 9372610
                },
                "borders": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "capital": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "Washington",
                        " D.C."
                    ]
                },
                "capitalInfo": {
                    "$ref": "#/definitions/v1.CapitalInfo"
                },
                "car": {
                    "$ref": "#/definitions/v1.Car"
                },
                "cca2": {
                    "type": "string",
                    "example": "US"
                },
                "cca3": {
                    "type": "string",
                    "example": "USA"
                },
                "ccn3": {
                    "type": "string",
                    "example": "840"
                },
                "cioc": {
                    "type": "string",
                    "example": "USA"
                },
                "coatOfArms": {
                    "$ref": "#/definitions/v1.CoatOfArms"
                },
                "continents": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "currencies": {
                    "$ref": "#/definitions/v1.Currencies"
                },
                "demonyms": {
                    "$ref": "#/definitions/v1.Demonyms"
                },
                "fifa": {
                    "type": "string",
                    "example": "USA"
                },
                "flag": {
                    "type": "string",
                    "example": "🇺🇸"
                },
                "flags": {
                    "$ref": "#/definitions/v1.Flags"
                },
                "gini": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "number"
                    }
                },
                "idd": {
                    "$ref": "#/definitions/v1.IDD"
                },
                "independent": {
                    "type": "boolean",
                    "example": true
                },
                "landlocked": {
                    "type": "boolean",
                    "example": false
                },
                "languages": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "latlng": {
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                },
                "maps": {
                    "$ref": "#/definitions/v1.Maps"
                },
                "name": {
                    "$ref": "#/definitions/v1.Name"
                },
                "population": {
                    "type": "integer",
                    "example": 334805269
                },
                "postalCode": {
                    "$ref": "#/definitions/v1.PostalCode"
                },
                "region": {
                    "type": "string",
                    "example": "Americas"
                },
                "startOfWeek": {
                    "type": "string",
                    "example": "sunday"
                },
                "status": {
                    "type": "string",
                    "example": "officially-assigned"
                },
                "subregion": {
                    "type": "string",
                    "example": "North America"
                },
                "timezones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "tld": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "translations": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "object",
                        "properties": {
                            "common": {
                                "type": "string"
                            },
                            "official": {
                                "type": "string"
                            }
                        }
                    }
                },
                "unMember": {
                    "type": "boolean",
                    "example": true
                }
            }
        },
        "v1.Currencies": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/v1.CurrencyInfo"
            }
        },
        "v1.CurrencyInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "example": "US Dollar"
                },
                "symbol": {
                    "type": "string",
                    "example": "$"
                }
            }
        },
        "v1.DemonymInfo": {
            "type": "object",
            "properties": {
                "f": {
                    "type": "string",
                    "example": "American"
                },
                "m": {
                    "type": "string",
                    "example": "American"
                }
            }
        },
        "v1.Demonyms": {
            "type": "object",
            "properties": {
                "eng": {
                    "$ref": "#/definitions/v1.DemonymInfo"
                },
                "fra": {
                    "$ref": "#/definitions/v1.DemonymInfo"
                }
            }
        },
        "v1.ErrorResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "Bad request"
                }
            }
        },
        "v1.Flags": {
            "type": "object",
            "properties": {
                "alt": {
                    "type": "string",
                    "example": "Flag of the United States"
                },
                "png": {
                    "type": "string",
                    "example": "https://restcountries.eu/data/usa.png"
                },
                "svg": {
                    "type": "string",
                    "example": "https://restcountries.eu/data/usa.svg"
                }
            }
        },
        "v1.IDD": {
            "type": "object",
            "properties": {
                "root": {
                    "type": "string",
                    "example": "+1"
                },
                "suffixes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "201",
                        "202"
                    ]
                }
            }
        },
        "v1.Maps": {
            "type": "object",
            "properties": {
                "googleMaps": {
                    "type": "string",
                    "example": "https://goo.gl/maps/..."
                },
                "openStreetMaps": {
                    "type": "string",
                    "example": "https://www.openstreetmap.org/..."
                }
            }
        },
        "v1.Name": {
            "type": "object",
            "properties": {
                "common": {
                    "type": "string",
                    "example": "United States"
                },
                "official": {
                    "type": "string",
                    "example": "United States of America"
                }
            }
        },
        "v1.PostalCode": {
            "type": "object",
            "properties": {
                "format": {
                    "type": "string",
                    "example": "#####-####"
                },
                "regex": {
                    "type": "string",
                    "example": "^\\d{5}(-\\d{4})?$"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/v1",
	Schemes:          []string{"https", "http"},
	Title:            "Atlas - Geographic Data API by DoROAD",
	Description:      "A comprehensive REST API providing detailed country information worldwide. This modern, high-performance service offers extensive data about countries, including demographics, geography, and international codes.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
