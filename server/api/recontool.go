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

type timeRangeParams struct {
	start      time.Time
	end        time.Time
	resolution int
	gps        bool
	vehicle    *recontool.Vehicle
}

// ReconTimeRange runs ReconTool on data taken from the server
func (r *ReconToolHandler) ReconTimeRange(res http.ResponseWriter, req *http.Request) {
	params, err := parseTimeRangeParams(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := r.store.GetMetricPointsRange(recontool.MetricNames, params.start, params.end, params.resolution, true)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	timestamps := makeTimestamps(params.start, params.end, params.resolution)
	results, err := recontool.RunReconTool(data, timestamps, params.vehicle, params.gps, false)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var resultsArr = [1]*recontool.AnalysisResult{results}
	json.NewEncoder(res).Encode(resultsArr)
}

func parseTimeRangeParams(req *http.Request) (*timeRangeParams, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("Error parsing form: %s", err)
	}
	formParams := []string{"startDate", "endDate", "resolution", "terrain"}
	paramStrings := make(map[string]string, len(formParams))
	for _, p := range formParams {
		paramStrings[p] = req.Form.Get(p)
		if paramStrings[p] == "" {
			return nil, fmt.Errorf("Missing %s", p)
		}
	}
	startDate, err := unixStringMillisToTime(paramStrings["startDate"])
	if err != nil {
		return nil, fmt.Errorf("Error parsing start date: %s", err)
	}
	endDate, err := unixStringMillisToTime(paramStrings["endDate"])
	if err != nil {
		return nil, fmt.Errorf("Error parsing end date: %s", err)
	}
	resolution64, err := strconv.ParseInt(paramStrings["resolution"], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Error parsing resolution: %s", err)
	}
	resolution := int(resolution64)
	if resolution <= 0 {
		return nil, fmt.Errorf("Resolution must be positive")
	}
	gpsTerrain, err := strconv.ParseBool(paramStrings["terrain"])
	if err != nil {
		return nil, fmt.Errorf("Error parsing terrain specifier: %s", err)
	}
	vehicle, err := extractVehicleForm(&req.Form)
	if err != nil {
		return nil, err
	}
	return &timeRangeParams{
		start:      startDate,
		end:        endDate,
		resolution: resolution,
		gps:        gpsTerrain,
		vehicle:    vehicle,
	}, nil
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

type csvParams struct {
	gps          bool
	plotAll      bool
	combineFiles bool
	numCsvs      int
	vehicle      *recontool.Vehicle
}

func parseCsvParams(req *http.Request) (*csvParams, error) {
	err := req.ParseMultipartForm(1048576)
	if err != nil {
		return nil, err
	}
	formParams := []string{"terrain", "autoPlots", "compileFiles"}
	paramVals := make(map[string]bool, len(formParams))
	for _, p := range formParams {
		paramString, ok := req.MultipartForm.Value[p]
		if !ok {
			return nil, fmt.Errorf("Missing %s", p)
		}
		paramVal, err := strconv.ParseBool(paramString[0])
		if err != nil {
			return nil, err
		}
		paramVals[p] = paramVal
	}
	numCsvs := len(req.MultipartForm.File)
	if numCsvs == 0 {
		return nil, fmt.Errorf("No CSVs present")
	}
	vehicle, err := extractVehicleMultipart(req.MultipartForm)
	if err != nil {
		return nil, err
	}
	return &csvParams{
		gps:          paramVals["terrain"],
		plotAll:      paramVals["autoPlots"],
		combineFiles: paramVals["compileFiles"],
		numCsvs:      numCsvs,
		vehicle:      vehicle,
	}, nil
}

// ReconCSV runs ReconTool on data provided as a CSV
func (r *ReconToolHandler) ReconCSV(res http.ResponseWriter, req *http.Request) {
	params, err := parseCsvParams(req)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing multipart form: %s", err), http.StatusBadRequest)
		return
	}
	fileChannels := make([]chan csvParse, len(req.MultipartForm.File))
	i := 0
	for _, file := range req.MultipartForm.File {
		channel := make(chan csvParse)
		go readUploadedCsv(file[0], params.plotAll, channel)
		fileChannels[i] = channel
		i++
	}
	var parsedCsvs []map[string][]float64
	var parsedTimestamps [][]int64
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
	if params.combineFiles {
		combinedCsv, combinedTimestamps := mergeParsedCsvs(parsedCsvs, parsedTimestamps)
		parsedCsvs = []map[string][]float64{combinedCsv}
		parsedTimestamps = [][]int64{combinedTimestamps}
	}
	results := make([]*recontool.AnalysisResult, len(parsedCsvs))
	resultChannels := make([]chan reconResult, len(parsedCsvs))
	for i := range results {
		resultChannels[i] = make(chan reconResult)
		go func(data map[string][]float64, timestamp []int64, channel chan reconResult) {
			result, err := recontool.RunReconTool(data, timestamp, params.vehicle, params.gps, params.plotAll)
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

func mergeParsedCsvs(csvs []map[string][]float64, timestamps [][]int64) (map[string][]float64, []int64) {
	combinedLength := 0
	arrayIndices := make([]int, len(timestamps))
	arraysLeft := make([][]int64, len(timestamps))
	csvsLeft := make([]map[string][]float64, len(timestamps))
	for i, t := range timestamps {
		combinedLength += len(t)
		arraysLeft[i] = t
		csvsLeft[i] = csvs[i]
	}
	newTimestamps := make([]int64, combinedLength)
	newCsv := make(map[string][]float64)
	totIndex := 0
	for len(arrayIndices) > 0 {
		indexArgMin := 0
		for i := 1; i < len(arrayIndices); i++ {
			if arraysLeft[i][arrayIndices[i]] < arraysLeft[indexArgMin][arrayIndices[indexArgMin]] {
				indexArgMin = i
			}
		}
		newTimestamps[totIndex] = arraysLeft[indexArgMin][arrayIndices[indexArgMin]]
		for metric, metricValue := range csvsLeft[indexArgMin] {
			_, cont := newCsv[metric]
			if !cont {
				newCsv[metric] = make([]float64, len(metricValue))
			}
			newCsv[metric] = append(newCsv[metric], metricValue[arrayIndices[indexArgMin]])
		}
		totIndex++
		arrayIndices[indexArgMin]++
		if arrayIndices[indexArgMin] >= len(arraysLeft[indexArgMin]) {
			copy(arrayIndices[indexArgMin:], arrayIndices[indexArgMin+1:])
			arrayIndices = arrayIndices[:len(arrayIndices)-1]
			copy(arraysLeft[indexArgMin:], arraysLeft[indexArgMin+1:])
			arraysLeft[len(arraysLeft)-1] = nil
			arraysLeft = arraysLeft[:len(arraysLeft)-1]
			copy(csvsLeft[indexArgMin:], csvsLeft[indexArgMin+1:])
			csvsLeft[len(csvsLeft)-1] = nil
			csvsLeft = csvsLeft[:len(csvsLeft)-1]
		}
	}
	return newCsv, newTimestamps
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
		metric, ok := recontool.LoggerMetricHeaders[trimmed]
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
	if columnsFound < len(recontool.LoggerMetricHeaders) {
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
