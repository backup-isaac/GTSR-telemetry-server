package computations

import (
	"fmt"
	"server/datatypes"
	"server/recontool"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackResistance(t *testing.T) {
	r := NewPackResistance()
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 120),
		makeDatapoint("BMS_Current", 0),
		makeDatapoint("Average_Bus_Voltage", 118),
		makeDatapoint("BMS_Current", 1),
	}, &datatypes.Datapoint{
		Metric: "Pack_Resistance",
		Value:  2.0,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 117),
		makeDatapoint("Average_Bus_Voltage", 120),
		makeDatapoint("BMS_Current", 2),
	}, &datatypes.Datapoint{
		Metric: "Pack_Resistance",
		Value:  1.5,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("Connection_Status", 0),
		makeDatapoint("BMS_Current", 0),
		makeDatapoint("Average_Bus_Voltage", 140),
		makeDatapoint("BMS_Current", 1),
		makeDatapoint("Connection_Status", 1),
		makeDatapoint("BMS_Current", 5),
		makeDatapoint("Average_Bus_Voltage", 139.5),
	}, &datatypes.Datapoint{
		Metric: "Pack_Resistance",
		Value:  0.5,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("Average_Bus_Voltage", 0),
		makeDatapoint("BMS_Current", 10),
		makeDatapoint("Average_Bus_Voltage", 138),
		makeDatapoint("BMS_Current", 4),
	}, &datatypes.Datapoint{
		Metric: "Pack_Resistance",
		Value:  0.5,
		Time:   pointTime,
	})
}

func TestPackEfficiency(t *testing.T) {
	e := NewPackEfficiency()
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Bus_Current", 4),
		makeDatapoint("Bus_Current", 6),
		makeDatapoint("Bus_Power", 700),
		makeDatapoint("Pack_Resistance", 0.1),
	}, &datatypes.Datapoint{
		Metric: "Pack_Efficiency",
		Value:  recontool.PackEfficiency(6, 700, 0.1),
		Time:   pointTime,
	})
	computationRunner(t, e, []*datatypes.Datapoint{
		makeDatapoint("Pack_Resistance", 0.2),
		makeDatapoint("Bus_Current", 15),
		makeDatapoint("Bus_Power", 1600),
	}, &datatypes.Datapoint{
		Metric: "Pack_Efficiency",
		Value:  recontool.PackEfficiency(15, 1600, 0.2),
		Time:   pointTime,
	})
}

func TestMinModuleVoltage(t *testing.T) {
	m := NewMinModuleVoltage()
	assert.Equal(t, []string{
		"Cell_Voltage_1", "Cell_Voltage_2", "Cell_Voltage_3",
		"Cell_Voltage_4", "Cell_Voltage_5", "Cell_Voltage_6",
		"Cell_Voltage_7", "Cell_Voltage_8", "Cell_Voltage_9",
		"Cell_Voltage_10", "Cell_Voltage_11", "Cell_Voltage_12",
		"Cell_Voltage_13", "Cell_Voltage_14", "Cell_Voltage_15",
		"Cell_Voltage_16", "Cell_Voltage_17", "Cell_Voltage_18",
		"Cell_Voltage_19", "Cell_Voltage_20", "Cell_Voltage_21",
		"Cell_Voltage_22", "Cell_Voltage_23", "Cell_Voltage_24",
		"Cell_Voltage_25", "Cell_Voltage_26", "Cell_Voltage_27",
		"Cell_Voltage_28", "Cell_Voltage_29", "Cell_Voltage_30",
		"Cell_Voltage_31", "Cell_Voltage_32", "Cell_Voltage_33",
		"Cell_Voltage_34", "Cell_Voltage_35",
	}, m.GetMetrics())
	computationRunner(t, m, moduleVoltagePoints(35, 3, 18, []int{5}), &datatypes.Datapoint{
		Metric: "Min_Cell_Voltage",
		Value:  3.0,
		Time:   pointTime,
	})
	computationRunner(t, m, moduleVoltagePoints(35, 15, 9, []int{}), &datatypes.Datapoint{
		Metric: "Min_Cell_Voltage",
		Value:  15.0,
		Time:   pointTime,
	})
}

