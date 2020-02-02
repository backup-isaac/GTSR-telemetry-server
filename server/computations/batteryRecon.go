package computations

import (
	"server/datatypes"
	"server/recontool"
	"time"
)

// PackResistance is the battery pack's resistance computed as
// ∂V/∂I
type PackResistance struct {
	packVoltages  []float64
	packCurrents  []float64
	time          time.Time
	rejectCurrent bool
}

// NewPackResistance returns an initialized PackResistance
func NewPackResistance() *PackResistance {
	return &PackResistance{
		packVoltages:  make([]float64, 0, 4096),
		packCurrents:  make([]float64, 0, 4096),
		rejectCurrent: false,
	}
}

// GetMetrics returns the PackResistance's metrics
func (r *PackResistance) GetMetrics() []string {
	return []string{"Average_Bus_Voltage", "BMS_Current"}
}

// Update signifies an update when there are an equal amount of voltages
// and currents received, and there are at least two of each.
// A point is thrown out if more of it have been received than the other
// metric. A point is also thrown out if it corresponds to a bus voltage
// lower than 1, to avoid bias in the regression
func (r *PackResistance) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Average_Bus_Voltage" {
		r.rejectCurrent = point.Value < 1
		if r.rejectCurrent || len(r.packCurrents) < len(r.packVoltages) {
			return false
		}
		r.packVoltages = append(r.packVoltages, point.Value)
	} else if r.rejectCurrent || len(r.packVoltages) < len(r.packCurrents) {
		return false
	} else {
		r.packCurrents = append(r.packCurrents, point.Value)
	}
	r.time = point.Time
	return len(r.packCurrents) > 2 && len(r.packCurrents) == len(r.packVoltages)
}

// Compute returns the pack's resistance in ohms
func (r *PackResistance) Compute() *datatypes.Datapoint {
	resistance := recontool.PackResistanceUnfiltered(r.packCurrents, r.packVoltages)
	if len(r.packCurrents) == 4096 {
		r.packCurrents = r.packCurrents[1:]
		r.packVoltages = r.packVoltages[1:]
	}
	return &datatypes.Datapoint{
		Metric: "Pack_Resistance",
		Value:  resistance,
		Time:   r.time,
	}
}

func init() {
	Register(NewPackResistance())
}
