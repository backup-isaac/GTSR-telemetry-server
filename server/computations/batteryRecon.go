package computations

import (
	"fmt"
	"server/datatypes"
	"server/recontool"
	"strconv"
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

// PackEfficiency computes the efficiency of the battery pack's high voltage bus
type PackEfficiency struct {
	busCurrent     *datatypes.Datapoint
	busPower       *datatypes.Datapoint
	packResistance *datatypes.Datapoint
}

// NewPackEfficiency returns an initialized PackEfficiency
func NewPackEfficiency() *PackEfficiency {
	return &PackEfficiency{}
}

// GetMetrics returns the PackEfficiency's metrics
func (e *PackEfficiency) GetMetrics() []string {
	return []string{"Bus_Current", "Bus_Power", "Pack_Resistance"}
}

// Update signifies an update when all required metrics have been received
func (e *PackEfficiency) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "Bus_Current":
		e.busCurrent = point
	case "Bus_Power":
		e.busPower = point
	case "Pack_Resistance":
		e.packResistance = point
	}
	return e.busCurrent != nil && e.busPower != nil && e.packResistance != nil
}

// Compute returns the pack's efficiency
func (e *PackEfficiency) Compute() *datatypes.Datapoint {
	latest := e.busCurrent.Time
	if e.busPower.Time.After(latest) {
		latest = e.busPower.Time
	}
	if e.packResistance.Time.After(latest) {
		latest = e.packResistance.Time
	}
	iBus := e.busCurrent.Value
	pBus := e.busPower.Value
	rPack := e.packResistance.Value
	e.busCurrent = nil
	e.busPower = nil
	e.packResistance = nil
	return &datatypes.Datapoint{
		Metric: "Pack_Efficiency",
		Value:  recontool.PackEfficiency(iBus, pBus, rPack),
		Time:   latest,
	}
}

// MinModuleVoltage calculates the battery module with the minimum voltage
type MinModuleVoltage struct {
	moduleVoltages []float64
	time           time.Time
	size           uint
}

// NewMinModuleVoltage returns an initialized MinModuleVoltage
func NewMinModuleVoltage() *MinModuleVoltage {
	return &MinModuleVoltage{
		moduleVoltages: make([]float64, sr3.VSer),
		size:           0,
	}
}

// GetMetrics returns the MinModuleVoltage's metrics
func (m *MinModuleVoltage) GetMetrics() []string {
	metrics := make([]string, sr3.VSer)
	var i uint
	for i = 0; i < sr3.VSer; i++ {
		metrics[i] = fmt.Sprintf("Cell_Voltage_%d", i+1)
	}
	return metrics
}

// Update signifies an update when all required metrics have been received
func (m *MinModuleVoltage) Update(point *datatypes.Datapoint) bool {
	ind, err := strconv.ParseUint(point.Metric[13:], 10, 32)
	if err != nil {
		return false
	}
	if m.moduleVoltages[ind-1] == 0 {
		m.size++
	}
	m.moduleVoltages[ind-1] = point.Value
	m.time = point.Time
	return m.size == sr3.VSer
}

// Compute finds the module with the lowest voltage
func (m *MinModuleVoltage) Compute() *datatypes.Datapoint {
	time := m.time
	var argmin uint
	var i uint
	for i = 1; i < sr3.VSer; i++ {
		if m.moduleVoltages[i] < m.moduleVoltages[argmin] {
			argmin = i
		}
	}
	m.size = 0
	m.moduleVoltages = make([]float64, sr3.VSer)
	return &datatypes.Datapoint{
		Metric: "Min_Cell_Voltage",
		Value:  float64(argmin + 1),
		Time:   time,
	}
}

// MaxModuleVoltage calculates the battery module with the maximum voltage
type MaxModuleVoltage struct {
	moduleVoltages []float64
	time           time.Time
	size           uint
}

