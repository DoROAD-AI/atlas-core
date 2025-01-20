# Atlas 🌍

**Atlas** is DoROAD's flagship Global Travel and Aviation Intelligence Data API. Version 2.0 represents a significant leap forward, providing a comprehensive, high-performance RESTful API for accessing detailed country information, extensive airport data, and up-to-date passport visa requirements.

**Atlas v2.0 is a proprietary product designed for enterprise use.** It offers advanced features, enhanced data coverage, and the reliability required for mission-critical applications in the travel, aviation, and logistics industries.

## Key Features of Atlas v2.0

Atlas v2.0 builds upon the foundation of v1, adding powerful new capabilities and expanding existing ones:

### 1. Enhanced Country Data (v1 Features)

*   **Comprehensive Data:** In-depth information on all countries, including demographics, geography, economy, international codes, and more.
*   **Flexible Querying:** Search by:
    *   Name (full or partial matching)
    *   Country codes (CCA2, CCN3, CCA3, CIOC, FIFA)
    *   Currency
    *   Language
    *   Capital city
    *   Region and subregion
    *   Translations
    *   Demonyms
    *   Independence status
    *   Calling code
*   **Field Filtering:** Control the size of responses by selecting specific fields.
*   **Case-Insensitive Search:** User-friendly searching.
*   **Input Validation:** Robust error handling for invalid parameters.

### 2. Advanced Airport Data & Intelligence

*   **Global Airport Coverage:** A comprehensive database of airports worldwide, including major hubs and smaller airfields.
*   **Detailed Airport Information:**
    *   ICAO and IATA codes
    *   Airport type (e.g., medium_airport, large_airport, heliport, closed)
    *   Name, latitude, longitude, elevation
    *   Continent, ISO country code, ISO region code
    *   Municipality
    *   Scheduled service status (yes/no)
    *   Links to home page and Wikipedia
    *   Keywords
*   **Runway Data:**
    *   Runway length and width
    *   Surface type
    *   Lighting status
    *   Closed status
    *   Runway identifiers (e.g., 04/22)
    *   Latitude, longitude, elevation, and heading for both ends
    *   Displaced threshold information
*   **Communication Frequencies:**
    *   Frequency type (e.g., APP, ATIS, GND, TWR)
    *   Description
    *   Frequency in MHz
*   **Advanced Airport Queries:**
    *   **By Code:** Retrieve an airport directly by its ICAO or IATA code.
    *   **By Region:** Find all airports within a given ISO region.
    *   **By Municipality:** Get airports in a specific city or town.
    *   **By Type:** Filter airports based on their type.
    *   **By Scheduled Service:** Identify airports with scheduled commercial flights.
    *   **By Keyword:** Search for airports using relevant keywords.
*   **Geospatial Queries:**
    *   **Within Radius:** Find all airports within a specified radius of a given latitude/longitude.
*   **Distance Calculation:** Calculate the distance between two airports using their ICAO or IATA codes.
*   **Flexible Search:** Search across multiple fields (name, city, codes) for quick airport lookups.

### 3. Comprehensive Passport and Visa Data

*   **Passport Data:** Access visa requirements for all passport/destination country combinations.
*   **Visa Requirement Details:**
    *   Visa-free status
    *   Visa on arrival
    *   E-visa
    *   Visa required
    *   Duration of stay (if applicable)
*   **Advanced Visa Queries:**
    *   **Visa-Free Destinations:** Get a list of countries where a given passport holder can travel visa-free.
    *   **Visa-on-Arrival Destinations:** List countries offering visa on arrival for a specific passport.
    *   **E-Visa Destinations:** Find countries where a passport holder can apply for an e-visa.
    *   **Visa-Required Destinations:** Identify countries requiring a pre-arranged visa for a given passport.
    *   **Visa Requirement Details:** Get specific visa details (duration, type) for a passport and destination.
    *   **Reciprocal Visa Agreements:** Check visa requirements both ways between two countries.
    *   **Visa Requirement Comparison:** Compare visa requirements for multiple passports to a single destination.
    *   **Passport Ranking:** Get a ranked list of passports based on the number of visa-free destinations.
    *   **Common Visa-Free Destinations:** Find shared visa-free destinations for a group of passports.

