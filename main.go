package main

import (
	"log"
	"time"

	"github.gatech.edu/GTSR/telemetry-server/api"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
	"github.gatech.edu/GTSR/telemetry-server/listener"
	"github.gatech.edu/GTSR/telemetry-server/storage"
)

func main() {
	go listener.Listen()
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Error initializing storage: %s", err)
	}
	defer store.Close()
	apiObj := api.NewAPI(store)
	go apiObj.StartServer()
	err = recordData(store)
	log.Fatalf("Error recording data: %s", err)
}

func recordData(store storage.Storage) error {
	points := make(chan *datatypes.Datapoint)
	err := listener.Subscribe(points)
	if err != nil {
		return err
	}
	defer listener.Unsubscribe(points)
	bufferedPoints := make([]*datatypes.Datapoint, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case point := <-points:
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
