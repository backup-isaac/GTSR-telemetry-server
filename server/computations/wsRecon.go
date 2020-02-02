package computations

import (
	"server/datatypes"
	"server/recontool"
)

// MotorControllerEfficiency computes motor controller efficiency
// based on phase current, bus voltage, and bus power
type MotorControllerEfficiency struct {
	phaseCCurrent *datatypes.Datapoint
	busVoltage    *datatypes.Datapoint
	busPower      *datatypes.Datapoint
}

// NewMotorControllerEfficiency returns an initialized MotorControllerEfficiency
func NewMotorControllerEfficiency() *MotorControllerEfficiency {
	return &MotorControllerEfficiency{}
}

// GetMetrics returns the MotorControllerEfficiency's metrics
func (e *MotorControllerEfficiency) GetMetrics() []string {
	return []string{"Phase_C_Current", "Average_Bus_Voltage", "Bus_Power"}
}

// Update signifies an update when all required metrics have been received
// A point with zero bus power will cause everything to be thrown out
func (e *MotorControllerEfficiency) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "Phase_C_Current":
		e.phaseCCurrent = point
	case "Average_Bus_Voltage":
		e.busVoltage = point
	case "Bus_Power":
		if point.Value == 0 {
			e.phaseCCurrent = nil
			e.busVoltage = nil
			e.busPower = nil
			return false
		}
		e.busPower = point
	}
	return e.phaseCCurrent != nil && e.busVoltage != nil && e.busPower != nil
}

// Compute returns the motor controller efficiency
func (e *MotorControllerEfficiency) Compute() *datatypes.Datapoint {
	latest := e.phaseCCurrent.Time
	if e.busVoltage.Time.After(latest) {
		latest = e.busVoltage.Time
	}
	if e.busPower.Time.After(latest) {
		latest = e.busPower.Time
	}
	phaseCurrent := e.phaseCCurrent.Value
	busVoltage := e.busVoltage.Value
	busPower := e.busPower.Value
	e.phaseCCurrent = nil
	e.busVoltage = nil
	e.busPower = nil
	return &datatypes.Datapoint{
		Metric: "Motor_Controller_Efficiency",
		Value:  recontool.MotorControllerEfficiency(phaseCurrent, busVoltage, busPower),
		Time:   latest,
	}
}

func init() {
	Register(NewLeftRightAverage("Bus_Voltage"))
	Register(NewMotorControllerEfficiency())
}
