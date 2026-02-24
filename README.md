# Partasala.is Scraper API

A Go-based REST API that interfaces with the partasala.is website to retrieve information about car brands, available cars, and their images.

## 🌐 Live Server

**Live API:** [http://kristofer.is:1667/](http://kristofer.is:1667/)

The API is currently running and available for public use!

## Features

- 🚗 Get list of all car brands
- 📋 Get cars by brand
- 🖼️ Get all images for specific cars
- 🔍 Search functionality across all cars
- 🌐 RESTful JSON API
- ⚡ Fast web scraping with goquery
- 🚀 High performance and concurrent scraping

## Installation

1. Install dependencies:
```bash
go mod download
```

2. Run the API server:
```bash
go run .
```

Or build and run:
```bash
go build -o partasala-api
./partasala-api
```

The API will be available at `http://localhost:8080`

## API Endpoints

### GET `/`
Returns API documentation and available endpoints.

**Example:**
```bash
curl http://localhost:8080/
```

### GET `/brands`
Get list of all car brands available on partasala.is.

**Response:**
```json
{
  "success": true,
  "count": 25,
  "data": [
    {
      "name": "Audi",
      "slug": "audi",
      "url": "https://partasala.is/bilaflokkur/audi/"
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/brands
```

### GET `/brands/<brand_slug>`
Get all cars for a specific brand.

**Parameters:**
- `brand_slug`: Brand identifier (e.g., `audi`, `bmw`, `toyota`)

**Response:**
```json
{
  "success": true,
  "brand": "audi",
  "count": 1,
  "data": [
    {
      "name": "AUDI A3 - SPORTBACK E-TRON",
      "slug": "audi-a3-sportback-e-tron",
      "url": "https://partasala.is/bilaskra/audi-a3-sportback-e-tron/",
      "thumbnail": "https://partasala.is/...",
      "brand": "audi"
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/brands/audi
```

### GET `/cars/<car_slug>`
Get detailed information and all images for a specific car.

**Parameters:**
- `car_slug`: Car identifier (e.g., `audi-a3-sportback-e-tron`)

**Response:**
```json
{
  "success": true,
  "data": {
    "name": "AUDI A3 – SPORTBACK E-TRON",
    "slug": "audi-a3-sportback-e-tron",
    "url": "https://partasala.is/bilaskra/audi-a3-sportback-e-tron/",
    "brand": "Audi",
    "description": "1400cc Bensin/Rafmagn ssk",
    "image_count": 5,
    "images": [
      {
        "url": "https://partasala.is/wp-content/uploads/2024/03/20240306_095934-scaled.jpg",
        "thumbnail": "https://partasala.is/wp-content/uploads/2024/03/20240306_095934-scaled-300x300.jpg"
      }
    ]
  }
}
```

**Example:**
```bash
curl http://localhost:8080/cars/audi-a3-sportback-e-tron
```

### GET `/search?q=<query>`
Search for cars by name across all brands.

**Parameters:**
- `q`: Search query (e.g., `audi`, `sportback`, `toyota`)

**Response:**
```json
{
  "success": true,
  "query": "audi",
  "count": 1,
  "data": [
    {
      "name": "AUDI A3 - SPORTBACK E-TRON",
      "slug": "audi-a3-sportback-e-tron",
      "url": "https://partasala.is/bilaskra/audi-a3-sportback-e-tron/",
      "thumbnail": "https://partasala.is/...",
      "brand": "audi",
      "match_type": "brand"
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/search?q=audi"
```

## Usage Examples

### Go
```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Get all brands
    resp, _ := http.Get("http://localhost:8080/brands")
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result)
}
```

### Python
```python
import requests

# Get all brands
response = requests.get('http://localhost:8080/brands')
brands = response.json()

# Get cars for a brand
response = requests.get('http://localhost:8080/brands/toyota')
cars = response.json()

# Get car details
response = requests.get('http://localhost:8080/cars/audi-a3-sportback-e-tron')
car_details = response.json()

# Search
response = requests.get('http://localhost:8080/search?q=audi')
results = response.json()
```

### JavaScript
```javascript
// Get all brands
fetch('http://localhost:8080/brands')
  .then(res => res.json())
  .then(data => console.log(data));

// Get cars for a brand
fetch('http://localhost:8080/brands/bmw')
  .tmain.go**: Main HTTP server with API routes using Gorilla Mux
- **scraper.go**: Web scraping logic using goquery
- **go.mod**: Go module

### curl
```bash
# Get all brands
curl http://localhost:8080/brands | jq

# Get cars for specific brand
curl http://localhost:8080/brands/toyota | jq

# Get car details
curl http://localhost:8080/cars/audi-a3-sportback-e-tron | jq

# Search
curl "http://localhost:8080/search?q=bmw" | jq
```

## Architecture

- **app.py**: Main Bottle application with API routes
- **scraper.py**: Web scraping logic using BeautifulSoup
- **requirements.txt**: Python dependencies

## Error Handling

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": "Error message here"
}
```

HTTP status codes:
- `200`: Success
- `400`: Bad request (missing parameters)
- `500`: Server error (scraping failed)

## Notes

- The API scrapes data in real-time from partasala.is
- Response times depend on the website's availability
- Consider implementing caching for production use
- Respect the website's robots.txt and terms of service

## License

MIT
