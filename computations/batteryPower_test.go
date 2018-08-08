package computations_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/GTSR/telemetry-server/computations"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

func TestBatteryPower(t *testing.T) {
	bp := computations.NewBatteryPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Bus_Voltage",
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
	assert.Equal(t, expectedPoint, actualPoint)

	done = bp.Update(&datatypes.Datapoint{
		Metric: "Bus_Voltage",
		Value:  150,
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
	assert.Equal(t, expectedPoint, actualPoint)
}
