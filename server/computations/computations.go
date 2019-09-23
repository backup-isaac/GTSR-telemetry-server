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
	streams := make(map[string][]chan *datatypes.Datapoint)
	for _, computation := range registry {
		stream := make(chan *datatypes.Datapoint, 100)
		for _, metric := range computation.GetMetrics() {
			streams[metric] = append(streams[metric], stream)
		}
		go compute(computation, stream)
	}

	points := make(chan *datatypes.Datapoint, 1000)
	listener.Subscribe(points)

	for point := range points {
		for _, stream := range streams[point.Metric] {
			stream <- point
		}
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
