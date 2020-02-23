package computations

import (
	"server/datatypes"
	"server/recontool"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVelocity(t *testing.T) {
	v := NewVelocity()
	computationRunner(t, v, []*datatypes.Datapoint{
		makeDatapoint("Average_Wavesculptor_RPM", 30),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(30, sr3.RMot),
		Time:   pointTime,
	})
	computationRunner(t, v, []*datatypes.Datapoint{
		makeDatapoint("Average_Wavesculptor_RPM", -10),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(-10, sr3.RMot),
		Time:   pointTime,
	})
}

func TestAcceleration(t *testing.T) {
	a := NewAcceleration()
	computationRunner(t, a, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 0),
		makeDatapoint("RPM_Derived_Velocity", 1),
		makeDatapoint("RPM_Derived_Velocity", 2),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Acceleration",
		Value:  2 / 0.002,
		Time:   pointTime.Add(-1 * time.Millisecond),
	})
	makeDatapoint("", 0)
	computationRunner(t, a, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 3),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Acceleration",
		Value:  2 / 0.003,
		Time:   pointTime.Add(-2 * time.Millisecond),
	})
	computationRunner(t, a, []*datatypes.Datapoint{
		makeDatapoint("Connection_Status", 0),
		makeDatapoint("RPM_Derived_Velocity", 1),
		makeDatapoint("RPM_Derived_Velocity", 0),
		makeDatapoint("Connection_Status", 1),
		makeDatapoint("RPM_Derived_Velocity", -1),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Acceleration",
		Value:  -2 / 0.003,
		Time:   pointTime.Add(-2 * time.Millisecond),
	})
}

func TestDistance(t *testing.T) {
	d := NewDistance()
	computationRunner(t, d, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 5),
		makeDatapoint("RPM_Derived_Velocity", 2),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  0.005,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, d, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 3),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  0.007,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, d, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 3),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  0.01,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, d, []*datatypes.Datapoint{
		makeDatapoint("Connection_Status", 0),
		makeDatapoint("RPM_Derived_Velocity", -1),
		makeDatapoint("Connection_Status", 1),
		makeDatapoint("RPM_Derived_Velocity", -2),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  -0.002,
		Time:   pointTime.Add(time.Millisecond * -2),
	})
}

func TestEmpiricalTorque(t *testing.T) {
	e := NewEmpiricalTorque("Test")
	assert.Equal(t, []string{"Test_Phase_C_Current", "Test_Wavesculptor_RPM"}, e.GetMetrics())
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Test_Wavesculptor_RPM", 60),
		makeDatapoint("Test_Wavesculptor_RPM", 40),
		makeDatapoint("Test_Phase_C_Current", 15),
	}, &datatypes.Datapoint{
		Metric: "Test_RPM_Derived_Torque",
		Value:  recontool.MotorTorque(40, 15, sr3.TMax),
		Time:   pointTime,
	})
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Test_Phase_C_Current", 23),
		makeDatapoint("Test_Wavesculptor_RPM", 80),
	}, &datatypes.Datapoint{
		Metric: "Test_RPM_Derived_Torque",
		Value:  recontool.MotorTorque(80, 23, sr3.TMax),
		Time:   pointTime,
	})
}

func TestModeledMotorForce(t *testing.T) {
	f := NewModeledMotorForce()
	computationRunner(t, f, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Acceleration", 4),
		makeDatapoint("RPM_Derived_Acceleration", 2),
		makeDatapoint("RPM_Derived_Velocity", 0),
		makeDatapoint("Terrain_Angle", 0.1),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Force",
		Value:  recontool.ModeledMotorForce(0, 2, 0.1, sr3),
		Time:   pointTime,
	})
	computationRunner(t, f, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Acceleration", 10),
		makeDatapoint("RPM_Derived_Velocity", 20),
		makeDatapoint("Terrain_Angle", -0.02),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Force",
		Value:  recontool.ModeledMotorForce(20, 10, -0.02, sr3),
		Time:   pointTime,
	})
}

func TestModeledMotorTorque(t *testing.T) {
	m := NewModeledMotorTorque()
	computationRunner(t, m, []*datatypes.Datapoint{
		makeDatapoint("Modeled_Motor_Force", 600),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Torque",
		Value:  600 * sr3.RMot,
		Time:   pointTime,
	})
	computationRunner(t, m, []*datatypes.Datapoint{
		makeDatapoint("Modeled_Motor_Force", 500),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Torque",
		Value:  500 * sr3.RMot,
		Time:   pointTime,
	})
}

func TestMotorEfficiency(t *testing.T) {
	e := NewMotorEfficiency()
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 120),
		makeDatapoint("Average_Bus_Voltage", 125),
		makeDatapoint("RPM_Derived_Torque", 10),
	}, &datatypes.Datapoint{
		Metric: "Motor_Efficiency",
		Value:  recontool.MotorEfficiency(125, 10),
		Time:   pointTime,
	})
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 120),
		makeDatapoint("RPM_Derived_Torque", -500),
	}, &datatypes.Datapoint{
		Metric: "Motor_Efficiency",
		Value:  1.0,
		Time:   pointTime,
	})
}

func TestEmpiricalMotorPower(t *testing.T) {
	p := NewEmpiricalMotorPower()
	computationRunner(t, p, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Torque", 10),
		makeDatapoint("RPM_Derived_Velocity", 45),
		makeDatapoint("RPM_Derived_Torque", 20),
		makeDatapoint("Phase_C_Current", 15),
		makeDatapoint("Drivetrain_Efficiency", 0.93),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Motor_Power",
		Value:  recontool.MotorPower(20, 45, 15, 0.93, sr3),
		Time:   pointTime,
	})
	computationRunner(t, p, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 45),
		makeDatapoint("RPM_Derived_Torque", 20),
		makeDatapoint("Phase_C_Current", 15),
		makeDatapoint("Drivetrain_Efficiency", 0),
		makeDatapoint("Phase_C_Current", 20),
		makeDatapoint("RPM_Derived_Torque", 25),
		makeDatapoint("Drivetrain_Efficiency", 0.95),
		makeDatapoint("RPM_Derived_Velocity", 70),
	}, &datatypes.Datapoint{
		Metric: "RPM_Derived_Motor_Power",
		Value:  recontool.MotorPower(25, 70, 20, 0.95, sr3),
		Time:   pointTime,
	})
}

func TestModeledMotorPower(t *testing.T) {
	p := NewModeledMotorPower()
	computationRunner(t, p, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 45),
		makeDatapoint("Modeled_Motor_Force", 500),
		makeDatapoint("RPM_Derived_Velocity", 50),
		makeDatapoint("Drivetrain_Efficiency", 0.93),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Power",
		Value:  recontool.ModelDerivedPower(500, 50, 0.93),
		Time:   pointTime,
	})
	computationRunner(t, p, []*datatypes.Datapoint{
		makeDatapoint("RPM_Derived_Velocity", 40),
		makeDatapoint("Modeled_Motor_Force", 700),
		makeDatapoint("Drivetrain_Efficiency", 0),
		makeDatapoint("Modeled_Motor_Force", 800),
		makeDatapoint("Drivetrain_Efficiency", 0.95),
		makeDatapoint("RPM_Derived_Velocity", 60),
	}, &datatypes.Datapoint{
		Metric: "Modeled_Motor_Power",
		Value:  recontool.ModelDerivedPower(800, 60, 0.95),
		Time:   pointTime,
	})
}
