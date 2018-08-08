package computations

var registry = make(map[string][]Computable)

// Register registers a computable with the given metrics.
// Call this function in init() in your computable file
func Register(computation Computable, metrics []string) {
	for _, metric := range metrics {
		registry[metric] = append(registry[metric], computation)
	}
}

// LoadComputables loads the computable metrics from the registry
func LoadComputables(metric string) []Computable {
	return registry[metric]
}
