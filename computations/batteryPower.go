package computations

import (
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

// BatteryPower is the power being output by the pack
type BatteryPower struct {
	standardComputation
}

// NewBatteryPower returns a BatteryPower initialized with the proper values
func NewBatteryPower() *BatteryPower {
	return &BatteryPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"BMS_Current", "Pack_Voltage", "Left_Bus_Voltage", "Right_Bus_Voltage"},
		},
	}
}

// Compute computes the battery power as the product of the BMS current
// and the bus voltage if the bus voltage value is nominal; otherwise,
// the pack voltage measurement is used
func (bp *BatteryPower) Compute() *datatypes.Datapoint {
	bp.Lock()
	defer bp.Unlock()
	bmsCurrent := bp.values["BMS_Current"]
	packVoltage := bp.values["Pack_Voltage"]
	busVoltage := (bp.values["Left_Bus_Voltage"] + bp.values["Right_Bus_Voltage"]) / 2
	point := &datatypes.Datapoint{
		Metric: "Battery_Power",
	}
	if busVoltage < 50.0 {
		point.Value = bmsCurrent * packVoltage
	} else {
		point.Value = bmsCurrent * busVoltage
	}
	bp.values = make(map[string]float64)
	return point
}

func init() {
	bp := NewBatteryPower()
	Register(bp, bp.fields)
}
