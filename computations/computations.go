package computations

import (
	"log"
	"sync"

	"github.gatech.edu/GTSR/telemetry-server/datatypes"
	"github.gatech.edu/GTSR/telemetry-server/listener"
	"github.gatech.edu/GTSR/telemetry-server/storage"
)

// Computable is the base interface which every computation must implement
type Computable interface {
	Update(point *datatypes.Datapoint) bool
	Compute() *datatypes.Datapoint
}

// standardComputation is a container for the normal computation which just needs one
// point of each metric type to perform its computation
type standardComputation struct {
	sync.Mutex
	values map[string]float64
	fields []string
}

// Update of standardComputation simply puts the point into the metrics map
// and returns whether the map is full
func (c *standardComputation) Update(point *datatypes.Datapoint) bool {
	c.Lock()
	defer c.Unlock()
	c.values[point.Metric] = point.Value
	return len(c.values) >= len(c.fields)
}

// RunComputations is the main function, which runs all the computations in the registry based on incoming points
func RunComputations() {
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Error initializing storage in computations: %s", err)
	}
	points := make(chan *datatypes.Datapoint)
	listener.Subscribe(points)
	for {
		point := <-points
		for _, computable := range LoadComputables(point.Metric) {
			if computable.Update(point) {
				go func(computable Computable) {
					computedPoint := computable.Compute()
					err := store.Insert([]*datatypes.Datapoint{computedPoint})
					if err != nil {
						log.Printf("Error inserting point into datastore: %s\n", err)
					}
					points <- computedPoint // For computations which rely on other computations
				}(computable)
			}
		}
	}
}
