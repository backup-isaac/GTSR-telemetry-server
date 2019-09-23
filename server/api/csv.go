package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"server/datatypes"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

var generating atomic.Value

func init() {
	generating.Store(false)
}

type csvStore interface {
	ListMetrics() ([]string, error)
	SelectMetricTimeRange(string, time.Time, time.Time) ([]*datatypes.Datapoint, error)
}

// CSVHandler handles requests related to the CSV generator tool
type CSVHandler struct {
	store csvStore
}

// NewCSVHandler returns an initialized CSVHandler
func NewCSVHandler(store csvStore) *CSVHandler {
	return &CSVHandler{store}
}

type generationRequest struct {
	start      time.Time
	end        time.Time
	resolution int
}

var genQueue = make(chan generationRequest)

func (c *CSVHandler) generationScheduler() {
	for req := range genQueue {
		generating.Store(true)
		c.generateCsv(req.start, req.end, req.resolution)
		generating.Store(false)
	}
}

// CsvDefault is the default handler for the /csv route
func (c *CSVHandler) CsvDefault(res http.ResponseWriter, req *http.Request) {
	if !generating.Load().(bool) {
		http.Redirect(res, req, "/csv/static/index.html", http.StatusFound)
	} else {
		http.Redirect(res, req, "/csv/static/generating.html", http.StatusFound)
	}
}

// IsGenerating returns whether a CSV is currently being generated
func (c *CSVHandler) IsGenerating(res http.ResponseWriter, req *http.Request) {
	json.NewEncoder(res).Encode(generating.Load().(bool))
}

// GenerateCsv generates the csv
func (c *CSVHandler) GenerateCsv(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
		return
	}
	startDateString := req.Form.Get("startDate")
	endDateString := req.Form.Get("endDate")
	resolutionString := req.Form.Get("resolution")
	if startDateString == "" || endDateString == "" || resolutionString == "" {
		http.Error(res, "malformatted query", http.StatusBadRequest)
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
		http.Error(res, "Resolution must be strictly greater than 0", http.StatusBadRequest)
		return
	}
	select {
	case genQueue <- generationRequest{startDate, endDate, resolution}:
	default:
		http.Error(res, "Already generating CSV", http.StatusLocked)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func unixStringMillisToTime(timeString string) (time.Time, error) {
	timeMillis, err := strconv.ParseInt(timeString, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, timeMillis*1e6), nil
}

func (c *CSVHandler) generateCsv(start time.Time, end time.Time, resolution int) {
	metrics, err := c.store.ListMetrics()
	if err != nil {
		log.Printf("Error getting metrics: %s\n", err)
		return
	}
	colChannels := make([]chan []float64, len(metrics))
	for i, metric := range metrics {
		colChannels[i] = make(chan []float64, 1)
		go func(metric string, colChan chan []float64) {
			column, err := c.GetSampledPointsForMetric(metric, start, end, resolution)
			if err != nil {
				log.Printf("Error getting values for metric %s: %s\n", metric, err)
				colChan <- nil
				return
			}
			colChan <- column
		}(metric, colChannels[i])
	}
	columns := make(map[string][]float64, len(metrics))
	for i, colChan := range colChannels {
		column := <-colChan
		if column != nil {
			columns[metrics[i]] = column
		}
	}
	WriteCsv(columns, start, end, resolution)
}

// GetSampledPointsForMetric returns sampled data for a particular metric in the time range specified by start and end
// at the given resolution
func (c *CSVHandler) GetSampledPointsForMetric(metric string, start time.Time, end time.Time, resolution int) ([]float64, error) {
	points, err := c.store.SelectMetricTimeRange(metric, start, end)
	if err != nil {
		return nil, err
	}
	duration := end.Sub(start)
	durMillis := duration.Nanoseconds() / 1e6
	resolutionDur := time.Duration(resolution) * time.Millisecond
	numRows := int(durMillis-1)/resolution + 1
	column := make([]float64, numRows)
	var last float64
	i := 0
	row := 0
	for timestamp := start; timestamp.Before(end); timestamp = timestamp.Add(resolutionDur) {
		for i < len(points) && points[i].Time.Before(timestamp) {
			i++
		}
		next := timestamp.Add(resolutionDur)
		if i >= len(points) || points[i].Time.Equal(next) || points[i].Time.After(next) {
			column[row] = last
		} else {
			column[row] = points[i].Value
			last = points[i].Value
		}
		row++
	}
	return column, nil
}

// WriteCsv writes the contents of columns to api/csv/telemetry.csv
// with the first column being the timestamp of each row based on the
// start, end and resolution
func WriteCsv(columns map[string][]float64, start time.Time, end time.Time, resolution int) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Could not find runtime caller")
		return
	}
	csvFn := path.Join(path.Dir(filename), "csv/telemetry.csv")
	file, err := os.Create(csvFn)
	if err != nil {
		log.Println("Could not create api/csv/telemetry.csv")
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rowContents := make([]string, len(columns)+1)
	columnIndices := make(map[string]int, len(columns))
	rowContents[0] = "time"
	col := 1
	for metric := range columns {
		rowContents[col] = metric
		columnIndices[metric] = col
		col++
	}
	writer.Write(rowContents)

	resolutionDur := time.Duration(resolution) * time.Millisecond
	row := 0
	for rowTime := start; rowTime.Before(end); rowTime = rowTime.Add(resolutionDur) {
		rowContents[0] = fmt.Sprintf("%d", rowTime.UnixNano()/1e6)
		for metric, column := range columns {
			rowContents[columnIndices[metric]] = fmt.Sprintf("%v", column[row])
		}
		writer.Write(rowContents)
		row++
	}
}

// RegisterRoutes registers the routes for the CSV service
func (c *CSVHandler) RegisterRoutes(router *mux.Router) {
	go c.generationScheduler()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/csv/static/").Handler(http.StripPrefix("/csv/static/", http.FileServer(http.Dir(path.Join(dir, "csv")))))

	router.HandleFunc("/csv", c.CsvDefault).Methods("GET")
	router.HandleFunc("/csv/isGenerating", c.IsGenerating).Methods("GET")
	router.HandleFunc("/csv/generateCsv", c.GenerateCsv).Methods("POST")
}
