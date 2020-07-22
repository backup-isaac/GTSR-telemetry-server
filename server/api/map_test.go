package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const fileUploadBaseURI = "/map/fileupload"

// Tests the /map/fileupload's response when no file is present in the request
// body. Everything else that the map.FileUpload() handler does is either
// tested in other areas, or is a well-tested func from the standard library.
func TestFileUploadHandler(t *testing.T) {
	req, err := http.NewRequest("POST", fileUploadBaseURI, nil)
	if err != nil {
		t.Fatal("Failed to construct request to the map.FileUpload() handler:",
			err.Error(),
		)
	}
	rr := httptest.NewRecorder()
	mapHandler := NewMapHandler()
	// Call the handler
	mapHandler.FileUpload(rr, req)
	// Compare status code with expected status code
	actualStatusCode := rr.Result().StatusCode
	expectedStatusCode := http.StatusBadRequest
	if actualStatusCode != expectedStatusCode {
		t.Errorf("Handler responded with unexpected status code. got: %d want: %d\n",
			actualStatusCode, expectedStatusCode,
		)
	}
	// Compare response body with expected response body
	actualResponseBody := strings.TrimSuffix(rr.Body.String(), "\n")
	expectedResponseBody := "Error getting uploaded file: missing form body"
	if actualResponseBody != expectedResponseBody {
		t.Errorf("Handler responded with unexpected response body.\n\tgot: %s\n\twant: %s\n",
			actualResponseBody, expectedResponseBody,
		)
	}
}
