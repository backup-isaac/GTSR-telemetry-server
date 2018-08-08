package computations_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/GTSR/telemetry-server/computations"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

func TestBusPower(t *testing.T) {
	bp := computations.NewBusPower()
	done := bp.Update(&datatypes.Datapoint{
		Metric: "Bus_Voltage",
		Value:  50,
	})
	assert.False(t, done)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Bus_Current",
		Value:  100,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Bus_Power",
		Value:  5000,
	}
	actualPoint := bp.Compute()
	assert.Equal(t, expectedPoint, actualPoint)
	done = bp.Update(&datatypes.Datapoint{
		Metric: "Bus_Voltage",
		Value:  1,
	})
	assert.False(t, done)
}