// NewMaxModuleVoltage returns an initialized MaxModuleVoltage
func NewMaxModuleVoltage() *MaxModuleVoltage {
	return &MaxModuleVoltage{
		moduleVoltages: make([]float64, sr3.VSer),
		size:           0,
	}
}

// GetMetrics returns the MaxModuleVoltage's metrics
func (m *MaxModuleVoltage) GetMetrics() []string {
	metrics := make([]string, sr3.VSer)
	var i uint
	for i = 0; i < sr3.VSer; i++ {
		metrics[i] = fmt.Sprintf("Cell_Voltage_%d", i+1)
	}
	return metrics
}

// Update signifies an update when all required metrics have been received
func (m *MaxModuleVoltage) Update(point *datatypes.Datapoint) bool {
	ind, err := strconv.ParseUint(point.Metric[13:], 10, 32)
	if err != nil {
		return false
	}
	if m.moduleVoltages[ind-1] == 0 {
		m.size++
	}
	m.moduleVoltages[ind-1] = point.Value
	m.time = point.Time
	return m.size == sr3.VSer
}

// Compute finds the module with the highest voltage
func (m *MaxModuleVoltage) Compute() *datatypes.Datapoint {
	time := m.time
	var argmax uint
	var i uint
	for i = 1; i < sr3.VSer; i++ {
		if m.moduleVoltages[i] > m.moduleVoltages[argmax] {
			argmax = i
		}
	}
	m.size = 0
	m.moduleVoltages = make([]float64, sr3.VSer)
	return &datatypes.Datapoint{
		Metric: "Max_Cell_Voltage",
		Value:  float64(argmax + 1),
		Time:   time,
	}
}

// ModuleVoltageImbalance calculates ∆maxModuleVoltage - ∆minModuleVoltage
type ModuleVoltageImbalance struct {
	moduleVoltages []float64
	time           time.Time
	size           uint
}

// NewModuleVoltageImbalance returns an initialized ModuleVoltageImbalance
func NewModuleVoltageImbalance() *ModuleVoltageImbalance {
	return &ModuleVoltageImbalance{
		moduleVoltages: make([]float64, sr3.VSer),
		size:           0,
	}
}

// GetMetrics returns the ModuleVoltageImbalance's metrics
func (m *ModuleVoltageImbalance) GetMetrics() []string {
	metrics := make([]string, sr3.VSer)
	var i uint
	for i = 0; i < sr3.VSer; i++ {
		metrics[i] = fmt.Sprintf("Cell_Voltage_%d", i+1)
	}
	return metrics
}

// Update signifies an update when all required metrics have been received
func (m *ModuleVoltageImbalance) Update(point *datatypes.Datapoint) bool {
	ind, err := strconv.ParseUint(point.Metric[13:], 10, 32)
	if err != nil {
		return false
	}
	if m.moduleVoltages[ind-1] == 0 {
		m.size++
	}
	m.moduleVoltages[ind-1] = point.Value
	m.time = point.Time
	return m.size == sr3.VSer
}

// Compute returns the imbalance of the battery pack
func (m *ModuleVoltageImbalance) Compute() *datatypes.Datapoint {
	time := m.time
	max := m.moduleVoltages[0]
	min := m.moduleVoltages[0]
	var i uint
	for i = 1; i < sr3.VSer; i++ {
		if m.moduleVoltages[i] > max {
			max = m.moduleVoltages[i]
		} else if m.moduleVoltages[i] < min {
			min = m.moduleVoltages[i]
		}
	}
	m.size = 0
	m.moduleVoltages = make([]float64, sr3.VSer)
	return &datatypes.Datapoint{
		Metric: "Cell_Voltage_Imbalance",
		Value:  max - min,
		Time:   time,
	}
}

func init() {
	InitSr3()
	Register(NewPackResistance())
	Register(NewPackEfficiency())
	Register(NewMinModuleVoltage())
	Register(NewMaxModuleVoltage())
	Register(NewModuleVoltageImbalance())
}
