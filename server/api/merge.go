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

// MergeDefault is the default handler for the /merge path.
func (m *MergeHandler) MergeDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/merge/static/index.html", http.StatusFound)
}

// MergePost is the POST handler for the /merge path.
func (m *MergeHandler) MergePost(w http.ResponseWriter, r *http.Request) {

  // Parse Form
  err := r.ParseForm()
  if err != nil {
    http.Error(w, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
    fmt.Println("Error parsing form: %s", err)
    return
  }

  // Get data from from
  timezone := r.Form.Get("timezone-offset")
  startDateString := r.Form.Get("start")
  endDateString := r.Form.Get("end")
  if startDateString == "" || endDateString == "" || timezone == "" {
    http.Error(w, "malformatted query", http.StatusBadRequest)
    return
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
