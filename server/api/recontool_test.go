package api_test

import (
	"net/http"
	"net/url"
	"server/api"
	"server/recontool"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTimeRangeParams(t *testing.T) {
	req := timeRangeRequest()
	goodForm := req.form
	req.Form = nil
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, "Error parsing form")
	req.Form = goodForm
	req.Form.Del("startDate")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Missing startDate")
	req.Form.Set("startDate", "foo")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Error parsing start date:")
	req.Form.Set("startDate", "3133690620000")
	req.Form.Set("endDate", "foo")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Error parsing end date:")
	req.Form.Set("endDate", "3133691220000")
	req.Form.Set("resolution", "foo")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Error parsing resolution:")
	req.Form.Set("resolution", "0")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Resolution must be positive")
	req.Form.Set("resolution", "500")
	req.Form.Set("terrain", "500")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "Error parsing terrain specifier: ")
	req.Form.Set("terrain", "false")
	req.Form.Set("Rmot", "foo")
	_, err := api.parseTimeRangeParams(req)
	assert.Errorf(t, err, "")
	req.Form.Set("Rmot", "0.3")
	params, err := api.parseTimeRangeParams(req)
	assert.NoError(t, err)
	assert.Equal(t, &timeRangeParams{
		start:      time.Date(2069, time.April, 20, 13, 37, 0, 0, time.UTC),
		end:        time.Date(2069, time.April, 20, 13, 47, 0, 0, time.UTC),
		resolution: 500,
		gps:        false,
		vehicle: &recontool.Vehicle{
			RMot:  0.3,
			M:     320,
			Crr1:  0.007,
			Crr2:  0.0005,
			CDa:   0.14,
			TMax:  100,
			QMax:  30,
			RLine: 0.07,
			VcMax: 4.15,
			VcMin: 2.65,
			VSer:  35,
		},
	}, params)
}

func timeRangeRequest() *http.Request {
	form := url.Values()
	form.Set("startDate", "3133690620000")
	form.Set("endDate", "3133691220000")
	form.Set("resolution", "500")
	form.Set("terrain", "false")
	form.Set("Rmot", "0.3")
	form.Set("m", "320")
	form.Set("Crr1", "0.007")
	form.Set("Crr2", "0.0005")
	form.Set("CDa", "0.14")
	form.Set("Tmax", "100")
	form.Set("Qmax", "30")
	form.Set("Rline", "0.07")
	form.Set("VcMax", "4.15")
	form.Set("VcMin", "2.65")
	form.Set("Vser", 35)
	return &http.Request{
		Form: form,
	}
}
