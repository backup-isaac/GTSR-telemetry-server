package track

import (
	"os"
	"server/datatypes"
	"server/listener"
	"testing"
	"time"
)

type pointNumPair struct {
	point *datatypes.RoutePoint
	num   int
}

type fakePointUploader struct {
	tracks chan int
	points chan *pointNumPair
}

func (f *fakePointUploader) UploadNewRoute(size int) {
	f.tracks <- size
}

func (f *fakePointUploader) UploadTCPPointMessage(point *datatypes.RoutePoint, num int) {
	f.points <- &pointNumPair{
		point: point,
		num:   num,
	}
}

func TestTrackUploader(t *testing.T) {
	infoPath = "track_uploader_TEST_config.json"
	defer os.Remove(infoPath)
	model := &Model{
		IsTrackInfoNew:      false,
		IsTrackInfoUploaded: true,
		PointNumber:         0,
	}
	err := model.Commit()
	if err != nil {
		t.Fatalf("Error commiting initial model: %+v", err)
	}
	routePath = "track_uploader_TEST_route.json"
	defer os.Remove(routePath)
	messenger := &fakePointUploader{
		tracks: make(chan int, 10),
		points: make(chan *pointNumPair, 10),
	}
	afterChan := make(chan time.Time, 1)
	// Lets us control the implementation of time.After
	after = func(time.Duration) <-chan time.Time {
		return afterChan
	}
	publisher := listener.GetDatapointPublisher()
	defer publisher.Close()
	track := &Track{
		newPoints: make(chan []*datatypes.RoutePoint, 10),
		model:     model,
		messenger: messenger,
		done:      make(chan bool),
	}
	defer track.Close()
	go track.uploader()
	// Assert no messages sent on startup
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	// Upload track
	route := []*datatypes.RoutePoint{{
		Speed:    0,
		Critical: false,
	}, {
		Speed:    1,
		Critical: true,
	}, {
		Speed:    2,
		Critical: true,
	}}
	err = track.UploadRoute(route)
	if err != nil {
		t.Fatalf("Error uploading track points: %+v", err)
	}
	// Assert no points yet since car isn't connected yet
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	// Send 0 connection status to ensure upload doesn't start
	point := &datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  0,
	}
	publisher.Publish(point)
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	// Connect car and assert that upload process starts
	publisher.Publish(&datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  1,
	})
	select {
	case trackLen := <-messenger.tracks:
		if trackLen != 2 {
			t.Errorf("Unexpected track length: want %q, got %q", 2, trackLen)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected track message")
	}
	// Test retries
	afterChan <- time.Now()
	select {
	case trackLen := <-messenger.tracks:
		if trackLen != 2 {
			t.Errorf("Unexpected track length: want %q, got %q", 2, trackLen)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected track message")
	}
	// ACK track info
	publisher.Publish(&datatypes.Datapoint{
		Metric: newTrackInfoACK,
		Value:  0,
	})
	// Expect first route point
	select {
	case point := <-messenger.points:
		if point.num != 0 {
			t.Errorf("Wrong point number: want %q, got %q", 0, point.num)
		}
		if point.point.Speed != 1 {
			t.Errorf("Incorrect speed in received point: want %q, got %q", 1, int(point.point.Speed))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	// Test retrying a route point
	afterChan <- time.Now()
	select {
	case <-messenger.points:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	// Test ACKing a packet
	publisher.Publish(&datatypes.Datapoint{
		Metric: packetACK,
		Value:  0,
	})
	select {
	case point := <-messenger.points:
		if point.num != 1 {
			t.Errorf("Wrong point number: want %q, got %q", 1, point.num)
		}
		if point.point.Speed != 2 {
			t.Errorf("Incorrect speed in received point: want %q, got %q", 2, int(point.point.Speed))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	// Test losing connection
	publisher.Publish(&datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  0,
	})
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	publisher.Publish(&datatypes.Datapoint{
		Metric: packetACK,
		Value:  1,
	})
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	// Reconnect
	publisher.Publish(&datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  1,
	})
	select {
	case point := <-messenger.points:
		if point.num != 1 {
			t.Errorf("Wrong point number: want %q, got %q", 1, point.num)
		}
		if point.point.Speed != 2 {
			t.Errorf("Incorrect speed in received point: want %q, got %q", 2, int(point.point.Speed))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	// Test wrong ACK
	publisher.Publish(&datatypes.Datapoint{
		Metric: packetACK,
		Value:  0,
	})
	select {
	case point := <-messenger.points:
		if point.num != 1 {
			t.Errorf("Wrong point number: want %q, got %q", 1, point.num)
		}
		if point.point.Speed != 2 {
			t.Errorf("Incorrect speed in received point: want %q, got %q", 2, int(point.point.Speed))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	// Test new track upload during old
	track.UploadRoute([]*datatypes.RoutePoint{{
		Speed:    1,
		Critical: true,
	}})
	select {
	case trackLen := <-messenger.tracks:
		if trackLen != 1 {
			t.Errorf("Unexpected track length: want %q, got %q", 1, trackLen)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected track message")
	}
	// ACK track sequence
	publisher.Publish(&datatypes.Datapoint{
		Metric: newTrackInfoACK,
		Value:  0,
	})
	select {
	case point := <-messenger.points:
		if point.num != 0 {
			t.Errorf("Wrong point number: want %q, got %q", 0, point.num)
		}
		if point.point.Speed != 1 {
			t.Errorf("Incorrect speed in received point: want %q, got %q", 1, int(point.point.Speed))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected point")
	}
	publisher.Publish(&datatypes.Datapoint{
		Metric: packetACK,
		Value:  0,
	})
	// Ensure uploadPoints has exited
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	afterChan <- time.Now()
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	<-afterChan
	// Restart connection with no track uploaded
	publisher.Publish(&datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  0,
	})
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
	publisher.Publish(&datatypes.Datapoint{
		Metric: "Connection_Status",
		Value:  1,
	})
	<-time.After(100 * time.Millisecond)
	messenger.assertNoUploads(t)
}

func (f *fakePointUploader) assertNoUploads(t *testing.T) {
	select {
	case <-f.tracks:
		t.Error("Unexpected track message on uploader startup")
	default:
	}
	select {
	case <-f.points:
		t.Error("Unexpected route point on uploader startup")
	default:
	}
}
