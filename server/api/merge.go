package api

import (
	"fmt"
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

// MergeDefault is the default handler for the /merge path.
func (m *MergeHandler) MergeDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/merge/static/index.html", http.StatusFound)
}

// MergePointsHandler receives form data from the site at "/merge".
func (m *MergeHandler) MergePointsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// TODO: Write a better error message omg
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "form data: %s, %s, %s",
		r.Form.Get("timezone-offset"),
		r.Form.Get("start-time"),
		r.Form.Get("end-time"),
	)
}

// RegisterRoutes registers the routes for the merge service.
func (m *MergeHandler) RegisterRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/merge/static/").Handler(http.StripPrefix("/merge/static/", http.FileServer(http.Dir(path.Join(dir, "merge")))))

	router.HandleFunc("/merge", m.MergeDefault).Methods("GET")
	router.HandleFunc("/merge", m.MergePointsHandler).Methods("POST")
}
