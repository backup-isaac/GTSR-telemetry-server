package api

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"
)

// ReconToolHandler handles requests related to ReconTool
type ReconToolHandler struct{}

// NewReconToolHandler returns an initialized ReconToolHandler
func NewReconToolHandler() *ReconToolHandler {
	return &ReconToolHandler{}
}

// ReconToolDefault is the default handler for /reconTool
func (r *ReconToolHandler) ReconToolDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/reconTool/static/index.html", http.StatusFound)
}

// ReconTimeRange runs ReconTool on data taken from the server
func (r *ReconToolHandler) ReconTimeRange(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
		return
	}
	startDateString := req.Form.Get("startDate")
	if startDateString == "" {
		http.Error(res, "Missing start date", http.StatusBadRequest)
		return
	}
	endDateString := req.Form.Get("endDate")
	if endDateString == "" {
		http.Error(res, "Missing end date", http.StatusBadRequest)
		return
	}
	resolutionString := req.Form.Get("resolution")
	if resolutionString == "" {
		http.Error(res, "Missing resolution", http.StatusBadRequest)
		return
	}
	startDate, err := unixStringMillisToTime(startDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing start date: %s", err), http.StatusBadRequest)
		return
	}
	endDate, err := unixStringMillisToTime(endDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing end date: %s", err), http.StatusBadRequest)
		return
	}
	resolution64, err := strconv.ParseInt(resolutionString, 10, 32)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing resolution: %s", err), http.StatusBadRequest)
		return
	}
	resolution := int(resolution64)
	if resolution <= 0 {
		http.Error(res, "Resolution must be positive", http.StatusBadRequest)
		return
	}
	res.Write([]byte(fmt.Sprintf("Request successful: start date %d, end date %d, resolution %d", startDate, endDate, resolution)))
}

// ReconCSV runs ReconTool on data provided as a CSV
func (r *ReconToolHandler) ReconCSV(res http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(1048576)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing multipart form: %s", err), http.StatusBadRequest)
		return
	}
	numCsvs := len(req.MultipartForm.File)
	if numCsvs > 0 {
		res.Write([]byte(fmt.Sprintf("Request successful: %d CSVs present", numCsvs)))
	} else {
		http.Error(res, fmt.Sprintf("No CSVs present"), http.StatusBadRequest)
	}
}

// RegisterRoutes registers the routes for the ReconTool service
func (r *ReconToolHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/reconTool/static/").Handler(http.StripPrefix("/reconTool/static/", http.FileServer(http.Dir(path.Join(dir, "reconTool")))))

	router.HandleFunc("/reconTool", r.ReconToolDefault).Methods("GET")
	router.HandleFunc("/reconTool/timeRange", r.ReconTimeRange).Methods("POST")
	router.HandleFunc("/reconTool/fromCSV", r.ReconCSV).Methods("POST")
}
