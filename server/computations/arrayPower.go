package computations

import (
	"server/datatypes"
)

// ArrayPower is the total power of zones comprising the solar array
type ArrayPower struct {
	standardComputation
}

func NewArrayPower() *ArrayPower {
	return &ArrayPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{	"MG_0_Input_Power",
								"Photon_Channel_0_Array_Voltage", "Photon_Channel_0_Array_Current", 
								"Photon_Channel_1_Array_Voltage", "Photon_Channel_1_Array_Current"},
		},
	}
}

func (a *ArrayPower) Compute() *datatypes.Datapoint {
	mg0Power := a.values["MG_0_Input_Power"] / 1000.0

	photon0Power := a.values["Photon_Channel_0_Array_Voltage"] * a.values["Photon_Channel_0_Array_Current"]
	photon1Power := a.values["Photon_Channel_1_Array_Voltage"] * a.values["Photon_Channel_1_Array_Current"]

	totalPower := mg0Power + photon0Power + photon1Power

	point := &datatypes.Datapoint{
		Metric: "Array_Power",
		Value: totalPower,
	}
	a.values = make(map[string]float64)

	return point
}

func init() {
	Register(NewArrayPower())
}