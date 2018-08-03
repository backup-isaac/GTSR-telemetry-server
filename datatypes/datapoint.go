package datatypes

import (
	"time"
)

// Datapoint is a container for raw data from the car
type Datapoint struct {
	// Metric is the name of the metric type for this datapoint
	// Examples: Wavesculptor RPM, BMS Current
	Metric string
	// Value of this datapoint
	Value float64
	// Map of tags associated with this datapoint (e.g. event tags)
	Tags map[string]string
	// Time is the time that this datapoint was inserted into the database
	Time time.Time
}
