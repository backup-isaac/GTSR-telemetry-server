package api

import (
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/gorilla/mux"
)

const mergePageFilePath = "merge/index.html"

// MergeHandler handles requests related to merging points from a local
// instance of the server onto the main remote server.
type MergeHandler struct{}

// NewMergeHandler returns a pointer to a new MergeHandler.
func NewMergeHandler() *MergeHandler {
	return &MergeHandler{}
}

// Handles requests to merge points
func mergeDefault(res http.ResponseWriter, req *http.Request) {

    // For GET requests, load the form for user to fill out
    fmt.Println("method:", req.Method)
    if req.Method == "GET" {
        t, _ := template.ParseFiles("index.html")
        t.Execute(res, nil)

    // For POST requests, parse the form the user filled out
    } /*else {

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

        // Check if times have been merged yet
        timeCheck,newStart,newEnd := checkTimes(start,end)
        if timeCheck == 2 {
            fmt.Println("All points in the range you selected have been merged before.")
            return
        }

        // Send points to remote server
        err = sendPoints(newStart,newEnd)
        if err != nil {
      		  http.Error(res, fmt.Sprintf("Error merging: %s", err), http.StatusBadRequest)
            fmt.Println("Error merging: %s", err)
        }
*/
    }
}



func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/merge/static/").Handler(http.StripPrefix("/merge/static/", http.FileServer(http.Dir(path.Join(dir, "merge")))))

	router.HandleFunc("/merge", m.MergeDefault)
}
