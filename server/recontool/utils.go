package recontool

func meanIf(data []float64, predicate func(float64) bool) float64 {
	sum := 0.0
	count := 0
	for _, d := range data {
		if predicate(d) {
			sum += d
			count++
		}
	}
	return sum / float64(count)
}

// CalculateSeries uses the provided calc() function to calculate a series
// of data points from the input data series
// Each output[i] = calc(input[0][i], input[1][i], input[2][i], ...)
func CalculateSeries(calc func(params ...float64) float64, inputs ...[]float64) []float64 {
	result := make([]float64, len(inputs[0]))
	for i := range inputs[0] {
		params := make([]float64, len(inputs))
		for j, v := range inputs {
			params[j] = v[i]
		}
		result[i] = calc(params...)
	}
	return result
}
