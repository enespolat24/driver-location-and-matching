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


## Project Diagram
![Project diagram](diagram.png)

# Quick Start

## 📚 API Documentation

- **Driver Location Service**: http://localhost:8087/swagger/index.html
- **Matching Service**: http://localhost:8088/swagger/index.html


### Prerequisites
- Docker
- Docker Compose

### Starting Services

```bash
# Copy environment file
cp .env.example .env

# Start all services
docker compose up -d

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
docker compose down

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

##  API Usage

### Driver Matching

```bash
curl -X POST http://localhost:3001/api/v1/match \
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
