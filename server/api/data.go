package api

import (
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/gorilla/mux"
)

// DataHandler handles requests related to the CAN configuration display tool
type DataHandler struct{}

// NewDataHandler is the DataHandler constructor
func NewDataHandler() *DataHandler {
	return &DataHandler{}
}

// DataDefault is the default handler for the /data path
func (d *DataHandler) DataDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/data/static/data.html", http.StatusFound)
}

// RegisterRoutes registers the routes for the data service
func (d *DataHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/data/static/").Handler(http.StripPrefix("/data/static/", http.FileServer(http.Dir(path.Join(dir, "data")))))

	router.HandleFunc("/data", d.DataDefault).Methods("GET")
}
