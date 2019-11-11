package track

import (
	"log"
	"server/datatypes"
	"server/listener"
	"server/message"
	"time"
)

const (
	newTrackInfoACK = "Track_Info_Control_Begin_ACK"
	packetACK       = "Track_Info_Control_Packet_ACK"
	timeout         = 3 * time.Second
)

// Track controls the logic for uploading new routes to the car
type Track struct {
	newPoints chan []*datatypes.RoutePoint
	model     *Model
	messenger *message.CarMessenger
}

// NewTrack initializes a new Track object with default values
func NewTrack(messenger *message.CarMessenger) (*Track, error) {
	m, err := ReadTrackInfoModel()
	if err != nil {
		return nil, err
	}
	t := &Track{
		newPoints: make(chan []*datatypes.RoutePoint, 1),
		model:     m,
		messenger: messenger,
	}
	go t.uploader()
	return t, nil
}

// UploadRoute saves the provided route and uploads it to the car
func (t *Track) UploadRoute(route []*datatypes.RoutePoint) error {
	err := putRoute(route)
	if err != nil {
		return err
	}
	t.newPoints <- filterCritical(route)
	return nil
}

// GetRoute returns the current saved route
func (t *Track) GetRoute() ([]*datatypes.RoutePoint, error) {
	return getRoute()
}

func (t *Track) uploader() {
	status := make(chan *datatypes.Datapoint, 10)
	err := listener.Subscribe(status, "Connection_Status")
	if err != nil {
		log.Fatalf("error subscribing to connection status: %+v", err)
	}
	track, err := getRoute()
	if err == nil {
		track = filterCritical(track)
	} else {
		t.model.IsTrackInfoNew = false
		t.model.IsTrackInfoUploaded = true
		t.model.Commit()
	}
	var quit chan bool
	connected := false
	for {
		// Listen for status frames or new routes
		select {
		case point := <-status:
			if point.Value == 0 {
				// If connection was lost, quit trying to upload
				connected = false
				if quit != nil {
					quit <- true
					quit = nil
				}
			} else {
				// If connection was established and the current route isn't done uploading,
				// continue the upload process
				connected = true
				quit = make(chan bool)
				if !t.model.IsTrackInfoUploaded {
					go t.uploadPoints(track, quit)
				}
			}
		case track = <-t.newPoints:
			// When we get a new route, cancel the current upload attempts and start
			// a new one
			if quit != nil {
				quit <- true
			}
			t.model.IsTrackInfoNew = true
			t.model.PointNumber = 0
			t.model.IsTrackInfoUploaded = false
			t.model.Commit()
			if connected {
				quit = make(chan bool)
				go t.uploadPoints(track, quit)
			}
		}
	}
}

func (t *Track) uploadPoints(track []*datatypes.RoutePoint, quit chan bool) {
	c := make(chan *datatypes.Datapoint, 10)
	listener.Subscribe(c, newTrackInfoACK, packetACK)
	for !t.model.IsTrackInfoUploaded {
		// Upload the relevant packet for our current state
		if t.model.IsTrackInfoNew {
			t.messenger.UploadNewRoute(len(track))
		} else {
			t.messenger.UploadTCPPointMessage(track[t.model.PointNumber], t.model.PointNumber)
		}
		// Update state based on the car's response
		select {
		case ack := <-c:
			// If we get an ACK corresponding to our current state, move to the next state
			switch ack.Metric {
			case newTrackInfoACK:
				if t.model.IsTrackInfoNew {
					t.model.IsTrackInfoNew = false
					t.model.Commit()
				}
			case packetACK:
				if int(ack.Value) == t.model.PointNumber {
					t.model.PointNumber++
					if t.model.PointNumber >= len(track) {
						t.model.IsTrackInfoUploaded = true
					}
					t.model.Commit()
				}
			default:
				log.Printf("Error: metric %q should not have been received in uploadPoints", ack.Metric)
			}
		case <-time.After(timeout):
			// Retry current send after 3 seconds
		case <-quit:
			// Car lost connection or new track was uploaded
			return
		}
	}
}
