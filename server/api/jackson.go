package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/aonemd/margopher"

	"github.com/gorilla/mux"
)

// JacksonHandler handles requests from /advice
type JacksonHandler struct{}

// NewJacksonHandler is the basic JacksonHandler constructor
func NewJacksonHandler() *JacksonHandler {
	return &JacksonHandler{}
}

// GenerateJacksonSpeech uses jackson_samples.txt, a file containing
// all of Jackson's Slack messages, to generate sagely advice
func (j *JacksonHandler) GenerateJacksonSpeech(res http.ResponseWriter, req *http.Request) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	dataFile := path.Join(dir, "jackson_samples.txt")

	response := make(map[string]string)
	response["response_type"] = "in_channel"
	chain := margopher.New()
	response["text"] = chain.ReadFile(dataFile)
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(response)
}

// RegisterRoutes registers the GET request route
// for the joke Jackson chatbot integration in Slack
func (j *JacksonHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/jackson", j.GenerateJacksonSpeech).Methods("GET")
}