func TestMaxModuleVoltage(t *testing.T) {
	m := NewMaxModuleVoltage()
	assert.Equal(t, []string{
		"Cell_Voltage_1", "Cell_Voltage_2", "Cell_Voltage_3",
		"Cell_Voltage_4", "Cell_Voltage_5", "Cell_Voltage_6",
		"Cell_Voltage_7", "Cell_Voltage_8", "Cell_Voltage_9",
		"Cell_Voltage_10", "Cell_Voltage_11", "Cell_Voltage_12",
		"Cell_Voltage_13", "Cell_Voltage_14", "Cell_Voltage_15",
		"Cell_Voltage_16", "Cell_Voltage_17", "Cell_Voltage_18",
		"Cell_Voltage_19", "Cell_Voltage_20", "Cell_Voltage_21",
		"Cell_Voltage_22", "Cell_Voltage_23", "Cell_Voltage_24",
		"Cell_Voltage_25", "Cell_Voltage_26", "Cell_Voltage_27",
		"Cell_Voltage_28", "Cell_Voltage_29", "Cell_Voltage_30",
		"Cell_Voltage_31", "Cell_Voltage_32", "Cell_Voltage_33",
		"Cell_Voltage_34", "Cell_Voltage_35",
	}, m.GetMetrics())
	computationRunner(t, m, moduleVoltagePoints(35, 3, 18, []int{5}), &datatypes.Datapoint{
		Metric: "Max_Cell_Voltage",
		Value:  18.0,
		Time:   pointTime,
	})
	computationRunner(t, m, moduleVoltagePoints(35, 15, 9, []int{}), &datatypes.Datapoint{
		Metric: "Max_Cell_Voltage",
		Value:  9.0,
		Time:   pointTime,
	})
}

func TestModuleVoltageImbalance(t *testing.T) {
	m := NewModuleVoltageImbalance()
	assert.Equal(t, []string{
		"Cell_Voltage_1", "Cell_Voltage_2", "Cell_Voltage_3",
		"Cell_Voltage_4", "Cell_Voltage_5", "Cell_Voltage_6",
		"Cell_Voltage_7", "Cell_Voltage_8", "Cell_Voltage_9",
		"Cell_Voltage_10", "Cell_Voltage_11", "Cell_Voltage_12",
		"Cell_Voltage_13", "Cell_Voltage_14", "Cell_Voltage_15",
		"Cell_Voltage_16", "Cell_Voltage_17", "Cell_Voltage_18",
		"Cell_Voltage_19", "Cell_Voltage_20", "Cell_Voltage_21",
		"Cell_Voltage_22", "Cell_Voltage_23", "Cell_Voltage_24",
		"Cell_Voltage_25", "Cell_Voltage_26", "Cell_Voltage_27",
		"Cell_Voltage_28", "Cell_Voltage_29", "Cell_Voltage_30",
		"Cell_Voltage_31", "Cell_Voltage_32", "Cell_Voltage_33",
		"Cell_Voltage_34", "Cell_Voltage_35",
	}, m.GetMetrics())
	computationRunner(t, m, moduleVoltagePoints(35, 3, 18, []int{5}), &datatypes.Datapoint{
		Metric: "Cell_Voltage_Imbalance",
		Value:  0.215,
		Time:   pointTime,
	})
	computationRunner(t, m, moduleVoltagePoints(35, 15, 9, []int{}), &datatypes.Datapoint{
		Metric: "Cell_Voltage_Imbalance",
		Value:  0.194,
		Time:   pointTime,
	})
}

func moduleVoltagePoints(number, argmin, argmax int, duplicates []int) []*datatypes.Datapoint {
	points := make([]*datatypes.Datapoint, number+len(duplicates))
	for i := 0; i < len(duplicates); i++ {
		adj := float64(duplicates[i]) / 1000.0
		if duplicates[i] == argmax {
			adj += 0.1
		} else if duplicates[i] == argmin {
			adj -= 0.1
		}
		points[i] = makeDatapoint(fmt.Sprintf("Cell_Voltage_%d", duplicates[i]), 3.50+adj)
	}
	for i := 1; i <= number; i++ {
		adj := float64(i) / 1000.0
		if i == argmax {
			adj += 0.1
		} else if i == argmin {
			adj -= 0.1
		}
		points[i-1+len(duplicates)] = makeDatapoint(fmt.Sprintf("Cell_Voltage_%d", i), 3.50+adj)
	}
	return points
}

func TestModuleResistance(t *testing.T) {
	r := NewModuleResistance(69)
	assert.Equal(t, []string{"Cell_Voltage_69", "BMS_Current", "Connection_Status"}, r.GetMetrics())
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("BMS_Current", 1),
		makeDatapoint("BMS_Current", 2),
		makeDatapoint("Cell_Voltage_69", 3.4),
		makeDatapoint("Cell_Voltage_69", 3.35),
	}, &datatypes.Datapoint{
		Metric: "Cell_Resistance_69",
		Value:  0.05,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("BMS_Current", -2),
	}, &datatypes.Datapoint{
		Metric: "Cell_Resistance_69",
		Value:  0.0125,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("BMS_Current", -3),
	}, &datatypes.Datapoint{
		Metric: "Cell_Resistance_69",
		Value:  0.01,
		Time:   pointTime,
	})
	computationRunner(t, r, []*datatypes.Datapoint{
		makeDatapoint("Connection_Status", 0),
		makeDatapoint("Cell_Voltage_69", 3.45),
		makeDatapoint("Connection_Status", 1),
		makeDatapoint("BMS_Current", 0),
		makeDatapoint("Cell_Voltage_69", 3.43),
		makeDatapoint("BMS_Current", 2),
	}, &datatypes.Datapoint{
		Metric: "Cell_Resistance_69",
		Value:  0.01,
		Time:   pointTime,
	})
}
