package api_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
	"time"

	"server/api"
	"server/datatypes"

	"github.com/stretchr/testify/assert"
)

type mockStore struct {
	metrics []string
	points  []*datatypes.Datapoint
}

func (m *mockStore) ListMetrics() ([]string, error) {
	return m.metrics, nil
}

func (m *mockStore) SelectMetricTimeRange(name string, start time.Time, end time.Time) ([]*datatypes.Datapoint, error) {
	return m.points, nil
}

func TestGetSampledPointsForMetric(t *testing.T) {
	start := time.Unix(0, 0)
	end := time.Unix(2, 0)
	resolution := 250
	metric := "CSV_Test_Points"
	points := []*datatypes.Datapoint{
		{
			Metric: metric,
			Value:  1,
			Time:   time.Unix(0, 250*1e6),
		},
		{
			Metric: metric,
			Value:  2,
			Time:   time.Unix(0, 251*1e6),
		},
		{
			Metric: metric,
			Value:  3,
			Time:   time.Unix(0, 752*1e6),
		},
		{
			Metric: metric,
			Value:  4,
			Time:   time.Unix(1, 250*1e6),
		},
	}
	expectedValues := []float64{0, 1, 1, 3, 3, 4, 4, 4}
	mockStore := &mockStore{points: points}
	apiObj := api.NewCSVHandler(mockStore)
	actualValues, err := apiObj.GetSampledPointsForMetric(metric, start, end, resolution)
	assert.NoError(t, err)
	assert.Equal(t, expectedValues, actualValues)
}

func TestWriteCsv(t *testing.T) {
	backupCsv(t)
	defer restoreCsv(t)
	start := time.Unix(0, 0)
	end := time.Unix(1, 0)
	resolution := 250
	columns := map[string][]float64{
		"Column_1": []float64{0, 1, 2, 3},
		"Column_2": []float64{4, 5, 6, 7},
	}
	api.WriteCsv(columns, start, end, resolution)
	_, callerFile, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	file, err := os.Open(path.Join(path.Dir(callerFile), "csv/telemetry.csv"))
	assert.NoError(t, err)
	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	assert.NoError(t, err)
	file.Close()
	assert.Equal(t, 5, len(lines))
	heading := lines[0]
	assert.Equal(t, 3, len(heading))
	assert.Equal(t, "time", heading[0])
	assert.Contains(t, heading, "Column_1")
	assert.Contains(t, heading, "Column_2")
	var col1Index, col2Index int
	for i, title := range heading[1:] {
		if title == "Column_1" {
			col1Index = i + 1
		} else if title == "Column_2" {
			col2Index = i + 1
		} else {
			assert.Fail(t, fmt.Sprintf("Unexpected heading: %s", title))
		}
	}
	columns["time"] = []float64{0, 250, 500, 750}
	actualColumns := map[string][]float64{
		"time":     make([]float64, 4),
		"Column_1": make([]float64, 4),
		"Column_2": make([]float64, 4),
	}
	for row, line := range lines[1:] {
		for col, item := range line {
			floatVal, err := strconv.ParseFloat(item, 64)
			assert.NoError(t, err)
			switch col {
			case 0:
				actualColumns["time"][row] = floatVal
			case col1Index:
				actualColumns["Column_1"][row] = floatVal
			case col2Index:
				actualColumns["Column_2"][row] = floatVal
			}
		}
	}
	assert.Equal(t, columns, actualColumns)
}

func backupCsv(t *testing.T) {
	_, callerFile, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	baseDir := path.Dir(callerFile)
	oldpath := path.Join(baseDir, "csv/telemetry.csv")
	newpath := path.Join(baseDir, "csv/telemetry_copy.csv")
	os.Rename(oldpath, newpath)
}

func restoreCsv(t *testing.T) {
	_, callerFile, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	baseDir := path.Dir(callerFile)
	oldpath := path.Join(baseDir, "csv/telemetry_copy.csv")
	newpath := path.Join(baseDir, "csv/telemetry.csv")
	os.Rename(oldpath, newpath)
}
