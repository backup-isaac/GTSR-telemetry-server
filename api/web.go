package api

import (
        "log"
        "net/http"
        "path"
        "runtime"

        "github.com/gorilla/mux"
)

// WebDefault is the default handler for the /web path
func (api *API) WebDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/web/static/lab1.html", http.StatusFound)
}

// RegisterCsvRoutes registers the routes for the CSV service
func (api *API) RegisterWebRoutes(router *mux.Router) {
        _, filename, _, ok := runtime.Caller(0)
        if !ok {
                log.Fatal("Could not find runtime caller")
        }
        dir := path.Dir(filename)
        router.PathPrefix("/web/static/").Handler(http.StripPrefix("/web/static/", http.FileServer(http.Dir(path.Join(dir, "web")))))
	
	router.HandleFunc("/web", api.WebDefault).Methods("GET")

}
