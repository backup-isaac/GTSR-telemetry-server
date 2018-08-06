package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"
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
	err := req.ParseForm()
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
		generating.Store(false)
		return
	}
	startDateString := req.Form.Get("startDate")
	endDateString := req.Form.Get("endDate")
	resolutionString := req.Form.Get("resolution")
	if startDateString == "" || endDateString == "" || resolutionString == "" {
		http.Error(res, "malformatted query", http.StatusBadRequest)
		fmt.Println("bad query")
		generating.Store(false)
		return
	}
	startDate, err := unixStringMillisToTime(startDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing start date: %s", err), http.StatusBadRequest)
		generating.Store(false)
		return
	}
	endDate, err := unixStringMillisToTime(endDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing end date: %s", err), http.StatusBadRequest)
		generating.Store(false)
		return
	}
	resolution64, err := strconv.ParseInt(resolutionString, 10, 32)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing resolution: %s", err), http.StatusBadRequest)
		generating.Store(false)
		return
	}
	resolution := int(resolution64)
	go api.generateCsv(startDate, endDate, resolution)
	res.WriteHeader(http.StatusOK)
}

func unixStringMillisToTime(timeString string) (time.Time, error) {
	timeMillis, err := strconv.ParseInt(timeString, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, timeMillis*1e6), nil
}

func (api *API) generateCsv(start time.Time, end time.Time, resolution int) {
	defer generating.Store(false)
	<-time.After(10 * time.Second)
}

// RegisterCsvRoutes registers the routes for the CSV service
func (api *API) RegisterCsvRoutes(router *mux.Router) {
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
