package computations

import (
	"telemetry-server/datatypes"
	"telemetry-server/listener"
)

// Computable is the base interface which every computation must implement
type Computable interface {
	Update(point *datatypes.Datapoint) bool
	Compute() *datatypes.Datapoint
}

// standardComputation is a container for the normal computation which just needs one
// point of each metric type to perform its computation
type standardComputation struct {
	values map[string]float64
	fields []string
}

// Update of standardComputation simply puts the point into the metrics map
// and returns whether the map is full
func (c *standardComputation) Update(point *datatypes.Datapoint) bool {
	c.values[point.Metric] = point.Value
	return len(c.values) >= len(c.fields)
}

// RunComputations is the main function, which runs all the computations in the registry based on incoming points
func RunComputations() {
	streams := make(map[string][]chan *datatypes.Datapoint)
	for computation, metrics := range registry {
		stream := make(chan *datatypes.Datapoint, 100)
		for _, metric := range metrics {
			streams[metric] = append(streams[metric], stream)
		}
		go computationThread(computation, stream)
	}

	points := make(chan *datatypes.Datapoint, 1000)
	listener.Subscribe(points)

	for {
		point := <-points
		for _, stream := range streams[point.Metric] {
			stream <- point
		}
	}
}

func computationThread(computation Computable, stream chan *datatypes.Datapoint) {
	publisher := listener.GetDatapointPublisher()
	for {
		point := <-stream
		if computation.Update(point) {
			publisher.Publish(computation.Compute())
		}
	}
}
