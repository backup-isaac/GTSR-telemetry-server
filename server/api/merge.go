package api

import (
	"net/http"
	"fmt"
	"html/template"
	"strings"

	"github.com/gorilla/mux"
)

// MergeHandler handles requests related to merging points from a local
// instance of the server onto the main remote server.
type MergeHandler struct{}

// NewMergeHandler returns a pointer to a new MergeHandler.
func NewMergeHandler() *MergeHandler {
	return &MergeHandler{}
}

func formatRFC3339(date string, timezone string) (string) {

      // Cut off milliseconds
      date = date[:16]

      if timezone[0] == '+' || timezone[0] == '-' {
          return date + timezone
      } else {
          return date + "Z"
      }
}

func (m *MergeHandler) defaultHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
    // attention: If you do not call ParseForm method, the following data can not be obtained form
    fmt.Println(r.Form) // print information on server side.
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
    fmt.Fprintf(w, "Hello! Go to localhost:9090/merge to merge points") // write data to response
}

func (m *MergeHandler) mergeHandler(res http.ResponseWriter, req *http.Request) {

    fmt.Println("method:", req.Method)
    if req.Method == "GET" {
        t, _ := template.ParseFiles("index1.html")
        t.Execute(res, nil)
    } else {

        // Parse Form
        req.ParseForm()

        // Get data from from
        timezone := req.Form.Get("timezone-offset")
        startDateString := req.Form.Get("start")
	      endDateString := req.Form.Get("end")

        // Format data
        start := formatRFC3339(startDateString,timezone)
        end := formatRFC3339(endDateString,timezone)

        // Check if values are received correctly
        fmt.Println("timezone: ", timezone)
        fmt.Println("start:    ", startDateString)
        fmt.Println("end:      ", endDateString)
        fmt.Println()

        // Check if values are formatted correctly
        fmt.Println("Formatted start: ",start)
        fmt.Println("Formatted end:   ",end)

        // Retrieve Datapoints


        // Inform user of success
        fmt.Fprintf(res, "Form submitted sucessfully! Check your command line for status updates")
    }
}

func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
    http.HandleFunc("/merge", m.defaultHandler)
  router.HandleFunc("/merge/merge", m.mergeHandler)
}
