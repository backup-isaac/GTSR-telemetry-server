package api

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

// MergeHandler handles requests related to merging points from a local
// instance of the server onto the main remote server.
type MergeHandler struct{}

// NewMergeHandler returns a pointer to a new MergeHandler.
func NewMergeHandler() *MergeHandler {
	return &MergeHandler{}
}

func formatRFC3339(date string, timezone string) string {

	// Cut off milliseconds
	date = date[:16]

	if timezone[0] == '+' || timezone[0] == '-' {
		return date + timezone
	}
	return date + "Z"
}

func (m *MergeHandler) mergeHandlerGet(res http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("merge/index.html")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(res, nil)
}

func (m *MergeHandler) mergeHandlerPost(res http.ResponseWriter, req *http.Request) {

	// Parse Form
	req.ParseForm()

	// Get data from from
	timezone := req.Form.Get("timezone-offset")
	startDateString := req.Form.Get("start")
	endDateString := req.Form.Get("end")

	// Format data
	start := formatRFC3339(startDateString, timezone)
	end := formatRFC3339(endDateString, timezone)

	// Check if values are received correctly
	fmt.Println("timezone: ", timezone)
	fmt.Println("start:    ", startDateString)
	fmt.Println("end:      ", endDateString)
	fmt.Println()

	// Check if values are formatted correctly
	fmt.Println("Formatted start: ", start)
	fmt.Println("Formatted end:   ", end)

	// Retrieve Datapoints

	// Inform user of success
	fmt.Fprintf(res, "Form submitted sucessfully! Check your command line for status updates")
}

// RegisterRoutes registers routes
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/merge", m.mergeHandlerGet).Methods("GET")
	router.HandleFunc("/merge", m.mergeHandlerPost).Methods("POST")
}
