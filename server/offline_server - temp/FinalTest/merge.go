package telemetry-server
package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"server/datatypes"
	"storage"
	"time"

	"github.com/gorilla/mux"
)

// Object that handles HTTP Merge requests
type MergeHandler struct {
	store *storage.Storage
}

// Returns a MergeHandler
func NewMergeHandler(store *storage.Storage) *MergeHandler {
	return &MergeHandler{store}
}

//Merge Function, to be called from local host via http page
//Move to html page loaded locally or command line or http requests
//Only available if not production
//Make sure that no duplicate points are sent, send the points that aren't overlapped
func (m *MergeHandler) startMerge(start time.Time, end time.Time) {

	//For each metric
	for _, metric := range m.store.ListMetrics() {

		//Get data from specified metric/timerange
		points, err := m.store.SelectMetricTimeRange(metric, start, end)
		if err != nil {
			return err
		}

		//Send n points at a time
		index := 0
		for index < len(points) {

			//Initialize json slice
			batch := []*datatypes.Datapoint{}

			//Add 20 points at a time to slice
			batch.append(points[index])
			index++
			for index%20 != 0 && index < len(points) {
				batch.append(points[index])
				index++
			}

			//Encode points
			jsonPoints, _ := json.Marshal(batch)

			//Post encoded point to html server
			resp, err = http.NewRequest("POST", "http://solarracing.me/api/merge", bytes.NewBuffer(&jsonPoints))
			if err != nil {
				return err
			}

		}

	}

}

//Merge Handler, handles post requests to "/merge"
func (m *MergeHandler) Merge(res http.ResponseWriter, req *http.Request) {

	//Initialize empty point
	points := []*datatypes.Datapoint{}

	//Parse json data into point
	err := json.NewDecoder(req.Body).Decode(&points)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	//Insert point into server
	m.store.Insert(points)
}

//Registers route for merging
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
  router.HandleFunc("/merge", api.Merge).Methods("POST")
  /*
	if !os.Getenv("PRODUCTION") { //Need to check if "PRODUCTION" is correct
		router.HandleFunc("/merge", api.Merge).Methods("POST")
	}*/
}
