package api

import (
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/gorilla/mux"
)

// MapDefault is the default handler for the /map path
func (api *API) MapDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/map/static/index.html", http.StatusFound)
}

// RegisterMapRoutes registers the routes for the map service
func (api *API) RegisterMapRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/map/static/").Handler(http.StripPrefix("/map/static/", http.FileServer(http.Dir(path.Join(dir, "map")))))

	router.HandleFunc("/map", api.MapDefault)
}
