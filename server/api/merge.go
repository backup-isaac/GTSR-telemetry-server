package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"server/api/merge"
	"server/datatypes"
	"server/storage"

	"github.com/gorilla/mux"
)

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

// MergeDefault is the default handler for the /merge path.
func (m *MergeHandler) MergeDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/merge/static/index.html", http.StatusFound)
}

// FormResponse mirrors the contents of the POST request body that the
// LocalMergeHandler parses.
type FormResponse struct {
	TimezoneOffset string `json:"timezoneOffset"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
}

// LocalMergeHandler receives form data from the site at "/merge".
func (m *MergeHandler) LocalMergeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data from the merge/index.html page.
	responses := FormResponse{}
	err := json.NewDecoder(r.Body).Decode(&responses)
	if err != nil {
		http.Error(w, "Could not retrieve input from form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse that form data into time.Time datatypes.
	startTime, err := formatRFC3339(responses.StartTime, responses.TimezoneOffset)
	if err != nil {
		errMsg := "Could not convert form input into a valid time.Time datatype: " + err.Error()
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	// endTime, err := formatRFC3339(r.FormValue("end-time"), tzOffset)
	endTime, err := formatRFC3339(responses.EndTime, responses.TimezoneOffset)
	if err != nil {
		errMsg := "Could not convert form input into a valid time.Time datatype: " + err.Error()
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Begin the upload process to the remote server
	err = m.merger.UploadLocalPointsToRemote(startTime, endTime)
	if err != nil {
		errMsg := "Error uploading local datapoints to remote server: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// TODO: temporary! Have the webpage react to the status code that this
	// handler returns rather than sending content here.
	// fmt.Fprintln(w, "Points collected locally merged successfully")
	w.WriteHeader(http.StatusNoContent)
}

// RemoteMergeHandler inserts provided datapoints into the data store on the
// remote server.
//
// This handler is intended to run on the remote server.
func (m *MergeHandler) RemoteMergeHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Reset this conditional
	// if os.Getenv("PRODUCTION") != "True" {
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

	err = m.merger.MergePointsOntoRemote(pointsToMerge)
	if err != nil {
		errMsg := "Failed to insert provided points into remote data store: " + err.Error()
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Everything went alright. Return 204.
	w.WriteHeader(http.StatusNoContent)
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
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/merge/static/").Handler(http.StripPrefix("/merge/static/", http.FileServer(http.Dir(path.Join(dir, "merge")))))

	router.HandleFunc("/merge", m.MergeDefault).Methods("GET")
	router.HandleFunc("/merge", m.LocalMergeHandler).Methods("POST")
	router.HandleFunc("/remotemerge", m.RemoteMergeHandler).Methods("POST")
}
