package computations

import (
	"server/datatypes"
	"server/recontool"
	"testing"
)

func TestTerrainAngle(t *testing.T) {
	a := NewTerrainAngle()
	computationRunner(t, a, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Acceleration", 4),
		makeDatapoint("RPM_Derived_Acceleration", 2),
		makeDatapoint("RPM_Derived_Velocity", 0),
		makeDatapoint("RPM_Derived_Torque", 12),
	}, &datatypes.Datapoint{
		Metric: "Terrain_Angle",
		Value:  recontool.DeriveTerrainAngle(12, 0, 2, sr3),
		Time:   pointTime,
	})
	computationRunner(t, a, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 0),
		makeDatapoint("RPM_Derived_Acceleration", 0),
		makeDatapoint("RPM_Derived_Torque", 10000),
		makeDatapoint("RPM_Derived_Torque", 10),
		makeDatapoint("RPM_Derived_Velocity", 20),
		makeDatapoint("RPM_Derived_Acceleration", 3),
	}, &datatypes.Datapoint{
		Metric: "Terrain_Angle",
		Value:  recontool.DeriveTerrainAngle(10, 20, 3, sr3),
		Time:   pointTime,
	})
}
