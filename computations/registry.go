package computations

var registry = make(map[Computable][]string)

// Register registers a computable with the given metrics.
// Call this function in init() in your computable file
func Register(computation Computable, metrics []string) {
	registry[computation] = metrics
}