### 4. Enterprise-Grade Features

*   **High Performance:** Optimized for speed and efficiency to handle large volumes of requests.
*   **Scalability:** Designed to scale horizontally to meet growing demands.
*   **Reliability:** Built with robust error handling and redundancy for maximum uptime.
*   **Security:** Secure API endpoints (HTTPS) and data protection measures.
*   **API Documentation:** Comprehensive documentation using Swagger (OpenAPI) for easy integration.
*   **Proprietary Data and Algorithms:** Atlas v2.0 incorporates proprietary data and algorithms for enhanced accuracy and insights.

### 5. Super Type Query: Advanced Cross-Domain AI-Powered Search

Atlas v2.0 introduces a sophisticated **Super Type Query** feature that enables users to perform advanced cross-domain searches across all data types, including **countries** and **airports**. This capability is particularly designed to support enterprise applications and AI models that require flexible and comprehensive data retrieval.

**Key Features:**

*   **Unified Search Endpoint:** A single endpoint to search across multiple data domains, reducing complexity and improving efficiency.
*   **Dynamic Filtering:** Support for a wide range of query parameters to fine-tune search results according to specific needs.
*   **AI Integration:** Facilitates AI and machine learning applications by providing rich, multi-faceted data essential for training and inference.
*   **Optimized for Enterprises:** Designed to handle complex queries efficiently, making it suitable for large-scale enterprise systems.

**Example Use Cases:**

*   **Travel and Logistics Planning:** Integrate country and airport data to optimize routes, logistics, and supply chain operations.
*   **Market Analysis:** Analyze regional data across different countries and airports to identify market opportunities.
*   **Risk Assessment:** Aggregate data for countries and airports to assess geopolitical risks, travel advisories, and compliance requirements.
*   **AI Model Training:** Provide comprehensive datasets for training AI models in natural language processing, predictive analytics, and more.

## API Endpoints (v2.0)

**Base URL:** `https://atlas.doroad.io/v2` (Production)

**Note:**  Access to v2.0 endpoints requires a commercial license.

### Country Endpoints

*   **`GET /v2/all`**: Get all countries (with optional field filtering).
*   **`GET /v2/countries`**: Same as `/v2/all`.
*   **`GET /v2/countries/{code}`**: Get a country by its CCA2, CCA3, CCN3, CIOC, or FIFA code.
*   **`GET /v2/name/{name}`**: Search for countries by name (partial or full).
*   **`GET /v2/alpha`**: Get countries by a list of CCA2 codes.
*   **`GET /v2/currency/{currency}`**: Search for countries by currency code.
*   **`GET /v2/demonym/{demonym}`**: Search for countries by demonym.
*   **`GET /v2/lang/{language}`**: Search for countries by language code.
*   **`GET /v2/capital/{capital}`**: Search for countries by capital city.
*   **`GET /v2/region/{region}`**: Search for countries by region.
*   **`GET /v2/subregion/{subregion}`**: Search for countries by subregion.
*   **`GET /v2/translation/{translation}`**: Search for countries by translated name.
*   **`GET /v2/independent`**: Get independent or non-independent countries.
*   **`GET /v2/alpha/{code}`**: Get a country by its CCA2 code.
*   **`GET /v2/ccn3/{code}`**: Get a country by its CCN3 code.
*   **`GET /v2/callingcode/{callingcode}`**: Get countries by calling code.

### Airport Endpoints

*   **`GET /v2/airports`**: Get all airports (grouped by country).
*   **`GET /v2/airports/{countryCode}`**: Get airports in a specific country.
*   **`GET /v2/airports/{countryCode}/{airportIdent}`**: Get an airport by ICAO or IATA code within a country.
*   **`GET /v2/airports/by-code/{airportCode}`**: Get an airport by ICAO or IATA code (globally).
*   **`GET /v2/airports/region/{isoRegion}`**: Get airports in an ISO region.
*   **`GET /v2/airports/municipality/{municipalityName}`**: Get airports in a municipality.
*   **`GET /v2/airports/type/{airportType}`**: Get airports of a specific type.
*   **`GET /v2/airports/scheduled`**: Get airports with scheduled service.
*   **`GET /v2/airports/{countryCode}/{airportIdent}/runways`**: Get runway information for an airport.
*   **`GET /v2/airports/{countryCode}/{airportIdent}/frequencies`**: Get communication frequencies for an airport.
*   **`GET /v2/airports/search?query={searchString}`**: Search for airports by name, city, or code.
*   **`GET /v2/airports/radius?latitude={latitude}&longitude={longitude}&radius={radiusInKm}`**: Get airports within a radius.
*   **`GET /v2/airports/distance?airport1={airportCode1}&airport2={airportCode2}`**: Calculate the distance between two airports.
*   **`GET /v2/airports/keyword/{keyword}`**: Get airports by keyword.

