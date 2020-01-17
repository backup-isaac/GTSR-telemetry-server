package api

import (
  	"fmt"
    //"os"
  	//"log"
  	"net/http"
    "html/template"
    //"strings"
    //"time"
    //"bytes"
    //"encoding/json"
    "server/storage"
    "github.com/gorilla/mux"
)


// Storage for MergeHandler
type MergeHandler struct {
    store *storage.Storage
}

// Passes server storage in for access
func NewMergeHandler(store *storage.Storage) *MergeHandler {
    return &MergeHandler{store}
}

// Handles requests to merge points
func (m *MergeHandler) mergeHandler(res http.ResponseWriter, req *http.Request) {

    // For GET requests, load the form for user to fill out
    fmt.Println("method:", req.Method)
    if req.Method == "GET" {
        t, _ := template.ParseFiles("/merge/index.html")
        t.Execute(res, nil)
    }

}

// Register routes to each handler
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
  router.HandleFunc("/merge", m.mergeHandler)
}
