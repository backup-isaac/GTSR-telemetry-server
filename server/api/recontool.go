package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"server/recontool"
	"server/storage"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// ReconToolHandler handles requests related to ReconTool
type ReconToolHandler struct {
	store *storage.Storage
}

type csvParse struct {
	csvData    map[string][]float64
	timestamps []int64
	err        error
}

type reconResult struct {
	analysisResult *recontool.AnalysisResult
	err            error
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
	gpsTerrainString := req.Form.Get("terrain")
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
	gpsTerrain, err := strconv.ParseBool(gpsTerrainString)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
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
		return
	}
	timestamps := makeTimestamps(startDate, endDate, resolution)
	results, err := recontool.RunReconTool(data, timestamps, vehicle, gpsTerrain, false)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var resultsArr = [1]*recontool.AnalysisResult{results}
	json.NewEncoder(res).Encode(resultsArr)
}

func makeTimestamps(start, end time.Time, resolution int) []int64 {
	timestamps := make([]int64, end.Sub(start).Milliseconds()/int64(resolution))
	resolutionDur := time.Duration(resolution) * time.Millisecond
	i := 0
	for timestamp := start; timestamp.Before(end); timestamp = timestamp.Add(resolutionDur) {
		timestamps[i] = timestamp.UnixNano() / 1e6
		i++
	}
	return timestamps
}

func extractVehicleForm(form *url.Values) (*recontool.Vehicle, error) {
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
			return &recontool.Vehicle{}, fmt.Errorf("missing parameter %s", v)
		}
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return &recontool.Vehicle{}, err
		}
		paramValuesFloat[i] = value
	}
	vSerString := form.Get("Vser")
	if vSerString == "" {
		return &recontool.Vehicle{}, fmt.Errorf("missing parameter Vser")
	}
	vSer, err := strconv.ParseInt(vSerString, 10, 32)
	if err != nil {
		return &recontool.Vehicle{}, err
	}
	if vSer < 0 {
		return &recontool.Vehicle{}, fmt.Errorf("Invalid uint literal %s", vSerString)
	}
	return &recontool.Vehicle{
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
	gpsTerrainString, ok := req.MultipartForm.Value["terrain"]
	if !ok {
		http.Error(res, "Missing GPS terrain", http.StatusBadRequest)
		return
	}
	gpsTerrain, err := strconv.ParseBool(gpsTerrainString[0])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	plotAllString, ok := req.MultipartForm.Value["autoPlots"]
	if !ok {
		http.Error(res, "Missing plot all", http.StatusBadRequest)
		return
	}
	plotAll, err := strconv.ParseBool(plotAllString[0])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	combineFilesString, ok := req.MultipartForm.Value["compileFiles"]
	if !ok {
		http.Error(res, "Missing plot all", http.StatusBadRequest)
		return
	}
	combineFiles, err := strconv.ParseBool(combineFilesString[0])
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
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
	fileChannels := make([]chan csvParse, len(req.MultipartForm.File))
	i := 0
	for _, file := range req.MultipartForm.File {
		channel := make(chan csvParse)
		go readUploadedCsv(file[0], plotAll, channel)
		fileChannels[i] = channel
		i++
	}
	var parsedCsvs []map[string][]float64
	var parsedTimestamps [][]int64
	if combineFiles {
		parsedCsvs = make([]map[string][]float64, 1)
		parsedTimestamps = make([][]int64, 1)
		first := true
		for _, fileChan := range fileChannels {
			csvParse := <-fileChan
			parsedCsv := csvParse.csvData
			parsedTimestamp := csvParse.timestamps
			err := csvParse.err
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			if first {
				first = false
				parsedCsvs[0] = parsedCsv
				parsedTimestamps[0] = parsedTimestamp
			} else {
				mergeParsedCsvs(parsedCsvs[0], parsedCsv)
				parsedTimestamps[0] = append(parsedTimestamps[0], parsedTimestamp...)
			}
		}
	} else {
		parsedCsvs = make([]map[string][]float64, len(fileChannels))
		parsedTimestamps = make([][]int64, len(fileChannels))
		for i, fileChan := range fileChannels {
			csvParse := <-fileChan
			parsedCsv := csvParse.csvData
			parsedTimestamp := csvParse.timestamps
			err := csvParse.err
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			parsedCsvs[i] = parsedCsv
			parsedTimestamps[i] = parsedTimestamp
		}
	}
	results := make([]*recontool.AnalysisResult, len(parsedCsvs))
	resultChannels := make([]chan reconResult, len(parsedCsvs))
	for i := range results {
		resultChannels[i] = make(chan reconResult)
		go func(data map[string][]float64, timestamp []int64, channel chan reconResult) {
			result, err := recontool.RunReconTool(data, timestamp, vehicle, gpsTerrain, plotAll)
			channel <- reconResult{result, err}
		}(parsedCsvs[i], parsedTimestamps[i], resultChannels[i])
	}
	for i, channel := range resultChannels {
		recon := <-channel
		result := recon.analysisResult
		err := recon.err
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		results[i] = result
	}
	json.NewEncoder(res).Encode(results)
}

