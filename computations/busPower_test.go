package computations_test

import (
	"testing"

	"telemetry-server/computations"
	"telemetry-server/datatypes"

	"github.com/stretchr/testify/assert"
)

func TestBusPower(t *testing.T) {
	bp := computations.NewBusPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Power",
		Value:  10,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Power",
		Value:  20,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Bus_Power",
		Value:  30,
	}
	point := bp.Compute()
	assert.Equal(t, expectedPoint, point)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Power",
		Value:  10,
	})
	assert.False(t, done)
}

func TestLeftBusPower(t *testing.T) {
	bp := computations.NewLeftBusPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Voltage",
		Value:  50,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Current",
		Value:  100,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Left_Bus_Power",
		Value:  5000,
	}
	actualPoint := bp.Compute()
	assert.Equal(t, expectedPoint, actualPoint)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Voltage",
		Value:  1,
	})
	assert.False(t, done)
}

func TestRightBusPower(t *testing.T) {
	bp := computations.NewRightBusPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Voltage",
		Value:  50,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Current",
		Value:  100,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Right_Bus_Power",
		Value:  5000,
	}
	actualPoint := bp.Compute()
	assert.Equal(t, expectedPoint, actualPoint)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Voltage",
		Value:  1,
	})
	assert.False(t, done)
}
