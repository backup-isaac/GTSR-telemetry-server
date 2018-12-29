package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"server/configs"
	"server/storage"

	"github.com/gorilla/mux"
)

// API is the object which handles HTTP API requests
type API struct {
	store storage.Storage
}

// NewAPI returns a new API initialized with the provided store
func NewAPI(store storage.Storage) *API {
	return &API{store: store}
}

// Default handles the default API query
func (api *API) Default(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Welcome to the Solar Racing Telemetry API!"))
}

// Metrics returns a list of the metrics tracked by InfluxDB
// Keep in mind this is just a list of names; it doesn't
// have the same functionality as on the old server
func (api *API) Metrics(res http.ResponseWriter, req *http.Request) {
	metrics, err := api.store.ListMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(metrics)
}

// LastActive returns the timestamp of the last seen datapoint in the store
func (api *API) LastActive(res http.ResponseWriter, req *http.Request) {
	metrics, err := api.store.ListMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	var maxTime time.Time
	for _, metric := range metrics {
		point, err := api.store.Latest(metric)
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
func (api *API) Configs(res http.ResponseWriter, req *http.Request) {
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
func (api *API) Latest(res http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if strings.Contains(name, ";") {
		http.Error(res, "Invalid metric name", 400)
		return
	}
	lastPoint, err := api.store.Latest(name)
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
func (api *API) Location(res http.ResponseWriter, req *http.Request) {
	latpt, err := api.store.Latest("GPS_Latitude")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	lngpt, err := api.store.Latest("GPS_Longitude")
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

// StartServer starts the HTTP server
func (api *API) StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/api", api.Default).Methods("GET")
	router.HandleFunc("/api/metrics", api.Metrics).Methods("GET")
	router.HandleFunc("/api/lastActive", api.LastActive).Methods("GET")
	router.HandleFunc("/api/configs", api.Configs).Methods("GET")
	router.HandleFunc("/api/latest", api.Latest).Methods("GET")
	router.HandleFunc("/api/location", api.Location).Methods("GET")

	api.RegisterCsvRoutes(router)
	api.RegisterMapRoutes(router)
	api.RegisterDataRoutes(router)
	api.RegisterJacksonRoutes(router)
	api.RegisterChatRoutes(router)
	fmt.Println("Starting HTTP server...")
	log.Fatal(http.ListenAndServe(":8888", router))
}
