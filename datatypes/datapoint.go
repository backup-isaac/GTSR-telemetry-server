package datatypes

// Datapoint is a container for raw data from the car
type Datapoint struct {
	// Metric is the name of the metric type for this datapoint
	// Examples: Wavesculptor RPM, BMS Current
	Metric string
	// Value of this datapoint
	Value interface{}
	// Map of tags associated with this datapoint (e.g. event tags)
	Tags map[string]string
}
