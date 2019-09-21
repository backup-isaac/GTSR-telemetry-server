package listener

import (
	"fmt"
	"sync"

	"server/datatypes"
)

// DatapointPublisher provides pub/sub functionality for Datapoint streams
type DatapointPublisher struct {
	publishChannel  chan *datatypes.Datapoint
	subscribers     []chan *datatypes.Datapoint
	subscribersLock *sync.Mutex
}

func newDatapointPublisher() *DatapointPublisher {
	publisher := &DatapointPublisher{
		publishChannel:  make(chan *datatypes.Datapoint, 10000),
		subscribers:     []chan *datatypes.Datapoint{},
		subscribersLock: new(sync.Mutex),
	}
	go publisher.publisherThread()
	return publisher
}

// single(ton) reaccs only :(
var globalPublisher *DatapointPublisher
var publisherLock sync.Mutex

// GetDatapointPublisher creates the global DatapointPublisher
// if it has not been created and returns it
func GetDatapointPublisher() *DatapointPublisher {
	publisherLock.Lock()
	defer publisherLock.Unlock()
	if globalPublisher != nil {
		return globalPublisher
	}
	publisher := newDatapointPublisher()
	globalPublisher = publisher
	return publisher
}

// Subscribe subscribes the given channel to the publisher. Whenever Publish is
// called on this publisher, the datapoint will be sent to the provided channel
func (publisher *DatapointPublisher) Subscribe(c chan *datatypes.Datapoint) error {
	publisher.subscribersLock.Lock()
	defer publisher.subscribersLock.Unlock()
	publisher.subscribers = append(publisher.subscribers, c)
	return nil
}

// Unsubscribe unsubscribes a channel from the publisher. Datapoints published
// to the publisher will no longer be sent to the provided channel
func (publisher *DatapointPublisher) Unsubscribe(c chan *datatypes.Datapoint) error {
	publisher.subscribersLock.Lock()
	defer publisher.subscribersLock.Unlock()
	for i, channel := range publisher.subscribers {
		if c == channel {
			publisher.subscribers = append(publisher.subscribers[:i], publisher.subscribers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Unsubscribe: channel not found in subscriber list")
}

// Publish publishes a given datapoint. The datapoint will be sent to all subscribing channels
func (publisher *DatapointPublisher) Publish(point *datatypes.Datapoint) {
	publisher.publishChannel <- point
}

// Close closes the given publisher, stopping the publishing thread and closing all subscriber channels
func (publisher *DatapointPublisher) Close() {
	publisherLock.Lock()
	defer publisherLock.Unlock()
	close(publisher.publishChannel)
	publisher.subscribersLock.Lock()
	defer publisher.subscribersLock.Unlock()
	for _, c := range publisher.subscribers {
		close(c)
	}
	if publisher == globalPublisher {
		globalPublisher = nil
	}
}

func (publisher *DatapointPublisher) publisherThread() {
	for point := range publisher.publishChannel {
		publisher.subscribersLock.Lock()
		for _, subscriber := range publisher.subscribers {
			subscriber <- point
		}
		publisher.subscribersLock.Unlock()
	}
}
