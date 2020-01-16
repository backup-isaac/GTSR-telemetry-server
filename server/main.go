package main

import (
	"log"
	"time"

	"server/api"
	"server/computations"
	"server/datatypes"
	"server/listener"
	"server/storage"
)

func main() {
	go listener.TCPListen()
	go listener.UDPListen()
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Error initializing storage: %s", err)
	}
	defer store.Close()
	go api.StartServer([]api.RouteHandler{
		api.NewChatHandler(),
		api.NewCore(store),
		api.NewCSVHandler(store),
		api.NewDataHandler(),
		api.NewFoodHandler(),
		api.NewJacksonHandler(),
		api.NewMapHandler(),
		api.NewReconToolHandler(store),
		api.NewMergeHandler(),
	})
	go computations.RunComputations()
	err = recordData(store)
	log.Fatalf("Error recording data: %s", err)
}

func recordData(store *storage.Storage) error {
	points := make(chan *datatypes.Datapoint, 1000)
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
