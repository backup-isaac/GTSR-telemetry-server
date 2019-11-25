package storage

import (
	"fmt"
	"log"
	"time"
)

// GetSampledPointsForMetric returns sampled data for a particular metric in the time range specified by start and end
// at the given resolution
func (s *Storage) GetSampledPointsForMetric(metric string, start time.Time, end time.Time, resolution int) ([]float64, error) {
	points, err := s.SelectMetricTimeRange(metric, start, end)
	if err != nil {
		return nil, err
	}
	duration := end.Sub(start)
	durMillis := duration.Nanoseconds() / 1e6
	resolutionDur := time.Duration(resolution) * time.Millisecond
	numRows := int(durMillis-1)/resolution + 1
	column := make([]float64, numRows)
	var last float64
	i := 0
	row := 0
	for timestamp := start; timestamp.Before(end); timestamp = timestamp.Add(resolutionDur) {
		for i < len(points) && points[i].Time.Before(timestamp) {
			i++
		}
		next := timestamp.Add(resolutionDur)
		if i >= len(points) || points[i].Time.Equal(next) || points[i].Time.After(next) {
			column[row] = last
		} else {
			column[row] = points[i].Value
			last = points[i].Value
		}
		row++
	}
	return column, nil
}

// GetAllMetricPointsRange returns sampled data for all metrics in the specified time range
func (s *Storage) GetAllMetricPointsRange(start time.Time, end time.Time, resolution int) (map[string][]float64, error) {
	metrics, err := s.ListMetrics()
	if err != nil {
		return nil, err
	}
	return s.GetMetricPointsRange(metrics, start, end, resolution, false)
}

// GetMetricPointsRange returns sampled data for the specified metrics in the specified time range
func (s *Storage) GetMetricPointsRange(metrics []string, start time.Time, end time.Time, resolution int, strict bool) (map[string][]float64, error) {
	colChannels := make([]chan []float64, len(metrics))
	for i, metric := range metrics {
		colChannels[i] = make(chan []float64, 1)
		go func(metric string, colChan chan []float64) {
			column, err := s.GetSampledPointsForMetric(metric, start, end, resolution)
			if err != nil {
				log.Printf("Error getting values for metric %s: %s\n", metric, err)
				colChan <- nil
				return
			}
			colChan <- column
		}(metric, colChannels[i])
	}
	columns := make(map[string][]float64, len(metrics))
	for i, colChan := range colChannels {
		column := <-colChan
		if column != nil {
			columns[metrics[i]] = column
		} else if strict {
			return nil, fmt.Errorf("Unable to get values for metric %s", metrics[i])
		}
	}
	return columns, nil
}
