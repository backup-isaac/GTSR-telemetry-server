package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"runtime"
	"server/datatypes"
	"server/listener"
	"server/message"
	"server/track"
	"strconv"

	"github.com/gorilla/mux"
)

// MapHandler handles requests related to the map service,
// which includes serving the Google Maps frontend for tracking the
// car, as well as the tool for uploading suggested speeds
type MapHandler struct {
	track *track.Track
}

// NewMapHandler is the basic MapHandler constructor
func NewMapHandler() *MapHandler {
	track, err := track.NewTrack(message.NewCarMessenger("GT", listener.NewTCPWriter()))
	if err != nil {
		log.Fatalf("Error getting Track: %+v", err)
	}
	return &MapHandler{
		track: track,
	}
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
	err = m.track.UploadRoute(points)
	if err != nil {
		http.Error(res, "Error uploading points: "+err.Error(), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

// Route gets the current route
func (m *MapHandler) Route(res http.ResponseWriter, req *http.Request) {
	route, err := m.track.GetRoute()
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	err = json.NewEncoder(res).Encode(route)
	if err != nil {
		http.Error(res, "Malformatted route: "+err.Error(), http.StatusInternalServerError)
	}
}

// ParseRouteCsv returns the parsed list of RoutePoints from the uploaded CSV file
func ParseRouteCsv(file multipart.File) ([]*datatypes.RoutePoint, error) {
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	columns, err := getColumns(header)
	if err != nil {
		return nil, err
	}
	var routePoints []*datatypes.RoutePoint
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

func parseRow(row []string, columns map[string]int) (*datatypes.RoutePoint, error) {
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
	return &datatypes.RoutePoint{
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
	router.HandleFunc("/map/route", m.Route).Methods("GET")
}
