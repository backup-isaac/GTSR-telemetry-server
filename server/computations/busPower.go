package computations

import (
	"server/datatypes"
)

// BusPower is the total bus power across both the left and right busses
type BusPower struct {
	standardComputation
}

// NewBusPower returns an initialized BusPower
func NewBusPower() *BusPower {
	return &BusPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Left_Bus_Power", "Right_Bus_Power"},
		},
	}
}

// Compute computes the bus power as the sum of the left and right bus powers
func (bp *BusPower) Compute() *datatypes.Datapoint {
	val := bp.values["Left_Bus_Power"] + bp.values["Right_Bus_Power"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Bus_Power",
		Value:  val,
	}
}

// LeftBusPower is the power of the left high voltage bus
type LeftBusPower struct {
	standardComputation
}

// NewLeftBusPower returns an initialized LeftBusPower
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
	val := bp.values["Left_Bus_Voltage"] * bp.values["Left_Bus_Current"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Left_Bus_Power",
		Value:  val,
	}
}

// RightBusPower is the power of the right high voltage bus
type RightBusPower struct {
	standardComputation
}

// NewRightBusPower returns an initialized RightBusPower
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
	val := bp.values["Right_Bus_Voltage"] * bp.values["Right_Bus_Current"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Right_Bus_Power",
		Value:  val,
	}
}

func init() {
	Register(NewBusPower())
	Register(NewLeftBusPower())
	Register(NewRightBusPower())
}
