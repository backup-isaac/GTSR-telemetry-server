package computations

import (
	"server/datatypes"
	"time"
)

// standardComputation is a container for the normal computation which just needs one
// point of each metric type to perform its computation
type standardComputation struct {
	values    map[string]float64
	timestamp time.Time
	fields    []string
}

// Update of standardComputation simply puts the point into the metrics map
// and returns whether the map is full
func (c *standardComputation) Update(point *datatypes.Datapoint) bool {
	c.values[point.Metric] = point.Value
	if point.Time.After(c.timestamp) {
		c.timestamp = point.Time
	}
	return len(c.values) >= len(c.fields)
}

// GetMetrics of standardComputation returns its list of fields
func (c *standardComputation) GetMetrics() []string {
	return c.fields
}
