# Torq - IP Geolocation Service

A Go-based microservice that provides IP geolocation functionality through a REST API. The service can determine the country of an IP address using various backend providers with comprehensive metrics and observability.

## Features

- 🌍 IP to country geolocation
- 🔌 Pluggable backend providers with JSON configuration
- 📊 OpenTelemetry metrics and tracing
- 🚦 Rate limiting with configurable RPS (Requests Per Second)
- 🐳 Docker support
- 🧪 Comprehensive testing
- 📈 Prometheus metrics endpoint
- 🔒 Security scanning

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/shaibs3/Torq.git
   cd Torq
   ```
2. **Set up database configuration**
   
   **Option 1: Environment Variable**
   
   **For CSV:**
   ```bash
   export IP_DB_CONFIG='{"dbtype": "csv", "extra_details": {"file_path": "./TestFiles/ip_data.csv"}}'
   ```
   
   **For PostgreSQL:**
   ```bash
   export IP_DB_CONFIG='{"dbtype": "postgres", "extra_details": {"conn_str": "postgres://username:password@localhost:5432/ipdb?sslmode=disable"}}'
   ```
   
   **Option 2: .env File**
   ```bash
   # Copy the example environment file
   cp .env_example .env
   
   # Edit the .env file with your configuration
   # For CSV:
   IP_DB_CONFIG='{"dbtype": "csv", "extra_details": {"file_path": "./TestFiles/ip_data.csv"}}'
   
   # For PostgreSQL:
   # IP_DB_CONFIG='{"dbtype": "postgres", "extra_details": {"conn_str": "postgres://username:password@localhost:5432/ipdb?sslmode=disable"}}'
   
   RPS_LIMIT=10
   RPS_BURST=20
   PORT=8080
   ```

3. **Run the application**
   ```bash
   # Using Go directly
   go run cmd/main.go

   # Or using Make
   make run
   ```

4. **Test the API**
   ```bash
   curl "http://localhost:8080/v1/find-country?ip=90.91.92.93"
   ```

5. **Check metrics**
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


## Configuration

### Database Configuration

The service supports flexible database configuration using JSON:

#### JSON Configuration (Recommended)

Set the `IP_DB_CONFIG` environment variable:

```bash
export IP_DB_CONFIG='{"dbtype": "csv", "extra_details": {"file_path": "/path/to/data.csv"}}'
```

#### Supported Database Types

The application uses type-safe enums for database types:

- `"csv"` - CSV file provider
- `"postgres"` - PostgreSQL database provider

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

**PostgreSQL Provider:**
```json
{
  "dbtype": "postgres",
  "extra_details": {
    "conn_str": "postgres://username:password@localhost:5432/ipdb?sslmode=disable"
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

#### PostgreSQL Table Structure

The PostgreSQL database should have a table with the following structure:
```sql
CREATE TABLE ip_locations (
    ip VARCHAR(45) PRIMARY KEY,
    city VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL
);
```

### Environment Variables

| Variable       | Description                                 | Default      |
|----------------|---------------------------------------------|--------------|
| `IP_DB_CONFIG` | JSON configuration for database provider    | -            |
| `PORT`         | Server port                                 | `8080`       |
| `RPS_LIMIT`    | Rate limit (requests per second)            | `10`         |
| `RPS_BURST`    | Number of burst requests allowed per second | `10`         |
| `LOG_LEVEL`    | Log level                                   | `info`       |
| `ENVIRONMENT`  | ENVIRONMENT                                 | `production` |




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
- **Rate limited Requests**: Number of rate limited requests






