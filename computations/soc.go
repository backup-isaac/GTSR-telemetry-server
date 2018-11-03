package computations

import (
	"telemetry-server/datatypes"
)

var Imat = []float64{0, 36, 60, 120, 180}

var V = [][]float64{
	{2.500000, 3.220000, 3.450000, 3.520000, 3.580000, 3.650000, 3.750000, 3.830000, 3.950000, 4.050000, 4.150000, 4.190000},
	{2.500000, 3.050000, 3.300000, 3.380000, 3.450000, 3.550000, 3.650000, 3.750000, 3.850000, 3.950000, 4.080000, 4.120000},
	{2.500000, 3.000000, 3.220000, 3.300000, 3.400000, 3.480000, 3.580000, 3.680000, 3.780000, 3.870000, 4.000000, 4.050000},
	{2.500000, 2.900000, 3.100000, 3.180000, 3.250000, 3.350000, 3.450000, 3.550000, 3.650000, 3.750000, 3.870000, 3.900000},
	{2.500000, 2.830000, 3.000000, 3.100000, 3.180000, 3.250000, 3.330000, 3.430000, 3.520000, 3.630000, 3.830000, 3.850000}}

var Q = [][]float64{
	{2998.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2886.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2884.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2855.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2825.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000}}

// SOCPercentage is the estimated state of charge of the pack
type SOCPercentage struct {
	values map[string]float64
}

// NewBatteryPower returns a SOCPercentage initialized with the proper values
func NewSOCPercentage() *SOCPercentage {
	return &SOCPercentage {
		values: make(map[string]float64),
	}
}

func (sp *SOCPercentage) GetMetrics() []string {
	return []string{"Min_Voltage", "BMS_Current"}
}

// Since min voltage updates so infrequently relative to BMS_Current,
// we only signify an update every time the Min_Voltage is updated
// So we override Update accordingly
func (sp *SOCPercentage) Update(point *datatypes.Datapoint) bool {
	sp.values[point.Metric] = point.Value
	return (len(sp.values) == 2) && point.Metric == "Min_Voltage"
}

// Returns the current percentage remaing using linear interpolation of current and voltage
func (sp *SOCPercentage) Compute() *datatypes.Datapoint {
	minVoltage := sp.values["Min_Voltage"]
	bmsCurrent := sp.values["BMS_Current"]
	point := &datatypes.Datapoint{
		Metric: "SOC_Percentage",
	}
	point.Value = lookup_percent(minVoltage, bmsCurrent)
	sp.values = make(map[string]float64)
	return point
}

func init() {
	Register(NewSOCPercentage())
}

func lookup_percent(voltage float64, current float64) float64 {
	I1, I2 := search(current, Imat[:])
	ind1, ind2 := search(voltage, V[I1][:])
	Q_1 := Q[I1][ind1]
	if ind1 != ind2 {
		Q_1 = inter(V[I1][ind1], V[I1][ind2], Q[I1][ind1], Q[I1][ind2], voltage)
	}

	ind3, ind4 := search(voltage, V[I2][:])
	Q_2 := Q[I2][ind3]
	if ind3 != ind4 {
		Q_2 = inter(V[I2][ind3], V[I2][ind4], Q[I2][ind3], Q[I2][ind4], voltage)
	}

	charge_consumed := Q_1
	if I1 != I2 {
		charge_consumed = inter(Imat[I1], Imat[I2], Q_1, Q_2, current)
	}
	// take the max charge possible consumed at 0Amp, to convert to percentage
	// subtract from 1 in order to have percent remaining
	percent_remaining := 1 - (charge_consumed / Q[0][0])
	return percent_remaining
}

func inter(x1 float64, x2 float64, y1 float64, y2 float64, x float64) float64 {
	m := (y2 - y1) / (x2 - x1)
	b := y1

	return m*(x-x1) + b
}

func search(val float64, arr []float64) (int, int) {
	start := 0
	end := len(arr)

	for start < end {
		midpoint := (start + end) / 2
		if arr[midpoint] == val {
			return midpoint, midpoint
		} else if arr[midpoint] > val {
			end = midpoint
		} else {
			start = midpoint + 1
		}
	}
	// check bounds before returning
	if start >= len(arr) {
		return len(arr) - 1, len(arr) - 1
	} else if start <= 0 {
		return 0, 0
	} else {
		return start - 1, start
	}
}
