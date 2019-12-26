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
