package computations

import (
	"math"
	"server/datatypes"
	"server/recontool"
)

// TerrainAngle computes the angle of the terrain that the vehicle is driving on
// by deriving the amount of gravitational force that is accelerating the vehicle
type TerrainAngle struct {
	torque       *datatypes.Datapoint
	velocity     *datatypes.Datapoint
	acceleration *datatypes.Datapoint
}

// NewTerrainAngle returns an initialized TerrainAngle
func NewTerrainAngle() *TerrainAngle {
	return &TerrainAngle{}
}

// GetMetrics returns the TerrainAngle's metrics
func (t *TerrainAngle) GetMetrics() []string {
	return []string{"RPM_Derived_Torque", "RPM_Derived_Velocity", "RPM_Derived_Acceleration"}
}

// Update signifies an update when all three required metrics have been received
func (t *TerrainAngle) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "RPM_Derived_Torque":
		t.torque = point
	case "RPM_Derived_Velocity":
		t.velocity = point
	case "RPM_Derived_Acceleration":
		t.acceleration = point
	}
	if t.torque != nil && t.velocity != nil && t.acceleration != nil {
		if math.IsNaN(recontool.DeriveTerrainAngle(t.torque.Value, t.velocity.Value, t.acceleration.Value, sr3)) {
			t.torque = nil
			t.velocity = nil
			t.acceleration = nil
		} else {
			return true
		}
	}
	return false
}

// Compute returns the terrain angle in radians
func (t *TerrainAngle) Compute() *datatypes.Datapoint {
	latest := t.torque.Time
	if t.velocity.Time.After(latest) {
		latest = t.velocity.Time
	}
	if t.acceleration.Time.After(latest) {
		latest = t.acceleration.Time
	}
	torque := t.torque.Value
	velocity := t.velocity.Value
	accel := t.acceleration.Value
	t.torque = nil
	t.velocity = nil
	t.acceleration = nil
	return &datatypes.Datapoint{
		Metric: "Terrain_Angle",
		Value:  recontool.DeriveTerrainAngle(torque, velocity, accel, sr3),
		Time:   latest,
	}
}

func init() {
	Register(NewTerrainAngle())
}
