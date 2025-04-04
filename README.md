# Atlas-Core 🌍

**Atlas-Core** is the open-source foundation of DoROAD's flagship Global Travel and Aviation Intelligence Data API. Built with modern Go practices, this powerful geographic information API provides detailed country data and geographic insights through a clean, RESTful interface.

## About Atlas-Core

Atlas-Core brings enterprise-grade geographic data to the open-source community. By releasing this foundation, we aim to:

- Contribute valuable tools to developers working on travel, geographic, and AI applications
- Build a community around geographic data standardization and enrichment
- Provide a springboard for innovative applications in travel technology
- Create awareness for the full Atlas product, which includes additional premium features

## Using Atlas-Core

### Via DoROAD's Hosted Service

The fastest way to get started with Atlas-Core is through our hosted API service:

- **Sign Up for API Access**: [https://portal.doroad.dev](https://portal.doroad.dev)
- **API Base URL**: [https://api.doroad.dev](https://api.doroad.dev)
- **Documentation**: [https://api.doroad.dev](https://api.doroad.dev)

### Authentication

After registering at [portal.doroad.dev](https://portal.doroad.dev), you'll receive an API key. Include this key in your requests:

```bash
curl -H "dapi-key: your_api_key" https://api.doroad.dev/atlasc/countries
```

### Quick Examples

```bash
# Get all countries
curl -H "dapi-key: your_api_key" https://api.doroad.dev/atlasc/countries

# Search by country name
curl -H "dapi-key: your_api_key" https://api.doroad.dev/atlasc/name/united

# Get specific fields for a country
curl -H "dapi-key: your_api_key" https://api.doroad.dev/atlasc/countries?fields=name,capital,currencies
```

## Features

### Core Features

- **Complete Country Information**: Comprehensive country data worldwide
- **Flexible Querying**: Multiple search criteria including:
  - Name (full/partial matching)
  - Country codes (CCA2, CCN3, CCA3, CIOC)
  - Currency
  - Language
  - Capital city
  - Region and subregion
  - Translations
  - Demonyms
  - Independence status
  - Calling code
- **Field Filtering**: Optimize response payload size
- **Modern API Design**: RESTful architecture with JSON responses
- **Interactive Documentation**: Swagger UI for easy exploration
- **Case-Insensitive Search**: Flexible searching
- **Input Validation**: Built-in parameter validation

### AI Integration Capabilities

Atlas-Core is designed to seamlessly integrate with AI and machine learning projects:

- **Structured Geographic Data**: Perfect for training location-aware models
- **Consistent Data Format**: Reliable structure for training data
- **Comprehensive Metadata**: Rich attributes for feature engineering
- **Clean REST Interface**: Easy to incorporate into AI pipelines
- **Batch Processing Support**: Efficient data retrieval for model training

**AI Use Cases:**
- Natural language processing for location entities
- Geospatial analysis and visualization
- Travel recommendation systems
- Location-based sentiment analysis
- Geographic classification models
- Transport and logistics optimization

## Self-Hosting Atlas-Core

### Prerequisites

- Go 1.20 or higher
- Git
- Docker (optional)

### Local Development Setup

1. **Clone and set up the project:**

   ```bash
   # Clone the repository
   git clone https://github.com/DoROAD-AI/atlas-core.git
   cd atlas-core

   # Initialize the module
   go mod tidy

   # Install Swagger tools
   go install github.com/swaggo/swag/cmd/swag@latest

   # Generate Swagger documentation
   swag init
   ```

2. **Configure the environment:**

   ```bash
   # Development mode on localhost:3101
   export ATLAS_ENV=development
   ```

3. **Run the server:**

   ```bash
   go run main.go
   ```

   or

   ```bash
   go clean
   go build
   ./atlas-core
   ```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o atlas-core
EXPOSE 3101
CMD ["./atlas-core"]
```

Build and run the Docker image:

```bash
docker build -t atlas-core .
docker run -p 3101:3101 -e ATLAS_ENV=production atlas-core
```

## API Documentation

Complete API documentation is available through our Swagger UI:

- **Hosted**: [https://api.doroad.dev](https://api.doroad.dev)
- **Local**: http://localhost:3101/swagger/index.html

## Difference Between Atlas-Core and Full Atlas

Atlas-Core is our open-source offering that provides fundamental country data capabilities. The full Atlas product offers additional premium features:

| Feature | Atlas-Core |
|---------|------------|
| Country Data | ✓ |
| Field Filtering | ✓ |
| Case-Insensitive Search | ✓ |
| Airport Data & Intelligence | - |
| Passport and Visa Data | - |
| Super Type Query | - |
| Enterprise-Grade Performance | - |
| Advanced Geospatial Queries | - |
| Support | Community |


For information on accessing the full Atlas product, please contact [atlas@doroad.ai](mailto:atlas@doroad.ai).

## Contributing

We welcome contributions to Atlas-Core! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Atlas-Core is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [REST Countries API](https://restcountries.com/)
- Built with [Gin Web Framework](https://gin-gonic.com/)
- Documentation powered by [Swag](https://github.com/swaggo/swag)

## Support and Contact

- GitHub Issues: [https://github.com/DoROAD-AI/atlas-core/issues](https://github.com/DoROAD-AI/atlas-core/issues)
- GitHub Discussions: [https://github.com/DoROAD-AI/atlas-core/discussions](https://github.com/DoROAD-AI/atlas-core/discussions)
- Email: [support@doroad.ai](mailto:support@doroad.ai)

## About DoROAD AI

DoROAD (DoRoad B.V.) is a pioneering travel technology company headquartered in The Netherlands, revolutionizing the travel industry through advanced technology and unwavering security and privacy protection.

Our vision extends beyond traditional travel technology, aiming to establish new standards for personalization, security, and efficiency in global travel management. By open-sourcing Atlas-Core, we're demonstrating our commitment to innovation and community collaboration.

---

Made with ❤️ by the DoROAD Team and 'E Roadman Dem.
