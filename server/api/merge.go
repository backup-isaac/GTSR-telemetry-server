package main

import (
  	"fmt"
  	"log"
  	"net/http"
    "html/template"
    "strings"
    "time"
    "bytes"
    "encoding/json"
    "server/storage"
)


// Storage for MergeHandler
type MergeHandler struct {
    store *storage.Storage
}

// Passes server storage in for access
func NewMergeHandler(store *storage.Storage) *MergeHandler {
    return &MergeHandler{store}
}


// Turns date/timezone strings into RFX3339Nano time.Time
func formatRFC3339(date string, timezone string) (time.Time,error) {

    // Make sure there are seconds
    if len(date) == 16 {
        date += ":00"
    }

    // Add nanoseconds
    date += ".999999999"

    // Add timezone
    if timezone[0] == '+' || timezone[0] == '-' {
        date = date + timezone
    } else {
        date = date + "Z"
    }

    // Parse string into time.Time() datatype
    t,err := time.Parse(time.RFC3339Nano,date)

    // Return the time and any errors
    return t,err
}

// Handles requests to merge points
func mergeHandler(res http.ResponseWriter, req *http.Request) {

    // For GET requests, load the form for user to fill out
    fmt.Println("method:", req.Method)
    if req.Method == "GET" {
        t, _ := template.ParseFiles("/merge/index.html")
        t.Execute(res, nil)

    // For POST requests, parse the form the user filled out
    } else {

        // Parse Form
        err := req.ParseForm()
      	if err != nil {
      		http.Error(res, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
          fmt.Println("Error parsing form: %s", err)
      		return
      	}

        // Get data from from
        timezone := req.Form.Get("timezone-offset")
        startDateString := req.Form.Get("start")
	      endDateString := req.Form.Get("end")
        if startDateString == "" || endDateString == "" || timezone == "" {
      		http.Error(res, "malformatted query", http.StatusBadRequest)
          fmt.Println("malformatted query")
          return
      	}

        // Format start time
        start,err := formatRFC3339(startDateString,timezone)
        if err != nil{
            http.Error(res, fmt.Sprintf("Error formatting to RFC3339: %s",err),http.StatusBadRequest)
        }

        // Format end time
        end,err := formatRFC3339(endDateString,timezone)
        if err != nil{
            http.Error(res, fmt.Sprintf("Error formatting to RFC3339: %s",err),http.StatusBadRequest)
        }

        // Check if values are received correctly
        fmt.Println("timezone: ", timezone)
        fmt.Println("start:    ", startDateString)
        fmt.Println("end:      ", endDateString)
        fmt.Println()

        // Check if values are formatted correctly
        fmt.Println("Formatted start: ",start)
        fmt.Println("Formatted end:   ",end)

        // Inform user of success
        fmt.Fprintf(res, "Form submitted sucessfully! Check your command line for status updates")
    }
}

// Send points between start and end times
func (m *MergeHandler) sendPoints(start time.Time, end time.Time) (error) {

  // Run this code only if it is the local server
  // Need to check this still
  if os.Getenv("PRODUCTION") == true{
    return nil
  }

  /*
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

			//Post encoded points to html server
			resp, err := http.NewRequest("POST", "http://solarracing.me/receive", bytes.NewBuffer(&jsonPoints))
			if err != nil {
				return err
			}
		}
	}
  */
}

/*
// Receive the sent points
func (m *MergeHandler) receivePoints(res http.ResponseWriter, req *http.Request) (error) {

  // Run this code only if it is the remote server
  // Need to check this still
  if os.Getenv("Production") == "False":
    return nil

	// Initialize empty points
	points := []*datatypes.Datapoint{}

	// Parse json data into points
	err := json.NewDecoder(req.Body).Decode(&points)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Insert points into server
	err = m.store.Insert(points)

  return err
}
*/

// Register routes to each handler
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
  router.HandleFunc("/merge", c.mergeHandler)
  //router.HandleFunc("/receive", c.receivePoints)
}
