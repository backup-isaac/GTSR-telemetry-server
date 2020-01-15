package main

import (
    "fmt"
    "log"
    "net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, welcome to this test server!")
}

func testHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Sanity Check: 0")
}

func main() {

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)

    http.HandleFunc("/test", testHandler)

    log.Fatal(http.ListenAndServe(":8081", nil))
}
