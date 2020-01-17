package main

import (

  	"fmt"
  	"log"
  	"net/http"
    "html/template"
    "strings"
)

/*
func stringToTime(input string, select string) (output int) {
    if select == "date"{

        // RFC3339Nano?
        // RFC3339     = "2006-01-02T15:04:05Z07:00"
        // RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
        // Z can be replaced by + or -

        // Parse string into times
        year := input[0:4]
        month := input[5:7]
        day := input[8:10]
        hour := input[11:13]
        minute := input[14:16]
        second := 0
        if len(input) > 16{
          second := input[17:]
        }
    }
}
*/

func formatRFC3339(date string, timezone string) (string) {

      // Cut off milliseconds
      date = date[:16]

      if timezone[0] == '+' || timezone[0] == '-' {
          return date + timezone
      } else {
          return date + "Z"
      }
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
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

func mergeHandler(res http.ResponseWriter, req *http.Request) {

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

func main() {
    http.HandleFunc("/", defaultHandler) // setting router rule
    http.HandleFunc("/merge", mergeHandler)
    err := http.ListenAndServe(":9090", nil) // setting listening port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
