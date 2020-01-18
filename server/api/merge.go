package api

import (
	"net/http"
	"fmt"
	"html/template"

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
func mergeHandler(res http.ResponseWriter, req *http.Request) {

    // For GET requests, load the form for user to fill out
    fmt.Println("method:", req.Method)
    if req.Method == "GET" {
        t, _ := template.ParseFiles("index.html")
        t.Execute(res, nil)
    }
}

func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
  router.HandleFunc("/merge", m.mergeHandler)
}
