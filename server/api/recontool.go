package api

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"server/recontool"
	"server/storage"
	"strconv"

	"github.com/gorilla/mux"
)

// ReconToolHandler handles requests related to ReconTool
type ReconToolHandler struct {
	store *storage.Storage
}

// NewReconToolHandler returns an initialized ReconToolHandler
func NewReconToolHandler(store *storage.Storage) *ReconToolHandler {
	return &ReconToolHandler{store: store}
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
	vehicle, err := extractVehicleForm(&req.Form)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := r.store.GetMetricPointsRange(recontool.MetricNames, startDate, endDate, resolution, true)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	res.Write([]byte(fmt.Sprintf("Request successful: vehicle %v, metrics %v", vehicle, data)))
}

func extractVehicleForm(form *url.Values) (recontool.Vehicle, error) {
	vehicleParamsFloat := []string{
		"Rmot",
		"m",
		"Crr1",
		"Crr2",
		"CDa",
		"Tmax",
		"Qmax",
		"Rline",
		"VcMax",
		"VcMin",
	}
	paramValuesFloat := make([]float64, len(vehicleParamsFloat))
	for i, v := range vehicleParamsFloat {
		valueString := form.Get(v)
		if valueString == "" {
			return recontool.Vehicle{}, fmt.Errorf("missing parameter %s", v)
		}
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return recontool.Vehicle{}, err
		}
		paramValuesFloat[i] = value
	}
	vSerString := form.Get("Vser")
	if vSerString == "" {
		return recontool.Vehicle{}, fmt.Errorf("missing parameter Vser")
	}
	vSer, err := strconv.ParseInt(vSerString, 10, 32)
	if err != nil {
		return recontool.Vehicle{}, err
	}
	if vSer < 0 {
		return recontool.Vehicle{}, fmt.Errorf("Invalid uint literal %s", vSerString)
	}
	return recontool.Vehicle{
		RMot:  paramValuesFloat[0],
		M:     paramValuesFloat[1],
		Crr1:  paramValuesFloat[2],
		Crr2:  paramValuesFloat[3],
		CDa:   paramValuesFloat[4],
		TMax:  paramValuesFloat[5],
		QMax:  paramValuesFloat[6],
		RLine: paramValuesFloat[7],
		VcMax: paramValuesFloat[8],
		VcMin: paramValuesFloat[9],
		VSer:  uint(vSer),
	}, nil
}

// ReconCSV runs ReconTool on data provided as a CSV
func (r *ReconToolHandler) ReconCSV(res http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(1048576)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing multipart form: %s", err), http.StatusBadRequest)
		return
	}
	numCsvs := len(req.MultipartForm.File)
	if numCsvs == 0 {
		http.Error(res, fmt.Sprintf("No CSVs present"), http.StatusBadRequest)
		return
	}

	vehicle, err := extractVehicleMultipart(req.MultipartForm)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Write([]byte(fmt.Sprintf("Request successful: %d CSVs present, vehicle %v", numCsvs, vehicle)))
}

func extractVehicleMultipart(form *multipart.Form) (recontool.Vehicle, error) {
	vehicleParamsFloat := []string{
		"Rmot",
		"m",
		"Crr1",
		"Crr2",
		"CDa",
		"Tmax",
		"Qmax",
		"Rline",
		"VcMax",
		"VcMin",
	}
	paramValuesFloat := make([]float64, len(vehicleParamsFloat))
	for i, v := range vehicleParamsFloat {
		valueStrings, cont := form.Value[v]
		if !cont || len(valueStrings) < 1 {
			return recontool.Vehicle{}, fmt.Errorf("%s not present", v)
		}
		value, err := strconv.ParseFloat(valueStrings[0], 64)
		if err != nil {
			return recontool.Vehicle{}, err
		}
		paramValuesFloat[i] = value
	}
	vSerStrings, cont := form.Value["Vser"]
	if !cont || len(vSerStrings) < 1 {
		return recontool.Vehicle{}, fmt.Errorf("Vser not present")
	}
	vSer, err := strconv.ParseInt(vSerStrings[0], 10, 32)
	if err != nil {
		return recontool.Vehicle{}, err
	}
	if vSer < 0 {
		return recontool.Vehicle{}, fmt.Errorf("Invalid uint literal %s", vSerStrings[0])
	}
	return recontool.Vehicle{
		RMot:  paramValuesFloat[0],
		M:     paramValuesFloat[1],
		Crr1:  paramValuesFloat[2],
		Crr2:  paramValuesFloat[3],
		CDa:   paramValuesFloat[4],
		TMax:  paramValuesFloat[5],
		QMax:  paramValuesFloat[6],
		RLine: paramValuesFloat[7],
		VcMax: paramValuesFloat[8],
		VcMin: paramValuesFloat[9],
		VSer:  uint(vSer),
	}, nil
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