func mergeParsedCsvs(csv1, csv2 map[string][]float64) {
	for metric, metricValues := range csv2 {
		arr, cont := csv1[metric]
		if cont {
			csv1[metric] = append(arr, metricValues...)
		} else {
			csv1[metric] = metricValues
		}
	}
}

func readUploadedCsv(fileHeader *multipart.FileHeader, plotAll bool, fileChannel chan csvParse) {
	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		fileChannel <- csvParse{nil, nil, err}
		return
	}
	reader := csv.NewReader(file)
	headerRow, err := reader.Read()
	if err != nil {
		fileChannel <- csvParse{nil, nil, err}
		return
	}
	columns, err := parseColumnNames(headerRow, plotAll)
	if err != nil {
		fileChannel <- csvParse{nil, nil, err}
		return
	}
	csvContents := make(map[string][]float64, len(columns)-1)
	for metric := range columns {
		if metric == "time" {
			continue
		}
		csvContents[metric] = make([]float64, 0, fileHeader.Size/int64(len(columns))/16)
	}
	timestamps := make([]int64, 0, fileHeader.Size/int64(len(columns))/16)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fileChannel <- csvParse{nil, nil, err}
			return
		}
		for metric, index := range columns {
			if index >= len(row) {
				continue
			}
			if metric == "time" {
				value, err := strconv.ParseInt(row[index], 10, 64)
				if err != nil {
					fileChannel <- csvParse{nil, nil, err}
					return
				}
				timestamps = append(timestamps, value)
			} else {
				value, err := strconv.ParseFloat(row[index], 64)
				if err != nil {
					fileChannel <- csvParse{nil, nil, err}
					return
				}
				csvContents[metric] = append(csvContents[metric], value)
			}
		}
	}
	fileChannel <- csvParse{csvContents, timestamps, nil}
}

func parseColumnNames(headers []string, plotAll bool) (map[string]int, error) {
	columnLocs := make(map[string]int)
	columnsFound := 0
	timeLoc := -1
	for i, colName := range headers {
		trimmed := strings.TrimLeft(colName, " ")
		if len(trimmed) == 0 {
			continue
		}
		metric, ok := recontool.MetricHeaderNames[trimmed]
		if ok {
			columnsFound++
		} else if trimmed == recontool.TimeHeaderName {
			if timeLoc != -1 {
				return nil, fmt.Errorf("Duplicate column %s", trimmed)
			}
			timeLoc = i
			metric = "time"
		} else if plotAll {
			metric = trimmed
		} else {
			continue
		}
		_, isDuplicate := columnLocs[metric]
		if isDuplicate {
			return nil, fmt.Errorf("Duplicate column %s", trimmed)
		}
		columnLocs[metric] = i
	}
	if columnsFound < len(recontool.MetricHeaderNames) {
		return nil, fmt.Errorf("Required column(s) missing from CSV. Found columns %v", columnLocs)
	}
	if timeLoc == -1 {
		return nil, fmt.Errorf("Missing time column")
	}
	return columnLocs, nil
}

func extractVehicleMultipart(form *multipart.Form) (*recontool.Vehicle, error) {
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
			return &recontool.Vehicle{}, fmt.Errorf("%s not present", v)
		}
		value, err := strconv.ParseFloat(valueStrings[0], 64)
		if err != nil {
			return &recontool.Vehicle{}, err
		}
		paramValuesFloat[i] = value
	}
	vSerStrings, cont := form.Value["Vser"]
	if !cont || len(vSerStrings) < 1 {
		return &recontool.Vehicle{}, fmt.Errorf("Vser not present")
	}
	vSer, err := strconv.ParseInt(vSerStrings[0], 10, 32)
	if err != nil {
		return &recontool.Vehicle{}, err
	}
	if vSer < 0 {
		return &recontool.Vehicle{}, fmt.Errorf("Invalid uint literal %s", vSerStrings[0])
	}
	return &recontool.Vehicle{
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
