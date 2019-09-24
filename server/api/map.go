package api

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"server/api/trackinfo"
	"server/listener"

	"github.com/gorilla/mux"
)

var trackInfoMutex = sync.Mutex{}

const trackInfoConfigPath = "trackinfo/track_info_config.json"

// MapHandler handles requests related to the map service,
// which includes serving the Google Maps frontend for tracking the
// car, as well as the tool for uploading suggested speeds
type MapHandler struct{}

// NewMapHandler is the basic MapHandler constructor
func NewMapHandler() *MapHandler {
	return &MapHandler{}
}

// RoutePoint is a point along the uploaded route
type RoutePoint struct {
	// Distance is the distance along the route for this point
	Distance float64 `json:"distance"`
	// Latitude is the GPS latitude of this point
	Latitude float64 `json:"latitude"`
	// Longitude is the GPS longitude of this point
	Longitude float64 `json:"longitude"`
	// Speed is the suggested speed for the car at this point
	Speed float64 `json:"speed"`
	// Critical is a flag for whether this is a significant datapoint
	// that should be sent to the car to be suggested to the driver
	Critical bool `json:"critical"`
}

// MapDefault is the default handler for the /map path
func (m *MapHandler) MapDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/map/static/index.html", http.StatusFound)
}

// FileUpload handles a CSV upload of a race route and suggested speeds
func (m *MapHandler) FileUpload(res http.ResponseWriter, req *http.Request) {
	file, _, err := req.FormFile("uploadedFile")
	if err != nil {
		http.Error(res, "Error getting uploaded file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	points, err := ParseRouteCsv(file)
	if err != nil {
		http.Error(res, "Error parsing provided CSV: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		http.Error(res, "Unable to save route JSON", http.StatusInternalServerError)
		return
	}
	jsonFile, err := os.Create(path.Join(path.Dir(callerFile), "/map/route.json"))
	if err != nil {
		http.Error(res, "Error saving route JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer jsonFile.Close()
	err = json.NewEncoder(jsonFile).Encode(points)
	if err != nil {
		http.Error(res, "Error saving route JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	go uploadPoints(points)

	err = editIsTrackInfoNew(true)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

// ParseRouteCsv returns the parsed list of RoutePoints from the uploaded CSV file
func ParseRouteCsv(file multipart.File) ([]*RoutePoint, error) {
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	columns, err := getColumns(header)
	if err != nil {
		return nil, err
	}
	var routePoints []*RoutePoint
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		point, err := parseRow(row, columns)
		if err != nil {
			return nil, err
		}
		routePoints = append(routePoints, point)
	}
	if len(routePoints) == 0 {
		return nil, fmt.Errorf("Unable to parse any route points from provided CSV")
	}
	return routePoints, nil
}

func parseRow(row []string, columns map[string]int) (*RoutePoint, error) {
	distance, err := strconv.ParseFloat(row[columns["distance"]], 64)
	if err != nil {
		return nil, err
	}
	latitude, err := strconv.ParseFloat(row[columns["latitude"]], 64)
	if err != nil {
		return nil, err
	}
	longitude, err := strconv.ParseFloat(row[columns["longitude"]], 64)
	if err != nil {
		return nil, err
	}
	speed, err := strconv.ParseFloat(row[columns["speed"]], 64)
	if err != nil {
		return nil, err
	}
	critical := row[columns["critical"]] == "1"
	return &RoutePoint{
		Distance:  distance,
		Latitude:  latitude,
		Longitude: longitude,
		Speed:     speed,
		Critical:  critical,
	}, nil
}

func getColumns(header []string) (map[string]int, error) {
	columns := make(map[string]int)
	for i, colName := range header {
		columns[colName] = i
	}
	if err := verifyColumns(columns); err != nil {
		return nil, err
	}
	return columns, nil
}

func verifyColumns(columns map[string]int) error {
	expectedColumns := []string{"distance", "latitude", "longitude", "speed", "critical"}
	if len(columns) != len(expectedColumns) {
		return fmt.Errorf("Mismatched number of columns in provided CSV: expected %d, got %d", len(expectedColumns), len(columns))
	}
	for _, colName := range expectedColumns {
		if _, ok := columns[colName]; !ok {
			return fmt.Errorf("Column '%s' not found in provided CSV", colName)
		}
	}
	return nil
}

func uploadPoints(points []*RoutePoint) {
	w := listener.NewTCPWriter()
	tag := []byte("GTSR")
	w.Write(tag)
	for _, point := range points {
		if point.Critical {
			writeFloat64As32(point.Distance)
			writeFloat64As32(point.Latitude)
			writeFloat64As32(point.Longitude)
			writeFloat64As32(point.Speed)
			time.Sleep(100 * time.Millisecond)
		}
	}
	w.Write(tag)
}

func writeFloat64As32(num float64) {
	num32 := float32(num)
	bits := math.Float32bits(num32)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)
	listener.NewTCPWriter().Write(buf)
}

// editIsTrackInfoNew overwrites the contents of track_info_config.json's
// "isTrackInfoNew" key with the provided bool
func editIsTrackInfoNew(value bool) error {
	trackInfoModel := trackinfo.Model{}

	err := readTrackInfoConfig(&trackInfoModel)
	if err != nil {
		return err
	}

	trackInfoModel.IsTrackInfoNew = value

	err = writeToTrackInfoConfig(&trackInfoModel)
	if err != nil {
		return err
	}

	return nil
}

// readTrackInfoConfig reads the track info config from disk and unmarshals it
// into the provided struct
func readTrackInfoConfig(m *trackinfo.Model) error {
	trackInfoMutex.Lock()
	defer trackInfoMutex.Unlock()

	configFile, err := ioutil.ReadFile(trackInfoConfigPath)
	if err != nil {
		return errors.New("Error reading track_info_config: " + err.Error())
	}

	json.Unmarshal(configFile, &m)
	return nil
}

// writeToTrackInfoConfig writes the contents of the provided struct to the
// track info config on disk
func writeToTrackInfoConfig(m *trackinfo.Model) error {
	trackInfoMutex.Lock()
	defer trackInfoMutex.Unlock()

	jsonAsBytes, err := json.Marshal(m)
	if err != nil {
		return errors.New("Error editing track_info_config: " + err.Error())
	}

	err = ioutil.WriteFile(trackInfoConfigPath, jsonAsBytes, 0644)
	if err != nil {
		return errors.New("Error writing changes to track_info_config: " + err.Error())
	}

	return nil
}

// RegisterRoutes registers the routes for the map service
func (m *MapHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/map/static/").Handler(http.StripPrefix("/map/static/", http.FileServer(http.Dir(path.Join(dir, "map")))))

	router.HandleFunc("/map", m.MapDefault).Methods("GET")
	router.HandleFunc("/map/fileupload", m.FileUpload).Methods("POST")
}
