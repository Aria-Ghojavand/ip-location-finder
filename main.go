package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ip_geolocation_requests_total",
			Help: "Total number of IP geolocation requests",
		},
		[]string{"country", "source"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ip_geolocation_request_duration_seconds",
			Help: "Duration of IP geolocation requests",
		},
		[]string{"source"},
	)

	cacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ip_geolocation_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ip_geolocation_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

type IPInfo struct {
	IP       string    `json:"ip" db:"ip"`
	Country  string    `json:"country" db:"country"`
	CachedAt time.Time `json:"cached_at" db:"cached_at"`
}

type IPStackResponse struct {
	CountryName string `json:"country_name"`
}

type GeolocationAPI struct {
	db     *sql.DB
	apiKey string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	api := &GeolocationAPI{
		db:     db,
		apiKey: os.Getenv("IPSTACK_API_KEY"),
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/geolocate/:ip", api.geolocateIP)

		v1.POST("/geolocate/bulk", api.geolocateBulkIPs)

		v1.GET("/cached", api.getCachedIPs)

		v1.DELETE("/cache/:ip", api.clearCacheIP)

		v1.DELETE("/cache", api.clearAllCache)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func initDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbName == "" {
		dbName = "geolocation"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS ip_locations (
		id SERIAL PRIMARY KEY,
		ip INET UNIQUE NOT NULL,
		country VARCHAR(100) NOT NULL,
		cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_ip_locations_ip ON ip_locations(ip);
	`

	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, err
	}

	return db, nil
}

func (api *GeolocationAPI) geolocateIP(c *gin.Context) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("api"))
	defer timer.ObserveDuration()

	ipStr := c.Param("ip")

	ip := net.ParseIP(ipStr)
	if ip == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IP address"})
		return
	}

	if info, found := api.getFromCache(ipStr); found {
		cacheHits.Inc()
		requestsTotal.WithLabelValues(info.Country, "cache").Inc()
		c.JSON(http.StatusOK, info)
		return
	}

	cacheMisses.Inc()

	country, err := api.getCountryFromAPI(ipStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get location"})
		return
	}

	info := &IPInfo{
		IP:       ipStr,
		Country:  country,
		CachedAt: time.Now(),
	}

	if err := api.saveToCache(info); err != nil {
		log.Printf("Failed to save to cache: %v", err)
	}

	requestsTotal.WithLabelValues(country, "external_api").Inc()
	c.JSON(http.StatusOK, info)
}

func (api *GeolocationAPI) getFromCache(ip string) (*IPInfo, bool) {
	query := `SELECT ip, country, cached_at FROM ip_locations WHERE ip = $1`

	var info IPInfo
	err := api.db.QueryRow(query, ip).Scan(&info.IP, &info.Country, &info.CachedAt)
	if err != nil {
		return nil, false
	}

	if time.Since(info.CachedAt) > 24*time.Hour {
		return nil, false
	}

	return &info, true
}

func (api *GeolocationAPI) saveToCache(info *IPInfo) error {
	query := `INSERT INTO ip_locations (ip, country, cached_at) 
			  VALUES ($1, $2, $3) 
			  ON CONFLICT (ip) 
			  DO UPDATE SET country = $2, cached_at = $3`

	_, err := api.db.Exec(query, info.IP, info.Country, info.CachedAt)
	return err
}

func (api *GeolocationAPI) getCountryFromAPI(ip string) (string, error) {
	if api.apiKey == "" {
		return api.getFreeGeoLocation(ip)
	}

	url := fmt.Sprintf("http://api.ipstack.com/%s?access_key=%s", ip, api.apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result IPStackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.CountryName == "" {
		return "Unknown", nil
	}

	return result.CountryName, nil
}

func (api *GeolocationAPI) getFreeGeoLocation(ip string) (string, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=country", ip)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	country, ok := result["country"].(string)
	if !ok || country == "" {
		return "Unknown", nil
	}

	return country, nil
}

func (api *GeolocationAPI) geolocateBulkIPs(c *gin.Context) {
	var request struct {
		IPs []string `json:"ips" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(request.IPs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 100 IPs allowed per request"})
		return
	}

	results := make([]IPInfo, 0, len(request.IPs))

	for _, ipStr := range request.IPs {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			results = append(results, IPInfo{
				IP:      ipStr,
				Country: "Invalid IP",
			})
			continue
		}

		if info, found := api.getFromCache(ipStr); found {
			cacheHits.Inc()
			requestsTotal.WithLabelValues(info.Country, "cache").Inc()
			results = append(results, *info)
			continue
		}

		cacheMisses.Inc()

		country, err := api.getCountryFromAPI(ipStr)
		if err != nil {
			results = append(results, IPInfo{
				IP:      ipStr,
				Country: "Error",
			})
			continue
		}

		info := IPInfo{
			IP:       ipStr,
			Country:  country,
			CachedAt: time.Now(),
		}

		if err := api.saveToCache(&info); err != nil {
			log.Printf("Failed to save to cache: %v", err)
		}

		requestsTotal.WithLabelValues(country, "external_api").Inc()
		results = append(results, info)
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (api *GeolocationAPI) getCachedIPs(c *gin.Context) {
	query := `SELECT ip, country, cached_at FROM ip_locations ORDER BY cached_at DESC LIMIT 1000`

	rows, err := api.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cached data"})
		return
	}
	defer rows.Close()

	var results []IPInfo
	for rows.Next() {
		var info IPInfo
		if err := rows.Scan(&info.IP, &info.Country, &info.CachedAt); err != nil {
			continue
		}
		results = append(results, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"cached_ips": results,
		"count":      len(results),
	})
}

func (api *GeolocationAPI) clearCacheIP(c *gin.Context) {
	ipStr := c.Param("ip")

	ip := net.ParseIP(ipStr)
	if ip == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IP address"})
		return
	}

	query := `DELETE FROM ip_locations WHERE ip = $1`
	result, err := api.db.Exec(query, ipStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cache"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP not found in cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared for IP", "ip": ipStr})
}

func (api *GeolocationAPI) clearAllCache(c *gin.Context) {
	query := `DELETE FROM ip_locations`
	result, err := api.db.Exec(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear all cache"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message":       "All cache cleared",
		"rows_affected": rowsAffected,
	})
}
