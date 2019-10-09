package listener

import (
	"fmt"
	"sync"

	"server/datatypes"
)

// DatapointPublisher provides pub/sub functionality for Datapoint streams
type DatapointPublisher struct {
	publishChannel      chan *datatypes.Datapoint
	specificSubscribers map[string][]chan *datatypes.Datapoint
	subscribers         []chan *datatypes.Datapoint
	subscribersLock     *sync.Mutex
}

func newDatapointPublisher() *DatapointPublisher {
	publisher := &DatapointPublisher{
		publishChannel:      make(chan *datatypes.Datapoint, 10000),
		specificSubscribers: make(map[string][]chan *datatypes.Datapoint),
		subscribers:         []chan *datatypes.Datapoint{},
		subscribersLock:     new(sync.Mutex),
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
// called on this publisher, the datapoint will be sent to the provided channel.
// If specific metrics are provided, the channel will only be sent datapoints
// for those specific metrics
func (publisher *DatapointPublisher) Subscribe(c chan *datatypes.Datapoint, metrics ...string) error {
	publisher.subscribersLock.Lock()
	defer publisher.subscribersLock.Unlock()
	if len(metrics) == 0 {
		publisher.subscribers = append(publisher.subscribers, c)
	}
	for _, metric := range metrics {
		publisher.specificSubscribers[metric] = append(publisher.specificSubscribers[metric], c)
	}
	return nil
}

// Unsubscribe unsubscribes a channel from the publisher. Datapoints published
// to the publisher will no longer be sent to the provided channel
func (publisher *DatapointPublisher) Unsubscribe(c chan *datatypes.Datapoint) error {
	publisher.subscribersLock.Lock()
	defer publisher.subscribersLock.Unlock()
	found := false
	for i, channel := range publisher.subscribers {
		if c == channel {
			publisher.subscribers = append(publisher.subscribers[:i], publisher.subscribers[i+1:]...)
			found = true
		}
	}
	for k := range publisher.specificSubscribers {
		for i, channel := range publisher.specificSubscribers[k] {
			if c == channel {
				publisher.specificSubscribers[k] = append(publisher.specificSubscribers[k][:i], publisher.specificSubscribers[k][i+1:]...)
				found = true
			}
		}
	}
	if !found {
		return fmt.Errorf("Unsubscribe: channel not found in subscriber list")
	}
	return nil
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
	closed := make(map[chan *datatypes.Datapoint]bool)
	for _, subscribers := range publisher.specificSubscribers {
		for _, c := range subscribers {
			if !closed[c] {
				close(c)
				closed[c] = true
			}
		}
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
		for _, subscriber := range publisher.specificSubscribers[point.Metric] {
			subscriber <- point
		}
		publisher.subscribersLock.Unlock()
	}
}
