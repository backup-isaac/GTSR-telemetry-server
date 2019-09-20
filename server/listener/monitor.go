package listener

import (
	"server/datatypes"
	"time"
)

var connStatusMetric = "Connection_Status"

// monitorConnection listens for data from the car, posting updates
// to Slack for when connection is established and lost.
func monitorConnection() {
	p := GetDatapointPublisher()
	points := make(chan *datatypes.Datapoint, 10)
	Subscribe(points)
	connected := false
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case point := <-points:
			if point.Metric == connStatusMetric {
				continue
			}
			timer.Stop()
			// Clear potential timer channel contents (nonblocking)
			select {
			case <-timer.C:
			default:
			}
			timer.Reset(10 * time.Second)
			if !connected {
				p.Publish(&datatypes.Datapoint{
					Metric: connStatusMetric,
					Value:  1,
					Time:   time.Now(),
				})
				connected = true
			}
		case <-timer.C:
			timer.Reset(10 * time.Second)
			if connected {
				p.Publish(&datatypes.Datapoint{
					Metric: connStatusMetric,
					Value:  0,
					Time:   time.Now(),
				})
				connected = false
			}
		}
	}
}
