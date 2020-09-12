package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"server/api/merge"
	"server/datatypes"
	"server/storage"

	"github.com/gorilla/mux"
)

// isUploading keeps track of whether we're currently sending a collection of
// datapoints from someone's web browser to a local server instance.
//
// merge_info_config.json stores info about the status of a merge operation from
// a local server to the remote server - not from a browser to a local server.
// Keeping track of this info in merge_info_config.json also wouldn't be a good
// idea beacuse reading to and writing from files aren't atomic operations.
var isUploading atomic.Value

func init() {
	isUploading.Store(false)
}

// MergeHandler handles requests related to merging points from a local
// instance of the server onto the main remote server.
type MergeHandler struct {
	merger *merge.Merger
	store  *storage.Storage
}

// NewMergeHandler returns a pointer to a new MergeHandler.
func NewMergeHandler(store *storage.Storage) *MergeHandler {
	merger, err := merge.NewMerger(store)
	if err != nil {
		log.Fatalf("Error instantiating a new Merger object: %v", err.Error())
	}

	return &MergeHandler{merger, store}
}

type uploadRequest struct {
	startTime *time.Time
	endTime   *time.Time
}

var uploadRequestQueue = make(chan uploadRequest)

// processUploadRequests is a goroutine that listens for new merge requests on
// the uploadRequestQueue chan. When a user fills out the form on the site at
// "/merge", m.LocalMergeHandler() validates that form data, then creates a new
// uploadRequest which will be picked up here.
func (m *MergeHandler) processUploadRequests() {
	for request := range uploadRequestQueue {
		isUploading.Store(true)
		m.uploadLocalPointsToRemote(request.startTime, request.endTime)
		isUploading.Store(false)
	}
}

// MergeDefault is the default handler for the /merge path.
func (m *MergeHandler) MergeDefault(w http.ResponseWriter, r *http.Request) {
	if !isUploading.Load().(bool) {
		http.Redirect(w, r, "/merge/static/index.html", http.StatusFound)
	} else {
		http.Redirect(w, r, "/merge/static/uploading.html", http.StatusFound)
	}
}

// IsUploading handles requests at the /merge/isUploading route.
func (m *MergeHandler) IsUploading(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(isUploading.Load().(bool))
}

// LocalMergeHandler receives form data from the site at "/merge".
func (m *MergeHandler) LocalMergeHandler(w http.ResponseWriter, r *http.Request) {
	statusCode, err := m.localMergeHandlerHelper(r)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// localMergeHandlerHelper performs the business logic for input validation and
// uploading local points to the remote server.
func (m *MergeHandler) localMergeHandlerHelper(r *http.Request) (int, error) {
	// Parse form data from the merge/index.html page.
	err := r.ParseForm()
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf("Could not retrieve input from form: %v", err.Error())
	}

	// Parse that form data into time.Time datatypes.
	tzOffset := r.Form.Get("timezone-offset")
	startTime, err := formatRFC3339(r.Form.Get("start-time"), tzOffset)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf(
				"Could not convert form input into a valid time.Time datatype: %v",
				err.Error(),
			)
	}
	endTime, err := formatRFC3339(r.Form.Get("end-time"), tzOffset)
	if err != nil {
		return http.StatusBadRequest,
			fmt.Errorf(
				"Could not convert form input into a valid time.Time datatype: %v",
				err.Error(),
			)
	}

	// Create an uploadRequest to upload our local points to the remote server.
	select {
	case uploadRequestQueue <- uploadRequest{startTime, endTime}:
		// No-op. Just push the new uploadRequest onto the queue.
	default:
		return http.StatusLocked,
			fmt.Errorf("Another request to upload points from a local server" +
				" to the remote server is in progress.")
	}

	return -1, nil
}

// RemoteMergeHandler inserts provided datapoints into the data store on the
// remote server.
//
// This handler is intended to run on the remote server.
func (m *MergeHandler) RemoteMergeHandler(w http.ResponseWriter, r *http.Request) {
	pointsToMerge := []*datatypes.Datapoint{}
	err := json.NewDecoder(r.Body).Decode(&pointsToMerge)
	if err != nil {
		errMsg := "Failed to unmarshal points from request body into Datapoints: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = m.merger.MergePointsOntoRemote(pointsToMerge)
	if err != nil {
		errMsg := "Failed to insert provided points into remote data store: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Everything went alright. Respond with 204.
	w.WriteHeader(http.StatusNoContent)
}

// uploadLocalPointsToRemote uses information in an uploadRequest to kick off
// the process of merging datapoints on a local server instance to the remote
// server.
func (m *MergeHandler) uploadLocalPointsToRemote(startTime, endTime *time.Time) {
	err := m.merger.UploadLocalPointsToRemote(startTime, endTime)
	if err != nil {
		errMsg := fmt.Errorf("Error uploading local datapoints to remote"+
			" server: %v", err.Error())
		log.Println(errMsg)
	}
}

// Turns date/timezone strings into RFX3339Nano time.Time types.
func formatRFC3339(date string, timezone string) (*time.Time, error) {
	if date == "" || timezone == "" {
		return nil, errors.New("Input to formatRFC3339 cannot be empty")
	}

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

	// Turn string into time datatype
	t, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// RegisterRoutes registers the routes for the merge service.
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
	go m.processUploadRequests()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/merge/static/").Handler(http.StripPrefix("/merge/static/", http.FileServer(http.Dir(path.Join(dir, "merge")))))

	router.HandleFunc("/merge", m.MergeDefault).Methods("GET")
	router.HandleFunc("/merge/isUploading", m.IsUploading).Methods("GET")
	router.HandleFunc("/merge", m.LocalMergeHandler).Methods("POST")
	router.HandleFunc("/remotemerge", m.RemoteMergeHandler).Methods("POST")
}
