package storage_test

import (
	"os"
	"testing"
	"time"

	"server/datatypes"
	"server/storage"

	"github.com/stretchr/testify/assert"
)

// These tests are skipped if this is not running in a Docker environment
// so that the tests can pass in a local environment without an InfluxDB
// connection

func TestStorage(t *testing.T) {
	_, ok := os.LookupEnv("IN_DOCKER")
	if !ok {
		return
	}
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	err = store.DeleteMetric("Unit_Test_1")
	assert.NoError(t, err)
	utc, err := time.LoadLocation("UTC")
	assert.NoError(t, err)
	datapoints := []*datatypes.Datapoint{
		{
			Metric: "Unit_Test_1",
			Value:  12345,
			Time:   time.Date(2069, time.April, 20, 4, 20, 0, 0, utc),
		},
		{
			Metric: "Unit_Test_1",
			Value:  54321,
			Time:   time.Date(2018, time.May, 21, 0, 0, 0, 0, utc),
		},
	}
	err = store.Insert(datapoints)
	assert.NoError(t, err)
	metrics, err := store.ListMetrics()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(metrics))
	unitTestInMetrics := false
	for _, metric := range metrics {
		if metric == "Unit_Test_1" {
			unitTestInMetrics = true
			break
		}
	}
	assert.True(t, unitTestInMetrics, "Unit_Test_1 not found in metrics")
	storedDatapoints, err := store.SelectMetric("Unit_Test_1")
	assert.NoError(t, err)
	assert.ElementsMatch(t, datapoints, storedDatapoints)
	storedDatapoints, err = store.SelectMetricTimeRange(
		"Unit_Test_1",
		time.Date(2060, time.January, 1, 0, 0, 0, 0, utc),
		time.Date(2070, time.January, 1, 0, 0, 0, 0, utc),
	)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []*datatypes.Datapoint{datapoints[0]}, storedDatapoints)
	latest, err := store.Latest("Unit_Test_1")
	assert.NoError(t, err)
	assert.Equal(t, datapoints[0], latest)
	err = store.DeleteMetric("Unit_Test_1")
	assert.NoError(t, err)
}

func TestLatestNonZero(t *testing.T) {
	_, ok := os.LookupEnv("IN_DOCKER")
	if !ok {
		return
	}
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	utc, err := time.LoadLocation("UTC")
	assert.NoError(t, err)
	datapoints := []*datatypes.Datapoint{{
		Metric: "Unit_Test_2",
		Value:  0,
		Time:   time.Date(2069, time.April, 20, 0, 0, 0, 0, utc),
	}, {
		Metric: "Unit_Test_2",
		Value:  12345,
		Time:   time.Date(2018, time.May, 21, 0, 0, 0, 0, utc),
	}}
	err = store.Insert(datapoints)
	assert.NoError(t, err)
	point, err := store.LatestNonZero("Unit_Test_2")
	assert.NoError(t, err)
	assert.Equal(t, datapoints[1], point)
}

func TestInsertEmptyPoints(t *testing.T) {
	_, ok := os.LookupEnv("IN_DOCKER")
	if !ok {
		return
	}
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	err = store.Insert([]*datatypes.Datapoint{})
	assert.NoError(t, err)
}

func TestInOrder(t *testing.T) {
	_, ok := os.LookupEnv("IN_DOCKER")
	if !ok {
		return
	}
	points := []*datatypes.Datapoint{
		{
			Metric: "Unit_Test_Order",
			Value:  0,
			Time:   time.Unix(3, 0).UTC(),
		},
		{
			Metric: "Unit_Test_Order",
			Value:  1,
			Time:   time.Unix(1, 0).UTC(),
		},
		{
			Metric: "Unit_Test_Order",
			Value:  2,
			Time:   time.Unix(2, 0).UTC(),
		},
	}
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	err = store.Insert(points)
	assert.NoError(t, err)
	storedPoints, err := store.SelectMetric("Unit_Test_Order")
	assert.NoError(t, err)
	points = []*datatypes.Datapoint{points[1], points[2], points[0]}
	assert.Equal(t, points, storedPoints)
	storedPoints, err = store.SelectMetricTimeRange("Unit_Test_Order", time.Unix(0, 0), time.Unix(3, 0))
	assert.NoError(t, err)
	assert.Equal(t, points, storedPoints)
	err = store.DeleteMetric("Unit_Test_Order")
	assert.NoError(t, err)
}

func TestIllegalMetricName(t *testing.T) {
	for _, metric := range []string{
		"metric space",
		"metric\nnewline",
	} {
		points := []*datatypes.Datapoint{{
			Metric: metric,
			Value:  0,
			Time:   time.Now(),
		}}
		store, err := storage.NewStorage()
		assert.NoError(t, err)
		err = store.Insert(points)
		assert.Errorf(t, err, "Metric %q should have been rejected", metric)
	}
}
