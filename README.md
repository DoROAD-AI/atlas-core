# Atlas 🌍

Atlas is a powerful, Go-based geographic information API that provides detailed country data and geographic insights. Built with modern Go practices and inspired by the REST Countries API, Atlas delivers comprehensive geographic data through a clean, RESTful interface.

## Using Atlas

### Via DoROAD's Hosted Service

The fastest way to get started with Atlas is through our hosted API service:

**Production Environment**
- Base URL: `https://atlas.doroad.io`
- Swagger Documentation: `https://atlas.doroad.io/swagger/index.html`

**Test/Staging Environment**
- Base URL: `https://atlas.doroad.dev`
- Swagger Documentation: `https://atlas.doroad.dev/swagger/index.html`

### Quick Examples

```bash
# Get all countries
curl https://atlas.doroad.io/v1/countries

# Search by country name
curl https://atlas.doroad.io/v1/name/united

# Get specific fields for a country
curl https://atlas.doroad.io/v1/countries?fields=name,capital,currencies
```

## Features

### Current Features
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
- **Field Filtering**: Optimize response payload size
- **Modern API Design**: RESTful architecture with JSON responses
- **Interactive Documentation**: Swagger UI for easy exploration
- **Case-Insensitive Search**: Flexible searching
- **Input Validation**: Built-in parameter validation

### Planned Features
- Airport information
- Geographic coordinate calculations
- City data
- Time zone utilities
- Geographic service integrations
- Advanced caching
- Rate limiting
- Response compression
- Usage metrics

## Self-Hosting Atlas

### Prerequisites
- Go 1.20 or higher
- Git
- Docker (optional)

### Local Development Setup

1. Clone and set up the project:
```bash
# Clone the repository
git clone https://github.com/DoROAD-AI/atlas.git
cd atlas

# Initialize the module
go mod init github.com/DoROAD-AI/atlas
go mod tidy

# Install Swagger tools
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swag init
```

2. Configure the environment:
```bash
# Development (localhost:8080)
export ATLAS_ENV=development

# Test environment (atlas.doroad.dev)
export ATLAS_ENV=test

# Production environment (atlas.doroad.io)
export ATLAS_ENV=production
```

3. Run the server:
```bash
go run main.go
```
or

```bash
go clean

go build

go run main.go
```

### Docker Deployment

```dockerfile
FROM golang:1.20-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o atlas

EXPOSE 8080
CMD ["./atlas"]
```

Build and run:
```bash
docker build -t atlas .
docker run -p 8080:8080 -e ATLAS_ENV=production atlas
```

## API Documentation

Complete API documentation is available through our Swagger UI:
- Production: https://atlas.doroad.io/swagger/index.html
- Test: https://atlas.doroad.dev/swagger/index.html
- Local: http://localhost:8080/swagger/index.html

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [REST Countries API](https://restcountries.com/)
- Built with [Gin Web Framework](https://gin-gonic.com/)
- Documentation powered by [Swag](https://github.com/swaggo/swag)

## Support and Contact

- GitHub Issues: [https://github.com/DoROAD-AI/atlas/issues](https://github.com/DoROAD-AI/atlas/issues)
- GitHub Discussions: [https://github.com/DoROAD-AI/atlas/discussions](https://github.com/DoROAD-AI/atlas/discussions)

## About DoROAD AI

DoROAD AI specializes in building intelligent infrastructure for modern applications. Visit [our GitHub](https://github.com/DoROAD-AI) to explore more of our open-source projects.

---
Made with ❤️ by DoROAD's Roadman Team