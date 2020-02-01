package track

import (
	"os"
	"reflect"
	"testing"

	"server/datatypes"
)

func TestPutGetRoute(t *testing.T) {
	routePath = "route_TEST.json"
	defer os.Remove(routePath)
	route := []*datatypes.RoutePoint{{
		Distance:  1,
		Latitude:  1,
		Longitude: 1,
		Speed:     1,
	}, {
		Distance:  2,
		Latitude:  2,
		Longitude: 2,
		Speed:     2,
	}, {
		Distance:  3,
		Latitude:  3,
		Longitude: 3,
		Speed:     3,
	}}
	err := putRoute(route)
	if err != nil {
		t.Fatalf("Error saving route: %+v", err)
	}
	readRoute, err := getRoute()
	if err != nil {
		t.Fatalf("Error reading route: %+v", err)
	}
	if !reflect.DeepEqual(route, readRoute) {
		t.Errorf("Read route does not match saved: want %+v, got %+v", route, readRoute)
	}
}

func TestFilterCritical(t *testing.T) {
	route := []*datatypes.RoutePoint{{
		Speed:    0,
		Critical: true,
	}, {
		Speed:    1,
		Critical: false,
	}, {
		Speed:    2,
		Critical: true,
	}, {
		Speed:    3,
		Critical: false,
	}}
	critical := filterCritical(route)
	if len(critical) != 2 {
		t.Fatalf("Unexpected count of critical points: want %+v, got %+v", 2, len(critical))
	}
	if critical[0].Speed != 0 {
		t.Errorf("Unexpected speed in critical points: want %+v, got %+v", 0, critical[0].Speed)
	}
	if critical[1].Speed != 2 {
		t.Errorf("Unexpected speed in critical points: want %+v, got %+v", 2, critical[1].Speed)
	}
}
