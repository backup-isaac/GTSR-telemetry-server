package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
	"github.gatech.edu/GTSR/telemetry-server/storage"
)

func TestStorage(t *testing.T) {
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
	storedDatapoints, err := store.SelectMetric("Unit_Test_1")
	assert.NoError(t, err)
	assert.ElementsMatch(t, datapoints, storedDatapoints)
	err = store.DeleteMetric("Unit_Test_1")
	assert.NoError(t, err)
}

func TestInsertEmptyPoints(t *testing.T) {
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	err = store.Insert([]*datatypes.Datapoint{})
	assert.NoError(t, err)
}
