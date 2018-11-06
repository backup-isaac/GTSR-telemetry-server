package api

import (
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/aonemd/margopher"

	"github.com/gorilla/mux"
)

// GenerateJacksonSpeech uses data.txt, a file containing
// all of Jackson's Slack messages, to generate sagely advice
func (api *API) GenerateJacksonSpeech(res http.ResponseWriter, req *http.Request) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	dataFile := path.Join(dir, "jackson_samples.txt")
	chain := margopher.New()
	res.Write([]byte(chain.ReadFile(dataFile)))
}

// RegisterJacksonRoutes registers the POST request route
// for the joke Jackson chatbot integration in Slack
func (api *API) RegisterJacksonRoutes(router *mux.Router) {
	router.HandleFunc("/jackson", api.GenerateJacksonSpeech).Methods("GET")
}
