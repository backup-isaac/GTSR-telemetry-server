package listener

import (
	"fmt"
	"sync"
)

// DatapointPublisher allows threads to subscribe to a particular publisher,
// in this case the tcp port listener
type DatapointPublisher interface {
	Subscribe(c chan *Datapoint) error
	Unsubscribe(c chan *Datapoint) error
	Publish(point *Datapoint)
}

// single(ton) reaccs only :(
var globalPublisher *datapointPublisher

// NewDatapointPublisher returns a new DatapointPublisher with the standard
// implementation, and starts the publisher thread
func NewDatapointPublisher() DatapointPublisher {
	if globalPublisher != nil {
		return globalPublisher
	}
	publisher := &datapointPublisher{
		PublishChannel:  make(chan *Datapoint),
		Subscribers:     []chan *Datapoint{},
		SubscribersLock: new(sync.Mutex),
	}
	globalPublisher = publisher
	go publisher.publisherThread()
	return publisher
}

type datapointPublisher struct {
	PublishChannel  chan *Datapoint
	Subscribers     []chan *Datapoint
	SubscribersLock *sync.Mutex
}

// Subscribe will add a channel to the list of Subscribers
func (publisher *datapointPublisher) Subscribe(c chan *Datapoint) error {
	publisher.SubscribersLock.Lock()
	defer publisher.SubscribersLock.Unlock()
	publisher.Subscribers = append(publisher.Subscribers, c)
	return nil
}

// Unsubscribe will remove a channel from the list of Subscribers
func (publisher *datapointPublisher) Unsubscribe(c chan *Datapoint) error {
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

// Publish data
func (publisher *datapointPublisher) Publish(point *Datapoint) {
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
