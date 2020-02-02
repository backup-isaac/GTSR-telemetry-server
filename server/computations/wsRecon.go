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

// DrivetrainEfficiency gives the total drivetrain efficiency
type DrivetrainEfficiency struct {
	motorEfficiency           *datatypes.Datapoint
	motorControllerEfficiency *datatypes.Datapoint
	packEfficiency            *datatypes.Datapoint
}

// NewDrivetrainEfficiency returns an initialized DrivetrainEfficiency
func NewDrivetrainEfficiency() *DrivetrainEfficiency {
	return &DrivetrainEfficiency{}
}

// GetMetrics returns the DrivetrainEfficiency's metrics
func (e *DrivetrainEfficiency) GetMetrics() []string {
	return []string{"Motor_Efficiency", "Motor_Controller_Efficiency", "Pack_Efficiency"}
}

// Update signifies an update when all required metrics have been received
func (e *DrivetrainEfficiency) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "Motor_Efficiency":
		e.motorEfficiency = point
	case "Motor_Controller_Efficiency":
		e.motorControllerEfficiency = point
	case "Pack_Efficiency":
		e.packEfficiency = point
	}
	return e.motorControllerEfficiency != nil && e.motorEfficiency != nil && e.packEfficiency != nil
}

// Compute computes drivetrain efficiency
func (e *DrivetrainEfficiency) Compute() *datatypes.Datapoint {
	latest := e.motorEfficiency.Time
	if e.motorControllerEfficiency.Time.After(latest) {
		latest = e.motorControllerEfficiency.Time
	}
	if e.packEfficiency.Time.After(latest) {
		latest = e.packEfficiency.Time
	}
	eMot := e.motorEfficiency.Value
	eMc := e.motorControllerEfficiency.Value
	ePack := e.packEfficiency.Value
	e.motorEfficiency = nil
	e.motorControllerEfficiency = nil
	e.packEfficiency = nil
	return &datatypes.Datapoint{
		Metric: "Drivetrain_Efficiency",
		Value:  recontool.DrivetrainEfficiency(eMc, ePack, eMot),
		Time:   latest,
	}
}

// ModeledBusCurrent derives bus current from modeled power
type ModeledBusCurrent struct {
	power   *datatypes.Datapoint
	voltage *datatypes.Datapoint
}

// NewModeledBusCurrent returns an initialized ModeledBusCurrent
func NewModeledBusCurrent() *ModeledBusCurrent {
	return &ModeledBusCurrent{}
}

// GetMetrics returns the ModeledBusCurrent's metrics
func (c *ModeledBusCurrent) GetMetrics() []string {
	return []string{"Modeled_Motor_Power", "Average_Bus_Voltage"}
}

// Update signifies an update when all required metrics have been received
func (c *ModeledBusCurrent) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Modeled_Motor_Power" {
		c.power = point
	} else if point.Metric == "Average_Bus_Voltage" {
		c.voltage = point
	}
	return c.power != nil && c.voltage != nil
}

// Compute computes modeled bus current in amps
func (c *ModeledBusCurrent) Compute() *datatypes.Datapoint {
	latest := c.power.Time
	if c.voltage.Time.After(latest) {
		latest = c.voltage.Time
	}
	p := c.power.Value
	vBus := c.voltage.Value
	c.power = nil
	c.voltage = nil
	return &datatypes.Datapoint{
		Metric: "Modeled_Bus_Current",
		Value:  p / vBus,
		Time:   latest,
	}
}

func init() {
	Register(NewLeftRightAverage("Bus_Voltage"))
	Register(NewLeftRightSum("Bus_Current"))
	Register(NewMotorControllerEfficiency())
	Register(NewDrivetrainEfficiency())
	Register(NewModeledBusCurrent())
	Register(NewChargeIntegral("Modeled_Bus"))
	Register(NewChargeIntegral("Bus"))
}
