package main

import (
	"log"
	"time"

	"github.gatech.edu/GTSR/telemetry-server/datatypes"
	"github.gatech.edu/GTSR/telemetry-server/listener"
	"github.gatech.edu/GTSR/telemetry-server/storage"
)

func main() {
	go listener.Listen()
	err := recordData()
	log.Fatalf("Error recording data: %s", err)
}

func recordData() error {
	points := make(chan *datatypes.Datapoint)
	err := listener.Subscribe(points)
	if err != nil {
		return err
	}
	defer listener.Unsubscribe(points)
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}
	defer store.Close()
	bufferedPoints := make([]*datatypes.Datapoint, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case point := <-points:
			point.Time = time.Now()
			bufferedPoints = append(bufferedPoints, point)
		case <-ticker.C:
			err = store.Insert(bufferedPoints)
			if err != nil {
				return err
			}
			bufferedPoints = make([]*datatypes.Datapoint, 0, 100)
		}
	}
}
