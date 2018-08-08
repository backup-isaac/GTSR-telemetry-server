package computations

import (
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

// BusPower is the power of the high voltage bus
type BusPower struct {
	standardComputation
}

// NewBusPower returns an initialized BusPower
func NewBusPower() *BusPower {
	return &BusPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Bus_Voltage", "Bus_Current"},
		},
	}
}

// Compute computes the bus power as the product of the bus voltage
// and the bus current
func (bp *BusPower) Compute() *datatypes.Datapoint {
	bp.Lock()
	defer bp.Unlock()
	val := bp.values["Bus_Voltage"] * bp.values["Bus_Current"]
	bp.values = make(map[string]float64)
	return &datatypes.Datapoint{
		Metric: "Bus_Power",
		Value:  val,
	}
}

func init() {
	bp := NewBusPower()
	Register(bp, bp.fields)
}
