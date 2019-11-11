package datatypes

// RoutePoint is a point along the uploaded route
type RoutePoint struct {
	// Distance is the distance along the route for this point
	Distance float64 `json:"distance"`
	// Latitude is the GPS latitude of this point
	Latitude float64 `json:"latitude"`
	// Longitude is the GPS longitude of this point
	Longitude float64 `json:"longitude"`
	// Speed is the suggested speed for the car at this point
	Speed float64 `json:"speed"`
	// Critical is a flag for whether this is a significant datapoint
	// that should be sent to the car to be suggested to the driver
	Critical bool `json:"critical"`
}
