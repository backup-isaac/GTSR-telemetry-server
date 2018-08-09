package computations

import (
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

// LeftBusPower is the power of the high voltage bus
type LeftBusPower struct {
	standardComputation
}

// NewLeftBusPower returns an initialized BusPower
func NewLeftBusPower() *LeftBusPower {
	return &LeftBusPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Left_Bus_Voltage", "Left_Bus_Current"},
		},
	}
}

// Compute computes the bus power as the product of the bus voltage
// and the bus current
func (bp *LeftBusPower) Compute() *datatypes.Datapoint {
	bp.Lock()
	defer bp.Unlock()
	val := bp.values["Left_Bus_Voltage"] * bp.values["Left_Bus_Current"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Left_Bus_Power",
		Value:  val,
	}
}

// RightBusPower is the power of the high voltage bus
type RightBusPower struct {
	standardComputation
}

// NewRightBusPower returns an initialized BusPower
func NewRightBusPower() *RightBusPower {
	return &RightBusPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Right_Bus_Voltage", "Right_Bus_Current"},
		},
	}
}

// Compute computes the bus power as the product of the bus voltage
// and the bus current
func (bp *RightBusPower) Compute() *datatypes.Datapoint {
	bp.Lock()
	defer bp.Unlock()
	val := bp.values["Right_Bus_Voltage"] * bp.values["Right_Bus_Current"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Right_Bus_Power",
		Value:  val,
	}
}

func init() {
	lbp := NewLeftBusPower()
	Register(lbp, lbp.fields)
	rbp := NewRightBusPower()
	Register(rbp, rbp.fields)
}
