package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Brand struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

type Car struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	URL       string  `json:"url"`
	Thumbnail *string `json:"thumbnail"`
	Brand     string  `json:"brand"`
	MatchType string  `json:"match_type,omitempty"`
}

type Image struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

type CarDetails struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	URL         string  `json:"url"`
	Brand       *string `json:"brand"`
	Description *string `json:"description"`
	ImageCount  int     `json:"image_count"`
	Images      []Image `json:"images"`
}

type PartasalaScraper struct {
	baseURL string
	client  *http.Client
}

func NewPartasalaScraper() *PartasalaScraper {
	return &PartasalaScraper{
		baseURL: "https://partasala.is",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *PartasalaScraper) getPage(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *PartasalaScraper) GetBrands() ([]Brand, error) {
	doc, err := s.getPage(s.baseURL)
	if err != nil {
		return nil, err
	}

	brands := []Brand{}
	seenBrands := make(map[string]bool)
	brandPattern := regexp.MustCompile(`/bilaflokkur/[^/]+/?$`)

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || !brandPattern.MatchString(href) {
			return
		}

		// Extract brand slug
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		brandSlug := parts[len(parts)-1]

		// Avoid duplicates
		if seenBrands[brandSlug] {
			return
		}
		seenBrands[brandSlug] = true

		brandName := strings.TrimSpace(sel.Text())

		brands = append(brands, Brand{
			Name: brandName,
			Slug: brandSlug,
			URL:  s.makeAbsoluteURL(href),
		})
	})

	// Sort by name
	sort.Slice(brands, func(i, j int) bool {
		return brands[i].Name < brands[j].Name
	})

	return brands, nil
}

func (s *PartasalaScraper) GetBrandCars(brandSlug string) ([]Car, error) {
	url := fmt.Sprintf("%s/bilaflokkur/%s/", s.baseURL, brandSlug)
	doc, err := s.getPage(url)
	if err != nil {
		return nil, err
	}

	cars := []Car{}
	seenCars := make(map[string]bool)
	carPattern := regexp.MustCompile(`/bilaskra/[^/]+/?$`)

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || !carPattern.MatchString(href) {
			return
		}

		// Extract car slug
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		carSlug := parts[len(parts)-1]

		// Avoid duplicates
		if seenCars[carSlug] {
			return
		}
		seenCars[carSlug] = true

		carName := strings.TrimSpace(sel.Text())

		// Try to find thumbnail image
		var thumbnail *string
		img := sel.Find("img")
		if imgSrc, exists := img.Attr("src"); exists {
			absoluteURL := s.makeAbsoluteURL(imgSrc)
			thumbnail = &absoluteURL
		}

		cars = append(cars, Car{
			Name:      carName,
			Slug:      carSlug,
			URL:       s.makeAbsoluteURL(href),
			Thumbnail: thumbnail,
			Brand:     brandSlug,
		})
	})

	return cars, nil
}

func (s *PartasalaScraper) GetCarDetails(carSlug string) (*CarDetails, error) {
	url := fmt.Sprintf("%s/bilaskra/%s/", s.baseURL, carSlug)
	doc, err := s.getPage(url)
	if err != nil {
		return nil, err
	}

	// Extract car name
	var carName string
	doc.Find("h1").Each(func(i int, sel *goquery.Selection) {
		if i == 0 {
			carName = strings.TrimSpace(sel.Text())
		}
	})

	// Extract description
	var description *string
	doc.Find("div").Each(func(i int, sel *goquery.Selection) {
		class, _ := sel.Attr("class")
		if strings.Contains(strings.ToLower(class), "description") ||
			strings.Contains(strings.ToLower(class), "content") ||
			strings.Contains(strings.ToLower(class), "lýsing") {
			desc := strings.TrimSpace(sel.Text())
			description = &desc
		}
	})

	// Extract brand/category
	var brand *string
	doc.Find("a[href*='/bilaflokkur/']").Each(func(i int, sel *goquery.Selection) {
		if i == 0 {
			brandName := strings.TrimSpace(sel.Text())
			brand = &brandName
		}
	})

	// Extract all images
	images := []Image{}
	seenImages := make(map[string]bool)
	sizePattern := regexp.MustCompile(`-\d+x\d+\.(jpg|jpeg|png|gif)`)

	// Look for img tags
	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		src, exists := sel.Attr("src")
		if !exists || !strings.Contains(src, "uploads") || strings.Contains(strings.ToLower(src), "logo") {
			return
		}

		// Get full-size image URL (remove size suffixes like -300x300)
		fullSrc := sizePattern.ReplaceAllString(src, ".$1")
		fullURL := s.makeAbsoluteURL(fullSrc)

		if seenImages[fullURL] {
			return
		}
		seenImages[fullURL] = true

		images = append(images, Image{
			URL:       fullURL,
			Thumbnail: s.makeAbsoluteURL(src),
		})
	})

	// Also look for links to images
	imagePattern := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || !imagePattern.MatchString(strings.ToLower(href)) {
			return
		}

		if !strings.Contains(href, "uploads") {
			return
		}

		fullURL := s.makeAbsoluteURL(href)

		if seenImages[fullURL] {
			return
		}
		seenImages[fullURL] = true

		images = append(images, Image{
			URL:       fullURL,
			Thumbnail: fullURL,
		})
	})

	return &CarDetails{
		Name:        carName,
		Slug:        carSlug,
		URL:         url,
		Brand:       brand,
		Description: description,
		ImageCount:  len(images),
		Images:      images,
	}, nil
}

func (s *PartasalaScraper) GetAllCars() ([]Car, error) {
	allCars := []Car{}

	// Get all brands
	brands, err := s.GetBrands()
	if err != nil {
		return nil, err
	}

	// Get cars from each brand
	for _, brand := range brands {
		cars, err := s.GetBrandCars(brand.Slug)
		if err != nil {
			// Continue even if one brand fails
			continue
		}
		allCars = append(allCars, cars...)
	}

	return allCars, nil
}

func (s *PartasalaScraper) SearchCars(query string) ([]Car, error) {
	queryLower := strings.ToLower(query)
	results := []Car{}

	// Get all brands first
	brands, err := s.GetBrands()
	if err != nil {
		return nil, err
	}

	// Search through each brand
	for _, brand := range brands {
		// Check if query matches brand name
		if strings.Contains(strings.ToLower(brand.Name), queryLower) {
			cars, err := s.GetBrandCars(brand.Slug)
			if err != nil {
				continue
			}
			for _, car := range cars {
				car.MatchType = "brand"
				results = append(results, car)
			}
		} else {
			// Search for cars within this brand
			cars, err := s.GetBrandCars(brand.Slug)
			if err != nil {
				continue
			}
			for _, car := range cars {
				if strings.Contains(strings.ToLower(car.Name), queryLower) {
					car.MatchType = "car_name"
					results = append(results, car)
				}
			}
		}
	}

	return results, nil
}

func (s *PartasalaScraper) makeAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return s.baseURL + href
	}
	return s.baseURL + "/" + href
}
