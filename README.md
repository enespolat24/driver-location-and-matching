# Driver Location and Matching Service

This case study focuses on building a system that matches a rider with the nearest available driver. It consists of two services: the Driver Location API, which manages driver location data stored in MongoDB, and the Matching API, which uses this data to find the closest driver based on a given point. The goal is to provide accurate and efficient driver-rider matching using geospatial queries.


This project contains two microservices:
- **Driver Location Service**: Manages driver locations
- **Matching Service**: Handles driver-rider matching logic

<br>
Assuming the JWT will come from another service because the responsibility of the matching service should only be to match riders and drivers. Additionally, according to the case requirements, there is no need for the matching service to be stateful, which is why the matching service does not have any database.

<br>

**i've used batch processing to import driver locations from Coordinate.csv**

Both APIs are built with clean, production-ready code and thorough error handling for reliability. They follow good architectural practices, using the hexagonal architecture to ensure separation of concerns and ease of testing.

API documentation is provided via OpenAPI, and unit/integration tests validate functionality. Additionally, a circuit breaker pattern is implemented to improve system resilience.

## 📊 Test Coverage

| Service | Coverage |
|---------|----------|
| **Driver Location Service** | 73.8%|
| **Matching Service** | 82.4%  |

## Project Diagram
![Project diagram](diagram.png)

# Quick Start

## 🚀 Get Started in 30 Seconds

```bash
# 1. Run tests
make test

# 2. Start everything (copies .env files + starts services)
make up

# 3. Stop when done
make down
```

## 📚 API Documentation

- **Driver Location Service**: http://localhost:8087/swagger/index.html
- **Matching Service**: http://localhost:8088/swagger/index.html

> 💡 **Tip**: Run `make swagger` to update API documentation after code changes


### Prerequisites
- Docker
- Docker Compose
- Go (for local development and testing)

### Starting Services

```bash
# Setup environment and start all services (one command!)
make up

# Check status
docker compose ps

# View logs
docker compose logs -f
```

### Default Service Endpoints
do not forget to double check these from .env file

- **Driver Location Service**: http://localhost:8087
- **Matching Service**: http://localhost:8088
- **MongoDB**: localhost:27017
- **Redis**: localhost:6379

### Stopping Services

```bash
# Stop all services
make down

# Stop and remove volumes (WARNING: All data will be deleted!)
docker compose down -v
```

### Individual Service Management

```bash
# Restart specific service
docker compose restart driver-location-service

# View specific service logs
docker compose logs -f matching-service

# Run only database services
docker compose up mongodb redis
```

## Development Commands

### Available Make Commands

```bash
# Run tests for both services
make test

# Update swagger documentation for both services  
make swagger

# Setup environment and start all services
make up

# Stop all services
make down
```

### Testing

```bash
# Run all tests
make test

# Run tests for specific service
cd the-driver-location-service && go test ./...
cd the-matching-service && go test ./...
```

### Documentation

```bash
# Update API documentation (requires swag tool)
make swagger

# Install swag tool if needed
go install github.com/swaggo/swag/cmd/swag@latest
```

##  API Usage

### Driver Matching

```bash
curl -X POST http://localhost:8088/api/v1/match \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "id": "rider123",
    "name": "John",
    "surname": "Doe",
    "location": {
      "type": "Point",
      "coordinates": [28.943153502720264, 41.02629698673695]
    },
    "radius": 500
  }'
```

### JWT Token Details

The matching service requires JWT authentication. You can use the following token for testing (generated with the secret from .env.example):

**Token**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRoZW50aWNhdGVkIjp0cnVlLCJ1c2VyX2lkIjoiODE1MzlhYzAtODQxOC00YzljLWE0MjYtM2ZiYmY5ZTJlZTZlIn0.WfaMox95-Gtna1GLfszAkpk6fHLKSvXu-Rxs3kEOGVA`

**Example Claim**:
```json
{
  "authenticated": true,
  "user_id": "81539ac0-8418-4c9c-a426-3fbbf9e2ee6e"
}
```

##  Driver Create Endpoint

> **Note:** 
> - To create a single driver, send a JSON array with one object.
> - To create multiple drivers, send a JSON array with multiple objects.

#### Single Driver Example
````
POST http://localhost:8087/api/v1/drivers
[
  {
    "location": {
      "type": "Point",
      "coordinates": [29.0, 41.0]
    }
  }
]
````

#### Batch Driver Example
````
POST http://localhost:8087/api/v1/drivers
[
  {
    "location": {
      "type": "Point",
      "coordinates": [29.0, 41.0]
    }
  },
  {
    "location": {
      "type": "Point",
      "coordinates": [30.0, 42.0]
    }
  }
]
````

---

## Monitoring & Dashboard

### Prometheus & Grafana

- **Prometheus** collects and monitors metrics from all services.
- **Grafana** automatically loads a ready-made dashboard and sets it as the home dashboard (default landing page).
- Dashboard file: `grafana/dashboards/echo-multi-service.json`
- Provisioning and auto-load configuration: `grafana/provisioning/`
- Anyone who clones the repo and runs `docker compose up` will see this dashboard as the home page upon logging into Grafana.

### Accessing Grafana
- URL: [http://localhost:3000](http://localhost:3000)
- After login, the "Echo Multi-Service HTTP Metrics" dashboard will automatically open as the home dashboard.

---
