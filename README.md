# Torq - IP Geolocation Service

A Go-based microservice that provides IP geolocation functionality through a REST API. The service can determine the country of an IP address using various backend providers with comprehensive metrics and observability.

## Features

- 🌍 IP to country geolocation
- 🔌 Pluggable backend providers with JSON configuration
- 📊 OpenTelemetry metrics and tracing
- 🐳 Docker support
- 🧪 Comprehensive testing
- 📈 Prometheus metrics endpoint
- 🔒 Security scanning
- 🏗️ Clean architecture with internal packages

## Quick Start

### Prerequisites

- Go 1.24 or higher
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

3. **Set up database configuration**
   
   **Option 1: JSON Configuration (Recommended)**
   ```bash
   export DB_CONFIG='{"dbtype": "csv", "extra_details": {"file_path": "./TestFiles/ip_data.csv"}}'
   ```
   
   **Option 2: Legacy Environment Variables**
   ```bash
   export IP_DB_PROVIDER=csv
   export IP_DB_PATH=./TestFiles/ip_data.csv
   ```

4. **Run the application**
   ```bash
   # Using Go directly
   go run cmd/main.go

   # Or using Make
   make run
   ```

5. **Test the API**
   ```bash
   curl "http://localhost:8080/v1/find-country?ip=90.91.92.93"
   ```

6. **Check metrics**
   ```bash
   curl "http://localhost:8080/metrics"
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

### Publishing to Docker Hub

1. **Set your Docker Hub username**
   ```bash
   export DOCKER_USERNAME=your-dockerhub-username
   ```

2. **Build and push to Docker Hub**
   ```bash
   make docker-build-push
   ```

   Or build and push separately:
   ```bash
   make docker-build
   make docker-push
   ```

3. **Pull and run from Docker Hub**
   ```bash
   docker pull your-dockerhub-username/torq:latest
   docker run -p 8080:8080 your-dockerhub-username/torq:latest
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
  "city": "Paris",
  "ip": "90.91.92.93"
}
```

**Error Response:**
```json
{
  "error": "IP not found"
}
```

### Health Check Endpoints

#### Liveness Probe

**Endpoint:** `GET /health/live`

**Example Request:**
```bash
curl "http://localhost:8080/health/live"
```

**Example Response:**
```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "torq"
}
```

#### Readiness Probe

**Endpoint:** `GET /health/ready`

**Example Request:**
```bash
curl "http://localhost:8080/health/ready"
```

**Example Response:**
```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "torq"
}
```

### Metrics Endpoint

**Endpoint:** `GET /metrics`

**Description:** Exposes Prometheus metrics for monitoring and observability.

**Example Request:**
```bash
curl "http://localhost:8080/metrics"
```

**Available Metrics:**
- `http_request_duration_seconds` - Request duration histogram
- `http_requests_total` - Total request count
- `http_error_requests_total` - Error request count (4xx, 5xx)
- `http_response_status_total` - Response status code count
- `http_requests_in_flight` - Currently active requests

## Configuration

### Database Configuration

The service supports flexible database configuration using JSON:

#### JSON Configuration (Recommended)

Set the `DB_CONFIG` environment variable:

```bash
export DB_CONFIG='{"dbtype": "csv", "extra_details": {"file_path": "/path/to/data.csv"}}'
```

#### Supported Database Types

The application uses type-safe enums for database types:

- `"csv"` - CSV file provider

#### Configuration Examples

**CSV Provider:**
```json
{
  "dbtype": "csv",
  "extra_details": {
    "file_path": "/path/to/ip_data.csv"
  }
}
```

#### CSV File Format

The CSV file should have the following format:
```csv
IP,CITY,COUNTRY
1.2.3.4,New York,USA
5.6.7.8,London,UK
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_CONFIG` | JSON configuration for database provider | - |
| `IP_DB_PROVIDER` | Legacy: IP database provider type | - |
| `IP_DB_PATH` | Legacy: Path to database file | - |
| `PORT` | Server port | `8080` |
| `RPS_LIMIT` | Rate limit (requests per second) | `10` |

## Project Structure

```
Torq/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── finder/                 # IP lookup business logic
│   ├── lookup/                 # Database providers
│   ├── router/                 # HTTP routing and middleware
│   ├── service_health/         # Health check handlers
│   └── limiter/                # Rate limiting
├── TestFiles/                  # Test data files
├── go.mod                      # Go module file
├── Makefile                    # Build automation
├── Dockerfile                  # Docker configuration
└── README.md                   # This file
```

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

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/lookup -v
```

## Observability

### Metrics

The service exposes comprehensive HTTP metrics via Prometheus:

- **Request Duration**: Histogram of request processing times
- **Request Count**: Total number of requests by method/path/status
- **Error Rate**: Count of error responses (4xx, 5xx)
- **Active Requests**: Currently in-flight requests

### Logging

The service uses structured logging with Zap:
- JSON format for production
- Request/response logging
- Error tracking with context

### Health Checks

- **Liveness**: Service is running
- **Readiness**: Service is ready to handle requests

## Error Handling

The factory provides descriptive errors for:
- Invalid JSON configuration
- Missing required fields
- Unsupported database types
- File not found errors
- Rate limit exceeded

## Type Safety

The application validates database types at runtime:
- Only supported types are accepted
- Prevents configuration typos
- Ensures only valid configurations are processed

## GitHub Actions and Docker Hub Integration

### Setting up Docker Hub Credentials in GitHub

To enable automatic Docker image publishing via GitHub Actions, you need to add your Docker Hub credentials as GitHub secrets:

1. **Create a Docker Hub Access Token**
   - Go to [Docker Hub](https://hub.docker.com/settings/security)
   - Click "New Access Token"
   - Give it a name (e.g., "GitHub Actions")
   - Copy the generated token

2. **Add Secrets to GitHub Repository**
   - Go to your GitHub repository
   - Navigate to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Add these secrets:
     - `DOCKERHUB_USERNAME`: Your Docker Hub username
     - `DOCKERHUB_TOKEN`: The access token you created

3. **Automatic Publishing**
   - The workflow will automatically build and push Docker images on:
     - Push to main/master branch
     - Tagged releases (v*)
   - Images will be tagged with:
     - Branch name (e.g., `main`)
     - Git SHA
     - Version tags (e.g., `v1.0.0`)

### Manual Docker Hub Publishing

For manual publishing without GitHub Actions:

```bash
# Login to Docker Hub
docker login

# Build and push
make docker-build-push DOCKER_USERNAME=your-username
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.