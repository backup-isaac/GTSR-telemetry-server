package track

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"server/datatypes"
	"sync"
)

var routePath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	routePath = path.Join(dir, "route.json")
}

var routeMutex sync.Mutex

func getRoute() ([]*datatypes.RoutePoint, error) {
	routeMutex.Lock()
	defer routeMutex.Unlock()
	bytes, err := ioutil.ReadFile(routePath)
	if err != nil {
		return nil, err
	}
	var route []*datatypes.RoutePoint
	err = json.Unmarshal(bytes, &route)
	return route, err
}

func putRoute(route []*datatypes.RoutePoint) error {
	routeMutex.Lock()
	defer routeMutex.Unlock()
	bytes, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf("error marshalling route: %w", err)
	}
	err = ioutil.WriteFile(routePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing route: %w", err)
	}
	return nil
}

func filterCritical(route []*datatypes.RoutePoint) []*datatypes.RoutePoint {
	var criticalPoints []*datatypes.RoutePoint
	for _, point := range route {
		if point.Critical {
			criticalPoints = append(criticalPoints, point)
		}
	}
	return criticalPoints
}
