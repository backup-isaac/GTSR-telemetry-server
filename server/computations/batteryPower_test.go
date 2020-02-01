package computations_test

import (
	"testing"

	"server/computations"
	"server/datatypes"

	"github.com/stretchr/testify/assert"
)

func TestBatteryPower(t *testing.T) {
	bp := computations.NewBatteryPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Voltage",
		Value:  0,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Voltage",
		Value:  0,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  50,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Pack_Voltage",
		Value:  100,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Battery_Power",
		Value:  5000,
	}
	actualPoint := bp.Compute()
	expectedPoint.Time = actualPoint.Time
	assert.Equal(t, expectedPoint, actualPoint)

	done = bp.Update(&datatypes.Datapoint{
		Metric: "Left_Bus_Voltage",
		Value:  100,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Right_Bus_Voltage",
		Value:  200,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  50,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Pack_Voltage",
		Value:  100,
	})
	assert.True(t, done)
	expectedPoint = &datatypes.Datapoint{
		Metric: "Battery_Power",
		Value:  7500,
	}
	actualPoint = bp.Compute()
	expectedPoint.Time = actualPoint.Time
	assert.Equal(t, expectedPoint, actualPoint)
}