### Passport and Visa Endpoints

*   **`GET /v2/passports/{passportCode}`**: Get visa requirements for a passport.
*   **`GET /v2/passports/{passportCode}/visas`**: Same as `/v2/passports/{passportCode}`.
*   **`GET /v2/passports/visa?fromCountry={fromCountry}&toCountry={toCountry}`**: Get visa requirements between two countries.
*   **`GET /v2/passports/{passportCode}/visa-free`**: Get visa-free destinations for a passport.
*   **`GET /v2/passports/{passportCode}/visa-on-arrival`**: Get visa-on-arrival destinations for a passport.
*   **`GET /v2/passports/{passportCode}/e-visa`**: Get e-visa destinations for a passport.
*   **`GET /v2/passports/{passportCode}/visa-required`**: Get visa-required destinations for a passport.
*   **`GET /v2/passports/{passportCode}/visa-details/{destinationCode}`**: Get detailed visa requirements for a passport and destination.
*   **`GET /v2/passports/reciprocal/{countryCode1}/{countryCode2}`**: Get reciprocal visa requirements between two countries.
*   **`GET /v2/passports/compare?passports={passportCode1},{passportCode2},...&destination={destinationCode}`**: Compare visa requirements for multiple passports to a destination.
*   **`GET /v2/passports/ranking`**: Get a ranked list of passports based on visa-free access.
*   **`GET /v2/passports/common-visa-free?passports={passportCode1},{passportCode2},...`**: Find common visa-free destinations for multiple passports.

### Advanced Search Endpoint

*   **`GET /v2/search`**: Perform a comprehensive search across all data types (countries, airports) based on provided query parameters.

    **Query Parameters**:

    *   **type**: *(optional)* Type of data to search for (`country`, `airport`). If omitted or set to `all`, searches across all data types.
    *   **name**: *(optional)* Name of the country or airport.
    *   **region**: *(optional)* Region of the country.
    *   **subregion**: *(optional)* Subregion of the country.
    *   **cca2**: *(optional)* Country code Alpha-2.
    *   **cca3**: *(optional)* Country code Alpha-3.
    *   **ccn3**: *(optional)* Country code Numeric.
    *   **capital**: *(optional)* Capital city of the country.
    *   **ident**: *(optional)* Airport Ident code (e.g., ICAO code).
    *   **iata_code**: *(optional)* Airport IATA code.
    *   **iso_country**: *(optional)* ISO country code for airports.
    *   **iso_region**: *(optional)* ISO region code for airports.
    *   **municipality**: *(optional)* Municipality of the airport.
    *   **airport_type**: *(optional)* Type of the airport (e.g., `medium_airport`, `heliport`).

    **Description**:

    The Super Type Query endpoint allows for flexible and comprehensive search capabilities across all data types in the Atlas API. By specifying various query parameters, users can retrieve data matching specific criteria, facilitating complex data retrieval needs for enterprise applications and AI models.

    **Examples**:

    *   **Search for countries in the region "Europe":**

        ```
        GET /v2/search?type=country&region=Europe
        ```

    *   **Search for airports with the municipality "Kingstown":**

        ```
        GET /v2/search?type=airport&municipality=Kingstown
        ```

    *   **Search across all data types for the name "Argyle":**

        ```
        GET /v2/search?name=Argyle
        ```

    **Use Cases in Enterprise and AI Context**:

    *   **AI-Powered Analytics:** Retrieve data across countries and airports to feed into AI models for predictive analytics, risk assessment, or optimization algorithms.
    *   **Enterprise Data Integration:** Seamlessly integrate comprehensive data into enterprise systems for unified reporting, dashboards, or decision-support tools.
    *   **Custom Applications:** Build tailored applications requiring dynamic data retrieval across multiple domains without needing multiple API calls.
    *   **Data Mining and Exploration:** Facilitate exploratory data analysis by applying flexible search criteria to discover patterns and insights.

    **Response Structure**:

    Depending on the `type` parameter, the response will include matched results from the specified data domain(s):

    *   If `type=country`, the response will be:

        ```json
        [
          {
            /* Country data */
          },
        ]
        ```

    *   If `type=airport`, the response will be:

        ```json
        [
          {
            /* Airport data */
          },
        ]
        ```

    *   If `type` is omitted or set to `all`, the response will include both countries and airports:

        ```json
        {
          "countries": [
            {
              /* Country data */
            },
          ],
          "airports": [
            {
              /* Airport data */
            },
          ]
        }
        ```

    **Notes**:

    *   At least one query parameter (other than `type`) must be provided.
    *   If no matches are found, the API will return a `404 Not Found` response.
    *   Supports partial and case-insensitive matching where applicable.

    **Error Handling**:

    *   **400 Bad Request**: Returned if required parameters are missing or invalid.
    *   **404 Not Found**: Returned if no matching data is found.

