package computations_test

import (
	"testing"

	"telemetry-server/computations"
	"telemetry-server/datatypes"

	"github.com/stretchr/testify/assert"
)

func TestSOCPercentage(t *testing.T) {
	sp := computations.NewSOCPercentage()
	done := sp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  38.4,
	})
	assert.False(t, done)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  0.0,
	})
	assert.False(t, done)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "Min_Voltage",
		Value:  2.5,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "SOC_Percentage",
		Value:  0.0,
	}
	actualPoint := sp.Compute()
	assert.InDelta(t, expectedPoint.Value, actualPoint.Value, 0.001)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  0.0,
	})
	assert.False(t, done)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "Min_Voltage",
		Value:  4.19,
	})
	assert.True(t, done)
	expectedPoint = &datatypes.Datapoint{
		Metric: "SOC_Percentage",
		Value:  1.0,
	}
	actualPoint = sp.Compute()
	assert.InDelta(t, expectedPoint.Value, actualPoint.Value, 0.001)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "BMS_Current",
		Value:  38.4,
	})
	assert.False(t, done)
	done = sp.Update(&datatypes.Datapoint{
		Metric: "Min_Voltage",
		Value:  3.09,
	})
	assert.True(t, done)
	expectedPoint = &datatypes.Datapoint{
		Metric: "SOC_Percentage",
		Value:  0.117903,
	}
	actualPoint = sp.Compute()
	assert.InDelta(t, expectedPoint.Value, actualPoint.Value, 0.001)

}
