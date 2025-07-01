# Torq - IP Geolocation Service

A Go-based microservice that provides IP geolocation functionality through a REST API. The service can determine the country of an IP address using various backend providers.

## Features

- 🌍 IP to country geolocation
- 🔌 Pluggable backend providers
- 🐳 Docker support
- 🧪 Comprehensive testing
- 📊 Health check endpoint
- 🔒 Security scanning

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd Torq
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your IP database provider configuration
   ```

4. **Run the application**
   ```bash
   # Using Go directly
   go run main.go

   # Or using Make
   make run
   ```

5. **Test the API**
   ```bash
   curl "http://localhost:8080/v1/find-country?ip=90.91.92.93"
   ```

### Using Docker

1. **Build the Docker image**
   ```bash
   make docker-build
   ```

2. **Run the container**
   ```bash
   make docker-run
   ```

3. **Test the API**
   ```bash
   curl "http://localhost:8080/v1/find-country?ip=90.91.92.93"
   ```

## API Documentation

### Find Country by IP

**Endpoint:** `GET /v1/find-country`

**Query Parameters:**
- `ip` (required): The IP address to look up

**Example Request:**
```bash
curl "http://localhost:8080/v1/find-country?ip=90.91.92.93"
```

**Example Response:**
```json
{
  "country": "France",
  "country_code": "FR",
  "ip": "90.91.92.93",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response:**
```json
{
  "error": "Invalid IP address",
  "message": "The provided IP address is not valid"
}
```

### Health Check

**Endpoint:** `GET /health`

**Example Request:**
```bash
curl "http://localhost:8080/health"
```

**Example Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Configuration

The service uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `IP_DB_PROVIDER` | IP database provider to use | `maxmind` |
| `PORT` | Server port | `8080` |

### Supported Providers

- **MaxMind**: Uses MaxMind GeoIP2 database
- **IP2Location**: Uses IP2Location database
- **Custom**: Custom provider implementation

## Development

### Available Make Commands

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Run security scan
make security

# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Stop Docker container
make docker-stop

# Clean build artifacts
make clean

# Show all available commands
make help
```

### Project Structure

```
Torq/
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── go.sum                  # Go module checksums
├── Makefile               # Build automation
├── Dockerfile             # Docker configuration
├── .dockerignore          # Docker ignore file
├── .github/
│   └── workflows/
│       └── ci.yml         # GitHub Actions CI/CD
├── lookup/                # IP lookup providers
├── CountryFinder/         # Country finder logic
└── README.md              # This file
```

## Testing

### Run Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./lookup -v
```

### Test Examples
```bash
# Test the API endpoint
curl "http://localhost:8080/v1/find-country?ip=8.8.8.8"

# Test with invalid IP
curl "http://localhost:8080/v1/find-country?ip=invalid"

# Test health endpoint
curl "http://localhost:8080/health"
```

## Deployment

### Docker Deployment

1. **Build the image**
   ```bash
   docker build -t torq:latest .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 -e IP_DB_PROVIDER=maxmind torq:latest
   ```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: torq
spec:
  replicas: 3
  selector:
    matchLabels:
      app: torq
  template:
    metadata:
      labels:
        app: torq
    spec:
      containers:
      - name: torq
        image: torq:latest
        ports:
        - containerPort: 8080
        env:
        - name: IP_DB_PROVIDER
          value: "maxmind"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the GitHub repository or contact the development team.