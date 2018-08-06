package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

var generating atomic.Value

func init() {
	generating.Store(false)
}

// CsvDefault is the default handler for the /csv route
func (api *API) CsvDefault(res http.ResponseWriter, req *http.Request) {
	if !generating.Load().(bool) {
		http.Redirect(res, req, "/csv/static/index.html", http.StatusFound)
	} else {
		http.Redirect(res, req, "/csv/static/generating.html", http.StatusFound)
	}
}

// IsGenerating returns whether a CSV is currently being generated
func (api *API) IsGenerating(res http.ResponseWriter, req *http.Request) {
	json.NewEncoder(res).Encode(generating.Load().(bool))
}

// GenerateCsv generates the csv
func (api *API) GenerateCsv(res http.ResponseWriter, req *http.Request) {
	if generating.Load().(bool) {
		http.Error(res, "Already generating CSV", http.StatusLocked)
		return
	}
	generating.Store(true)
	// TODO: implement CSV generation
	http.Error(res, "Not implemented", http.StatusNotFound)
	go func() {
		<-time.After(10 * time.Second)
		generating.Store(false)
	}()
}

// RegisterCSVRoutes registers the routes for the CSV service
func (api *API) RegisterCSVRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/csv/static/").Handler(http.StripPrefix("/csv/static/", http.FileServer(http.Dir(path.Join(dir, "csv")))))

	router.HandleFunc("/csv", api.CsvDefault).Methods("GET")
	router.HandleFunc("/csv/isGenerating", api.IsGenerating).Methods("GET")
	router.HandleFunc("/csv/generateCsv", api.GenerateCsv).Methods("POST")
}
