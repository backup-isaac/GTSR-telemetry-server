package computations

import (
	"server/datatypes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLeftRightSum(t *testing.T) {
	s := NewLeftRightSum("Test")
	assert.Equal(t, []string{"Left_Test", "Right_Test"}, s.GetMetrics())
	computationRunner(t, s, []*datatypes.Datapoint{
		makeDatapoint("Left_Test", 1),
		makeDatapoint("Left_Test", 2),
		makeDatapoint("Right_Test", -1),
	}, &datatypes.Datapoint{
		Metric: "Test",
		Value:  1.0,
		Time:   pointTime,
	})
	s1 := NewLeftRightSum("Test_Foo")
	assert.Equal(t, []string{"Left_Test_Foo", "Right_Test_Foo"}, s1.GetMetrics())
	computationRunner(t, s1, []*datatypes.Datapoint{
		makeDatapoint("Left_Test_Foo", 100),
		makeDatapoint("Right_Test_Foo", -31),
	}, &datatypes.Datapoint{
		Metric: "Test_Foo",
		Value:  69.0,
		Time:   pointTime,
	})
	computationRunner(t, s, []*datatypes.Datapoint{
		makeDatapoint("Right_Test", 4),
		makeDatapoint("Left_Test", 3),
	}, &datatypes.Datapoint{
		Metric: "Test",
		Value:  7.0,
		Time:   pointTime,
	})
}

func TestLeftRightAverage(t *testing.T) {
	s := NewLeftRightAverage("Test")
	assert.Equal(t, []string{"Left_Test", "Right_Test"}, s.GetMetrics())
	computationRunner(t, s, []*datatypes.Datapoint{
		makeDatapoint("Left_Test", 1),
		makeDatapoint("Left_Test", 2),
		makeDatapoint("Right_Test", -1),
	}, &datatypes.Datapoint{
		Metric: "Average_Test",
		Value:  0.5,
		Time:   pointTime,
	})
	s1 := NewLeftRightAverage("Test_Foo")
	assert.Equal(t, []string{"Left_Test_Foo", "Right_Test_Foo"}, s1.GetMetrics())
	computationRunner(t, s1, []*datatypes.Datapoint{
		makeDatapoint("Left_Test_Foo", 100),
		makeDatapoint("Right_Test_Foo", -31),
	}, &datatypes.Datapoint{
		Metric: "Average_Test_Foo",
		Value:  34.5,
		Time:   pointTime,
	})
	computationRunner(t, s, []*datatypes.Datapoint{
		makeDatapoint("Right_Test", 4),
		makeDatapoint("Left_Test", 3),
	}, &datatypes.Datapoint{
		Metric: "Average_Test",
		Value:  3.5,
		Time:   pointTime,
	})
}

func TestChargeIntegral(t *testing.T) {
	i := NewChargeIntegral("Test")
	assert.Equal(t, []string{"Test_Current", "Connection_Status"}, i.GetMetrics())
	computationRunner(t, i, []*datatypes.Datapoint{
		makeDatapoint("Test_Current", 5),
		makeDatapoint("Test_Current", 2),
	}, &datatypes.Datapoint{
		Metric: "Test_Charge_Consumed",
		Value:  0.005,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, i, []*datatypes.Datapoint{
		makeDatapoint("Test_Current", 3),
	}, &datatypes.Datapoint{
		Metric: "Test_Charge_Consumed",
		Value:  0.007,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, i, []*datatypes.Datapoint{
		makeDatapoint("Test_Current", 3),
	}, &datatypes.Datapoint{
		Metric: "Test_Charge_Consumed",
		Value:  0.01,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
	computationRunner(t, i, []*datatypes.Datapoint{
		makeDatapoint("Connection_Status", 0),
		makeDatapoint("Test_Current", -1),
		makeDatapoint("Test_Current", -2),
	}, &datatypes.Datapoint{
		Metric: "Test_Charge_Consumed",
		Value:  -0.001,
		Time:   pointTime.Add(time.Millisecond * -1),
	})
}
