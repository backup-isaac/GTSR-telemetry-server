package listener

import (
	"fmt"
	"sync"
)

// DatapointPublisher allows threads to subscribe to a particular publisher,
// in this case the tcp port listener
type DatapointPublisher struct {
	PublishChannel  chan *Datapoint
	Subscribers     []chan *Datapoint
	SubscribersLock *sync.Mutex
}

var publisher *DatapointPublisher

// Subscribe will add a channel to the list of Subscribers
func Subscribe(c chan *Datapoint) error {
	if publisher == nil {
		return fmt.Errorf("Subscribe called before Listen")
	}
	publisher.SubscribersLock.Lock()
	defer publisher.SubscribersLock.Unlock()
	publisher.Subscribers = append(publisher.Subscribers, c)
	return nil
}

// Unsubscribe will remove a channel from the list of Subscribers
func Unsubscribe(c chan *Datapoint) error {
	if publisher == nil {
		return fmt.Errorf("Unsubscribe called before Listen")
	}
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
func Publish(point *Datapoint) {
	publisher.PublishChannel <- point
}

func publisherThread() {
	for {
		point := <-publisher.PublishChannel
		publisher.SubscribersLock.Lock()
		for _, subscriber := range publisher.Subscribers {
			subscriber <- point
		}
		publisher.SubscribersLock.Unlock()
	}
}
