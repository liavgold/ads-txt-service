# ADS-TXT-SERVICE

## Overview

This project is a RESTful API service built in Go that fetches and parses ads.txt files from a given domain, returning a JSON response with all advertiser domains listed in the file and the number of times each appears. The service includes caching, custom rate limiting, and comprehensive testing to ensure production readiness.



## Features

RESTful API: Provides an endpoint to retrieve advertiser domains from a given domain's ads.txt file.

Health Check: Includes a /health endpoint to monitor service status.

Caching Mechanism: A configurable caching system that supports Redis and allows easy switching between multiple cache types (in-memory, Redis, file-based) through environment variables.



Custom Rate Limiting: Implements a configurable token bucket rate limiter without relying on external libraries, allowing control of request limits per second.


Production Ready: Includes proper HTTP status codes, error handling, logging, and graceful shutdown.



Testing: Comprehensive unit and integration tests covering edge cases.
## Getting Started

Explain how to set up and run your project locally.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/liavgold/ads-txt-service.git
    ```

2.  **Install dependencies (if any):**
    ```bash
    go mod download
    ```
3.  **Run the application:**
    ```bash
    go run main.go
    ```
4.  **Create a .env file and copy .env.template to it**

## API

### GET /api/ads?domain=msn.com


Fetches and parses the ads.txt file for the specified domain.

Request:

GET /api/ads/msn.com

Response (200 OK):
```json
{
  "domain": "msn.com",
  "total_advertisers": 189,
  "advertisers": [
    {
      "domain": "google.com",
      "count": 102
    },
    {
      "domain": "appnexus.com",
      "count": 60
    },
    {
      "domain": "rubiconproject.com",
      "count": 27
    }
  ],
  "cached": false,
  "timestamp": "2025-07-13T10:30:45Z"
}
```

Error Responses:

400 Bad Request: Invalid domain format.

429 Too Many Requests: Rate limit exceeded.

# Docker Setup
 ```bash
    docker-compose up --build
```

# CI/CD

The repository includes a GitHub Actions workflow (.github/workflows/ci.yml) that:

Runs tests on push/pull requests.

Builds and pushes the Docker image to a container registry (configure registry details in the workflow).