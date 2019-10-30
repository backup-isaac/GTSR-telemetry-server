package computations

import (
	"server/datatypes"
)

var imat = []float64{0, 36, 60, 120, 180}

var v = [][]float64{
	{2.500000, 3.220000, 3.450000, 3.520000, 3.580000, 3.650000, 3.750000, 3.830000, 3.950000, 4.050000, 4.150000, 4.190000},
	{2.500000, 3.050000, 3.300000, 3.380000, 3.450000, 3.550000, 3.650000, 3.750000, 3.850000, 3.950000, 4.080000, 4.120000},
	{2.500000, 3.000000, 3.220000, 3.300000, 3.400000, 3.480000, 3.580000, 3.680000, 3.780000, 3.870000, 4.000000, 4.050000},
	{2.500000, 2.900000, 3.100000, 3.180000, 3.250000, 3.350000, 3.450000, 3.550000, 3.650000, 3.750000, 3.870000, 3.900000},
	{2.500000, 2.830000, 3.000000, 3.100000, 3.180000, 3.250000, 3.330000, 3.430000, 3.520000, 3.630000, 3.830000, 3.850000}}

var q = [][]float64{
	{2998.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2886.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2884.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2855.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000},
	{2825.000000, 2700.000000, 2400.000000, 2100.000000, 1800.000000, 1500.000000, 1200.000000, 900.000000, 600.000000, 300.000000, 30.000000, 0.000000}}

// SOCPercentage is the estimated state of charge of the pack
type SOCPercentage struct {
	values map[string]float64
}

// NewSOCPercentage returns a SOCPercentage initialized with the proper values
func NewSOCPercentage() *SOCPercentage {
	return &SOCPercentage{
		values: make(map[string]float64),
	}
}

// GetMetrics returns the SOC Percentage's metrics
func (sp *SOCPercentage) GetMetrics() []string {
	return []string{"Min_Voltage", "BMS_Current"}
}

// Update waits for a BMS current and min voltage reading.
// Since min voltage updates so infrequently relative to BMS_Current,
// we only signify an update every time the Min_Voltage is updated
// So we override Update accordingly
func (sp *SOCPercentage) Update(point *datatypes.Datapoint) bool {
	sp.values[point.Metric] = point.Value
	return (len(sp.values) == 2) && point.Metric == "Min_Voltage"
}

// Compute returns the current percentage remaining using linear interpolation of current and voltage
func (sp *SOCPercentage) Compute() *datatypes.Datapoint {
	minVoltage := sp.values["Min_Voltage"]
	bmsCurrent := sp.values["BMS_Current"]
	point := &datatypes.Datapoint{
		Metric: "SOC_Percentage",
	}
	point.Value = lookupPercent(minVoltage, bmsCurrent)
	sp.values = make(map[string]float64)
	return point
}

func init() {
	Register(NewSOCPercentage())
}

func lookupPercent(voltage float64, current float64) float64 {
	I1, I2 := search(current, imat[:])
	ind1, ind2 := search(voltage, v[I1][:])
	Q1 := q[I1][ind1]
	if ind1 != ind2 {
		Q1 = inter(v[I1][ind1], v[I1][ind2], q[I1][ind1], q[I1][ind2], voltage)
	}

	ind3, ind4 := search(voltage, v[I2][:])
	Q2 := q[I2][ind3]
	if ind3 != ind4 {
		Q2 = inter(v[I2][ind3], v[I2][ind4], q[I2][ind3], q[I2][ind4], voltage)
	}

	chargeConsumed := Q1
	if I1 != I2 {
		chargeConsumed = inter(imat[I1], imat[I2], Q1, Q2, current)
	}
	// take the max charge possible consumed at 0Amp, to convert to percentage
	// subtract from 1 in order to have percent remaining
	percentRemaining := 1 - (chargeConsumed / q[0][0])
	return percentRemaining
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
