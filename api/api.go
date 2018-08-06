package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.gatech.edu/GTSR/telemetry-server/storage"
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

// StartServer starts the HTTP server
func (api *API) StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/api", api.Default).Methods("GET")
	router.HandleFunc("/api/metrics", api.Metrics).Methods("GET")
	router.HandleFunc("/api/lastActive", api.LastActive).Methods("GET")

	fmt.Println("Starting HTTP server...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
