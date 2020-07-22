package computations

import (
	"server/datatypes"
	"server/recontool"
)

// MotorControllerEfficiency computes motor controller efficiency
// based on phase current, bus voltage, and bus power
type MotorControllerEfficiency struct {
	standardComputation
}

// NewMotorControllerEfficiency returns an initialized MotorControllerEfficiency
func NewMotorControllerEfficiency() *MotorControllerEfficiency {
	return &MotorControllerEfficiency{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Phase_C_Current", "Average_Bus_Voltage", "Bus_Power"},
		},
	}
}

// Update signifies an update when all required metrics have been received
// A point with zero bus power will cause everything to be thrown out
func (e *MotorControllerEfficiency) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Bus_Power" && point.Value == 0 {
		e.values = make(map[string]float64)
		return false
	}
	e.values[point.Metric] = point.Value
	if point.Time.After(e.timestamp) {
		e.timestamp = point.Time
	}
	return len(e.values) >= len(e.fields)
}

// Compute returns the motor controller efficiency
func (e *MotorControllerEfficiency) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Motor_Controller_Efficiency",
		Value: recontool.MotorControllerEfficiency(
			e.values["Phase_C_Current"],
			e.values["Average_Bus_Voltage"],
			e.values["Bus_Power"],
		),
		Time: e.timestamp,
	}
	e.values = make(map[string]float64)
	return datapoint
}

// DrivetrainEfficiency gives the total drivetrain efficiency
type DrivetrainEfficiency struct {
	standardComputation
}

// NewDrivetrainEfficiency returns an initialized DrivetrainEfficiency
func NewDrivetrainEfficiency() *DrivetrainEfficiency {
	return &DrivetrainEfficiency{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Motor_Efficiency", "Motor_Controller_Efficiency", "Pack_Efficiency"},
		},
	}
}

// Compute computes drivetrain efficiency
func (e *DrivetrainEfficiency) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Drivetrain_Efficiency",
		Value: recontool.DrivetrainEfficiency(
			e.values["Motor_Controller_Efficiency"],
			e.values["Pack_Efficiency"],
			e.values["Motor_Efficiency"],
		),
		Time: e.timestamp,
	}
	e.values = make(map[string]float64)
	return datapoint
}

// ModeledBusCurrent derives bus current from modeled power
type ModeledBusCurrent struct {
	standardComputation
	// power   *datatypes.Datapoint
	// voltage *datatypes.Datapoint
}

// NewModeledBusCurrent returns an initialized ModeledBusCurrent
func NewModeledBusCurrent() *ModeledBusCurrent {
	return &ModeledBusCurrent{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Modeled_Motor_Power", "Average_Bus_Voltage"},
		},
	}
}

// Update signifies an update when all required metrics have been received
func (c *ModeledBusCurrent) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Average_Bus_Voltage" && point.Value == 0 {
		c.values = make(map[string]float64)
		return false
	}
	c.values[point.Metric] = point.Value
	if point.Time.After(c.timestamp) {
		c.timestamp = point.Time
	}
	return len(c.values) >= len(c.fields)
}

// Compute computes modeled bus current in amps
func (c *ModeledBusCurrent) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Modeled_Bus_Current",
		Value:  c.values["Modeled_Motor_Power"] / c.values["Average_Bus_Voltage"],
		Time:   c.timestamp,
	}
	c.values = make(map[string]float64)
	return datapoint
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
