package listener

import (
	"fmt"
	"sync"

	"telemetry-server/datatypes"
)

// DatapointPublisher allows threads to subscribe to a particular publisher,
// in this case the tcp port listener
type DatapointPublisher interface {
	// Subscribe will add a channel to the list of Subscribers
	Subscribe(c chan *datatypes.Datapoint) error
	// Unsubscribe will remove a channel from the list of Subscribers
	Unsubscribe(c chan *datatypes.Datapoint) error
	// Publish data
	Publish(point *datatypes.Datapoint)
}

// single(ton) reaccs only :(
var globalPublisher *datapointPublisher
var publisherLock sync.Mutex

// GetDatapointPublisher creates the global DatapointPublisher
// if it has not been created and returns it
func GetDatapointPublisher() DatapointPublisher {
	publisherLock.Lock()
	defer publisherLock.Unlock()
	if globalPublisher != nil {
		return globalPublisher
	}
	publisher := &datapointPublisher{
		PublishChannel:  make(chan *datatypes.Datapoint, 10000),
		Subscribers:     []chan *datatypes.Datapoint{},
		SubscribersLock: new(sync.Mutex),
	}
	globalPublisher = publisher
	go publisher.publisherThread()
	return publisher
}

type datapointPublisher struct {
	PublishChannel  chan *datatypes.Datapoint
	Subscribers     []chan *datatypes.Datapoint
	SubscribersLock *sync.Mutex
}

func (publisher *datapointPublisher) Subscribe(c chan *datatypes.Datapoint) error {
	publisher.SubscribersLock.Lock()
	defer publisher.SubscribersLock.Unlock()
	publisher.Subscribers = append(publisher.Subscribers, c)
	return nil
}

func (publisher *datapointPublisher) Unsubscribe(c chan *datatypes.Datapoint) error {
	publisher.SubscribersLock.Lock()
	defer publisher.SubscribersLock.Unlock()
	for i, channel := range publisher.Subscribers {
		if c == channel {
			publisher.Subscribers = append(publisher.Subscribers[:i], publisher.Subscribers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Unsubscribe: channel not found in subscriber list")
}

func (publisher *datapointPublisher) Publish(point *datatypes.Datapoint) {
	publisher.PublishChannel <- point
}

func (publisher *datapointPublisher) publisherThread() {
	for {
		point := <-publisher.PublishChannel
		publisher.SubscribersLock.Lock()
		for _, subscriber := range publisher.Subscribers {
			subscriber <- point
		}
		publisher.SubscribersLock.Unlock()
	}
}
