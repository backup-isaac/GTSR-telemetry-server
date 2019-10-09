package computations

import (
	"server/datatypes"
	"server/listener"
)

// Computable is the base interface which every computation must implement
type Computable interface {
	Update(point *datatypes.Datapoint) bool
	Compute() *datatypes.Datapoint
	GetMetrics() []string
}

var registry []Computable

// Register registers a computation
func Register(computation Computable) {
	registry = append(registry, computation)
}

// RunComputations is the main function, which spawns goroutines for every computation and routes
// incoming data points to their associated computations
func RunComputations() {
	for _, computation := range registry {
		stream := make(chan *datatypes.Datapoint, 100)
		listener.Subscribe(stream, computation.GetMetrics()...)
		go compute(computation, stream)
	}
}

func compute(computation Computable, stream chan *datatypes.Datapoint) {
	publisher := listener.GetDatapointPublisher()
	for point := range stream {
		if computation.Update(point) {
			publisher.Publish(computation.Compute())
		}
	}
}
