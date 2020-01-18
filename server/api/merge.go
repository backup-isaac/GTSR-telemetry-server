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

func(m *MergeHandler) MergePoints(w http.ResponseWriter, r *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
		return
	}
	startDateString := req.Form.Get("startDate")
	endDateString := req.Form.Get("endDate")

	if startDateString == "" || endDateString == "" || resolutionString == "" {
		http.Error(res, "malformatted query", http.StatusBadRequest)
		return
	}
	startDate, err := unixStringMillisToTime(startDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing start date: %s", err), http.StatusBadRequest)
		return
	}
	endDate, err := unixStringMillisToTime(endDateString)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error parsing end date: %s", err), http.StatusBadRequest)
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

	router.HandleFunc("/merge", m.MergeDefault).Methods("GET")
	router.HandleFunc("/merge/mergePoints", m.MergePoints).Methods("POST")
}
