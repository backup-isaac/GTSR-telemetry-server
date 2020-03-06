package computations

import (
	"math"
	"server/datatypes"
	"server/recontool"
)

// TerrainAngle computes the angle of the terrain that the vehicle is driving on
// by deriving the amount of gravitational force that is accelerating the vehicle
type TerrainAngle struct {
	standardComputation
}

// NewTerrainAngle returns an initialized TerrainAngle
func NewTerrainAngle() *TerrainAngle {
	return &TerrainAngle{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"RPM_Derived_Torque", "RPM_Derived_Velocity", "RPM_Derived_Acceleration"},
		},
	}
}

// Update signifies an update when all three required metrics have been received
func (t *TerrainAngle) Update(point *datatypes.Datapoint) bool {
	t.values[point.Metric] = point.Value
	if point.Time.After(t.timestamp) {
		t.timestamp = point.Time
	}
	if len(t.values) >= len(t.fields) {
		if math.IsNaN(recontool.DeriveTerrainAngle(
			t.values["RPM_Derived_Torque"],
			t.values["RPM_Derived_Velocity"],
			t.values["RPM_Derived_Acceleration"],
			sr3),
		) {
			t.values = make(map[string]float64)
		} else {
			return true
		}
	}
	return false
}

// Compute returns the terrain angle in radians
func (t *TerrainAngle) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Terrain_Angle",
		Value: recontool.DeriveTerrainAngle(
			t.values["RPM_Derived_Torque"],
			t.values["RPM_Derived_Velocity"],
			t.values["RPM_Derived_Acceleration"],
			sr3),
		Time: t.timestamp,
	}
	t.values = make(map[string]float64)
	return datapoint
}

func init() {
	Register(NewTerrainAngle())
}