## Building and Running Atlas v2.0 (Internal Instructions)

These instructions are for internal DoROAD teams building and running Atlas v2.0 from source.

### Prerequisites

*   **Go:** Version 1.20 or higher.
*   **Git:** For version control.
*   **Docker (Optional):** For containerized deployment.
*   **Commercial License:** A valid commercial license for Atlas v2.0 is required.

### Setting Up the Development Environment

1.  **Clone the Repository:**

    ```bash
    git clone https://github.com/DoROAD-AI/atlas.git
    cd atlas
    ```

2.  **Initialize Go Modules:**

    ```bash
    go mod init github.com/DoROAD-AI/atlas
    go mod tidy
    ```

3.  **Install Swagger Tools (for documentation):**

    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```

### Configuring the Environment

Atlas uses environment variables for configuration.

1.  **`ATLAS_ENV`:** Sets the environment:

    *   `development`: For local development (default).
    *   `test`: For the test/staging environment.
    *   `production`: For the production environment.

2.  **`PORT` (Optional):** Sets the port the API will listen on (default: `3101`).

**Example:**

```bash
# For development on localhost:3101
export ATLAS_ENV=development

# For production, served on port 80
export ATLAS_ENV=production
export PORT=80
```

### Building and Running the API

**1. Building:**

```bash
go build -o atlas .
```

**2. Running:**

```bash
./atlas
```

Or use `go run` for development:

```bash
go clean
go build
go run main.go
```

**Generating Swagger Documentation:**

```bash
swag init
```

This will generate the `docs` folder with the Swagger (OpenAPI) documentation. You can access the interactive documentation at `http://localhost:3101/swagger/index.html` (replace `3101` with your configured port if necessary).

### Docker Deployment (Optional)

1.  **Create a `Dockerfile`:**

    ```dockerfile
    FROM golang:1.20-alpine
    WORKDIR /app
    COPY . .
    RUN go mod download
    RUN go build -o atlas .
    EXPOSE 3101
    CMD ["./atlas"]
    ```

2.  **Build the Docker Image:**

    ```bash
    docker build -t atlas-v2 .
    ```

3.  **Run the Docker Container:**

    ```bash
    docker run -p 3101:3101 -e ATLAS_ENV=production atlas-v2
    ```

    *   `-p 3101:3101` maps port 3101 on your host to port 3101 inside the container.
    *   `-e ATLAS_ENV=production` sets the environment variable within the container.

## License

**Atlas v1** is licensed under the **MIT License**.

**Atlas v2 and later** are **proprietary** and require a **commercial license** from DoROAD. Please contact `support@doroad.ai` for information on obtaining a commercial license.

## Support

*   **Internal Support:** Contact the technical team for assistance.
*   **Commercial License Support:**  `support@doroad.ai`

---

Made with ❤️ by DoROAD's Roadman Team