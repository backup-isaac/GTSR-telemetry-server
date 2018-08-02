package main

import (
	"log"

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
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}
	defer store.Close()
	for {
		err = store.Insert([]*datatypes.Datapoint{<-points})
		if err != nil {
			return err
		}
	}
}
