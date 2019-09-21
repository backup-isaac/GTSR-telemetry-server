package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/configs"
	"server/storage"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Core is the object which handles HTTP Core requests
type Core struct {
	store *storage.Storage
}

// NewCore returns a new API initialized with the provided store
func NewCore(store *storage.Storage) *Core {
	return &Core{store: store}
}

// Default handles the default API query
func (c *Core) Default(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Welcome to the Solar Racing Telemetry API!"))
}

// Metrics returns a list of the metrics tracked by InfluxDB
// Keep in mind this is just a list of names; it doesn't
// have the same functionality as on the old server
func (c *Core) Metrics(res http.ResponseWriter, req *http.Request) {
	metrics, err := c.store.ListMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(metrics)
}

// LastActive returns the timestamp of the last seen datapoint in the store
func (c *Core) LastActive(res http.ResponseWriter, req *http.Request) {
	metrics, err := c.store.ListMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	var maxTime time.Time
	for _, metric := range metrics {
		point, err := c.store.Latest(metric)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if point.Time.After(maxTime) {
			maxTime = point.Time
		}
	}
	res.Write([]byte(maxTime.String()))
}

// Configs returns a list of the CAN configurations
// This is closer to the functionality of the old server's metrics query
func (c *Core) Configs(res http.ResponseWriter, req *http.Request) {
	canConfigMap, err := configs.LoadConfigs()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	var canConfigList []*configs.CanConfigType
	for _, configs := range canConfigMap {
		canConfigList = append(canConfigList, configs...)
	}
	sort.Slice(canConfigList, func(i, j int) bool {
		return canConfigList[i].CanID < canConfigList[j].CanID
	})
	encoder := json.NewEncoder(res)
	encoder.SetIndent("", "  ")
	encoder.Encode(canConfigList)
}

// Latest returns the last known value of the metric specified by name
func (c *Core) Latest(res http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if strings.Contains(name, ";") {
		http.Error(res, "Invalid metric name", 400)
		return
	}
	lastPoint, err := c.store.Latest(name)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if lastPoint == nil {
		http.Error(res, fmt.Sprintf("No data found for metric %s", name), http.StatusBadRequest)
		return
	}
	json.NewEncoder(res).Encode(lastPoint.Value)
}

// Location returns the current position of the car
func (c *Core) Location(res http.ResponseWriter, req *http.Request) {
	latpt, err := c.store.LatestNonZero("GPS_Latitude")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	lngpt, err := c.store.LatestNonZero("GPS_Longitude")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if latpt == nil || lngpt == nil {
		http.Error(res, "Location metrics (GPS_Latitude/GPS_Longitude) not found", http.StatusInternalServerError)
		return
	}
	location := map[string]float64{
		"lat": latpt.Value,
		"lng": lngpt.Value,
	}
	json.NewEncoder(res).Encode(location)
}

// RegisterRoutes registers the routes handled by the API core
func (c *Core) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api", c.Default).Methods("GET")
	router.HandleFunc("/api/metrics", c.Metrics).Methods("GET")
	router.HandleFunc("/api/lastActive", c.LastActive).Methods("GET")
	router.HandleFunc("/api/configs", c.Configs).Methods("GET")
	router.HandleFunc("/api/latest", c.Latest).Methods("GET")
	router.HandleFunc("/api/location", c.Location).Methods("GET")
}
