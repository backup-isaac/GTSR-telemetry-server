package computations_test

import (
	"testing"

	"server/computations"
	"server/datatypes"

	"github.com/stretchr/testify/assert"
)

func TestArrayPower(t *testing.T) {
	array := computations.NewArrayPower()
	done := array.Update(&datatypes.Datapoint{
		Metric: "MG_0_Input_Power",
		Value:  10000,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_0_Array_Voltage",
		Value:  0,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_0_Array_Current",
		Value:  5,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_1_Array_Voltage",
		Value:  70,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_1_Array_Current",
		Value:  6,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Array_Power",
		Value:  430,
	}
	actualPoint := array.Compute()
	assert.Equal(t, expectedPoint, actualPoint)

	done := array.Update(&datatypes.Datapoint{
		Metric: "MG_0_Input_Power",
		Value:  10000,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_0_Array_Voltage",
		Value:  10,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_0_Array_Current",
		Value:  5,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_1_Array_Voltage",
		Value:  70,
	})
	assert.False(t, done)
	done = array.Update(&datatypes.Datapoint{
		Metric: "Photon_Channel_1_Array_Current",
		Value:  6,
	})
	assert.True(t, done)
	expectedPoint := &datatypes.Datapoint{
		Metric: "Array_Power",
		Value:  480,
	}
	actualPoint = array.Compute()
	assert.Equal(t, expectedPoint, actualPoint)
}
