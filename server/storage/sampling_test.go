package storage_test

import (
	"os"
	"server/datatypes"
	"server/storage"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSampledPointsForMetric(t *testing.T) {
	_, ok := os.LookupEnv("IN_DOCKER")
	if !ok {
		return
	}
	store, err := storage.NewStorage()
	assert.NoError(t, err)
	start := time.Unix(0, 0)
	end := time.Unix(2, 0)
	resolution := 250
	metric := "Unit_Test_Sampled_Points_For_Metric"
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
	err = store.Insert(points)
	assert.NoError(t, err)

	expectedValues := []float64{0, 1, 1, 3, 3, 4, 4, 4}
	actualValues, err := store.GetSampledPointsForMetric(metric, start, end, resolution)
	assert.NoError(t, err)
	assert.Equal(t, expectedValues, actualValues)
}
