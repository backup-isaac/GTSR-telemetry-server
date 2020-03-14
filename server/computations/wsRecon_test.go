package computations

import (
	"server/datatypes"
	"server/recontool"
	"testing"
)

func TestMotorControllerEfficiency(t *testing.T) {
	mc := NewMotorControllerEfficiency()
	computationRunner(t, mc, []*datatypes.Datapoint{
		makeDatapoint("Phase_C_Current", 10),
		makeDatapoint("Average_Bus_Voltage", 125),
		makeDatapoint("Average_Bus_Voltage", 120),
		makeDatapoint("Bus_Power", 600),
	}, &datatypes.Datapoint{
		Metric: "Motor_Controller_Efficiency",
		Value:  recontool.MotorControllerEfficiency(10, 120, 600),
		Time:   pointTime,
	})
	computationRunner(t, mc, []*datatypes.Datapoint{
		makeDatapoint("Phase_C_Current", 30),
		makeDatapoint("Average_Bus_Voltage", 130),
		makeDatapoint("Bus_Power", 0),
		makeDatapoint("Phase_C_Current", 10),
		makeDatapoint("Average_Bus_Voltage", 135),
		makeDatapoint("Bus_Power", 450),
	}, &datatypes.Datapoint{
		Metric: "Motor_Controller_Efficiency",
		Value:  recontool.MotorControllerEfficiency(10, 135, 450),
		Time:   pointTime,
	})
}

func TestDrivetrainEfficiency(t *testing.T) {
	dt := NewDrivetrainEfficiency()
	computationRunner(t, dt, []*datatypes.Datapoint{
		makeDatapoint("Motor_Efficiency", 0.99),
		makeDatapoint("Motor_Controller_Efficiency", 0.98),
		makeDatapoint("Motor_Efficiency", 0.98),
		makeDatapoint("Pack_Efficiency", 0.97),
	}, &datatypes.Datapoint{
		Metric: "Drivetrain_Efficiency",
		Value:  recontool.DrivetrainEfficiency(0.98, 0.97, 0.98),
		Time:   pointTime,
	})
	computationRunner(t, dt, []*datatypes.Datapoint{
		makeDatapoint("Motor_Efficiency", 0.97),
		makeDatapoint("Motor_Controller_Efficiency", 0.96),
		makeDatapoint("Pack_Efficiency", 0.97),
	}, &datatypes.Datapoint{
		Metric: "Drivetrain_Efficiency",
		Value:  recontool.DrivetrainEfficiency(0.96, 0.97, 0.97),
		Time:   pointTime,
	})
}

func TestModeledBusCurrent(t *testing.T) {
	bc := NewModeledBusCurrent()
	computationRunner(t, bc, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 115),
		makeDatapoint("Average_Bus_Voltage", 122),
		makeDatapoint("Modeled_Motor_Power", 900),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Bus_Current",
		Value:  900.0 / 122,
		Time:   pointTime,
	})
	computationRunner(t, bc, []*datatypes.Datapoint{
		makeDatapoint("Modeled_Motor_Power", 1000),
		makeDatapoint("Average_Bus_Voltage", 0),
		makeDatapoint("Average_Bus_Voltage", 125),
		makeDatapoint("Modeled_Motor_Power", 1200),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Bus_Current",
		Value:  1200.0 / 125,
		Time:   pointTime,
	})
}
