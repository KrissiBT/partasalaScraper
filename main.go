package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Count   int         `json:"count,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type SearchResponse struct {
	Success bool        `json:"success"`
	Query   string      `json:"query"`
	Count   int         `json:"count"`
	Data    interface{} `json:"data"`
}

type BrandResponse struct {
	Success bool        `json:"success"`
	Brand   string      `json:"brand"`
	Count   int         `json:"count"`
	Data    interface{} `json:"data"`
}

type CarResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

var scraper *PartasalaScraper

func main() {
	scraper = NewPartasalaScraper()

	r := mux.NewRouter()

	// Enable CORS middleware
	r.Use(corsMiddleware)

	// Routes
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/brands", getBrandsHandler).Methods("GET")
	r.HandleFunc("/brands/{brand_slug}", getBrandCarsHandler).Methods("GET")
	r.HandleFunc("/cars", getAllCarsHandler).Methods("GET")
	r.HandleFunc("/cars/{car_slug}", getCarDetailsHandler).Methods("GET")
	r.HandleFunc("/search", searchCarsHandler).Methods("GET")

	log.Println("Starting Partasala.is Scraper API...")
	log.Println("API Documentation: http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	doc := map[string]interface{}{
		"name":    "Partasala.is Scraper API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"/brands": map[string]interface{}{
				"method":      "GET",
				"description": "Get list of all car brands",
				"response":    "Array of brand objects with name and URL",
			},
			"/brands/<brand_slug>": map[string]interface{}{
				"method":      "GET",
				"description": "Get list of cars for a specific brand",
				"parameters": map[string]string{
					"brand_slug": "Brand identifier (e.g., audi, bmw, toyota)",
				},
				"response": "Array of car objects with name, URL, and thumbnail",
			},
			"/cars": map[string]interface{}{
				"method":      "GET",
				"description": "Get all available cars across all brands",
				"response":    "Array of all car objects with name, URL, and thumbnail",
			},
			"/cars/<car_slug>": map[string]interface{}{
				"method":      "GET",
				"description": "Get details and images for a specific car",
				"parameters": map[string]string{
					"car_slug": "Car identifier from the car URL",
				},
				"response": "Car object with name, description, and array of image URLs",
			},
			"/search": map[string]interface{}{
				"method":      "GET",
				"description": "Search for cars by name",
				"parameters": map[string]string{
					"q": "Search query",
				},
				"response": "Array of matching cars",
			},
		},
	}

	json.NewEncoder(w).Encode(doc)
}

func getBrandsHandler(w http.ResponseWriter, r *http.Request) {
	brands, err := scraper.GetBrands()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Count:   len(brands),
		Data:    brands,
	})
}

func getBrandCarsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	brandSlug := vars["brand_slug"]

	cars, err := scraper.GetBrandCars(brandSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(BrandResponse{
		Success: true,
		Brand:   brandSlug,
		Count:   len(cars),
		Data:    cars,
	})
}

func getAllCarsHandler(w http.ResponseWriter, r *http.Request) {
	cars, err := scraper.GetAllCars()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Count:   len(cars),
		Data:    cars,
	})
}

func getCarDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carSlug := vars["car_slug"]

	carDetails, err := scraper.GetCarDetails(carSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(CarResponse{
		Success: true,
		Data:    carDetails,
	})
}

func searchCarsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Missing search query parameter \"q\"",
		})
		return
	}

	results, err := scraper.SearchCars(strings.ToLower(query))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(SearchResponse{
		Success: true,
		Query:   query,
		Count:   len(results),
		Data:    results,
	})
}
