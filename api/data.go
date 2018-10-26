package api

import (
        "log"
        "net/http"
        "path"
        "runtime"

        "github.com/gorilla/mux"
)

// DataDefault is the default handler for the /data path
func (api *API) DataDefault(res http.ResponseWriter, req *http.Request) {
        http.Redirect(res, req, "/data/static/data.html", http.StatusFound)
}

// RegisterCsvRoutes registers the routes for the CSV service
func (api *API) RegisterDataRoutes(router *mux.Router) {
        _, filename, _, ok := runtime.Caller(0)
        if !ok {
log.Fatal("Could not find runtime caller")
        }
        dir := path.Dir(filename)
      router.PathPrefix("/data/static/").Handler(http.StripPrefix("/data/static/", http.FileServer(http.Dir(path.Join(dir, "data")))))

        router.HandleFunc("/data", api.DataDefault).Methods("GET")

}

