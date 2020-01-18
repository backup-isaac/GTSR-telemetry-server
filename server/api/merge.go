package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"server/datatypes"
	"server/storage"

	"github.com/gorilla/mux"
)

const (
	mergePageFilePath = "merge/index.html"
	remoteMergeURL    = "https://solarracing.me/remotemerge"
)

// MergeHandler handles requests related to merging points from a local
// instance of the server onto the main remote server.
type MergeHandler struct {
	store *storage.Storage
}

// NewMergeHandler returns a pointer to a new MergeHandler.
func NewMergeHandler(store *storage.Storage) *MergeHandler {
	return &MergeHandler{store}
}

// MergeDefault is the default handler for the /merge path.
func (m *MergeHandler) MergeDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/merge/static/index.html", http.StatusFound)
}

// MergePointsHandler receives form data from the site at "/merge".
func (m *MergeHandler) MergePointsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data from the merge/index.html page.
	err := r.ParseForm()
	if err != nil {
		// TODO: Write a better error message omg
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse that form data into time.Time datatypes.
	tzOffset := r.Form.Get("timezone-offset")
	startTime, err := formatRFC3339(r.Form.Get("start-time"), tzOffset)
	if err != nil {
		// TODO: Write a better error message omg
		http.Error(w, "Failed to parse form data as a valid type: "+err.Error(), http.StatusInternalServerError)
		return
	}
	endTime, err := formatRFC3339(r.Form.Get("end-time"), tzOffset)
	if err != nil {
		// TODO: Write a better error message omg
		http.Error(w, "Failed to parse form data as a valid type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all points (of all metric types) within the specified time range.
	pointsToMerge := []*datatypes.Datapoint{}
	metrics, err := m.store.ListMetrics()
	if err != nil {
		errMsg := "Failed to list all metrics in the data store. This shouldn't happen."
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	for _, metric := range metrics {
		// We can safely dereference these pointers since we handle any errors
		// during their creation above.
		newPoints, err := m.store.SelectMetricTimeRange(
			metric, *startTime, *endTime,
		)
		if err != nil {
			// TODO: Write a better error message omg
			http.Error(w, "fk", http.StatusInternalServerError)
			return
		}
		pointsToMerge = append(pointsToMerge, newPoints...)
	}

	// Marshal the list of metrics that we've collected to JSON format.
	pointsAsJSON, err := json.Marshal(pointsToMerge)
	if err != nil {
		errMsg := "Failed to marshal data points into JSON: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Merge points into remote server's data store.
	res, err := http.Post(remoteMergeURL, "application/json", bytes.NewBuffer(pointsAsJSON))
	if err != nil {
		errMsg := "Failed to send POST request to solarracing.me/remotemerge: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	if res.StatusCode != 204 {
		errMsg := "POST request to solarracing.me/remotemerge did not return 204:" + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
}

// RemoteMergeHandler takes datapoints and inserts them into the data store on
// the remote server.
//
// The endpoint that this handler is responsible for should only be hit on the
// remote server.
func (m *MergeHandler) RemoteMergeHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PRODUCTION") == "False" {
		errMsg := "This endpoint should only be hit on the remote server"
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	pointsToMerge := []*datatypes.Datapoint{}
	err := json.NewDecoder(r.Body).Decode(&pointsToMerge)
	if err != nil {
		errMsg := "Failed to unmarshal points from request body into Datapoints: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = m.store.Insert(pointsToMerge)
	if err != nil {
		errMsg := "Failed to insert all points into remote data store: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Everything went alright. Return 204.
	w.WriteHeader(http.StatusNoContent)
}

// Turns date/timezone strings into RFX3339Nano time.Time types.
func formatRFC3339(date string, timezone string) (*time.Time, error) {
	// If there aren't seconds on the date string passed in, add them.
	if len(date) == 16 {
		date += ":00"
	}
	// Add nanoseconds
	// date += ".999999999"
	// Add timezone
	if timezone[0] == '+' || timezone[0] == '-' {
		date += timezone
	} else {
		date += "Z"
	}

	log.Println(date)

	// Turn string into time datatype
	t, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// RegisterRoutes registers the routes for the merge service.
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/merge/static/").Handler(http.StripPrefix("/merge/static/", http.FileServer(http.Dir(path.Join(dir, "merge")))))

	router.HandleFunc("/merge", m.MergeDefault).Methods("GET")
	router.HandleFunc("/merge", m.MergePointsHandler).Methods("POST")
	router.HandleFunc("/remotemerge", m.RemoteMergeHandler).Methods("POST")
}
