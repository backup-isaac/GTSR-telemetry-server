package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// RouteHandler describes objects which handle requests
// to specific HTTP endpoints
type RouteHandler interface {
	RegisterRoutes(*mux.Router)
}

// StartServer registers all of the routes handled
// by the server and runs the HTTP server
func StartServer(handlers []RouteHandler) {
	router := mux.NewRouter()
	for _, handler := range handlers {
		handler.RegisterRoutes(router)
	}
	log.Println("Starting HTTP server...")
	log.Fatal(http.ListenAndServe(":8888", router))
}
