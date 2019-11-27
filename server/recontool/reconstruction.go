package recontool

// AnalysisResult contains the results of ReconTool analysis
type AnalysisResult struct {
	// RawValues are the raw data provided to ReconTool. nil if plots of raw values were not requested
	RawValues map[string][]float64 `json:"raw_values"`
	// RawTimestamps are the raw timestamps provided to ReconTool. nil if plots of raw values were not requested
	RawTimestamps []int64 `json:"raw_timestamps"`
}

// RunReconTool runs ReconTool on data provided as a mapping of metrics to
// time series of their values and returns computed values
func RunReconTool(data map[string][]float64, rawTimestamps []int64, vehicle *Vehicle, gpsTerrain, plotAll bool) (*AnalysisResult, error) {
	result := AnalysisResult{}
	if plotAll {
		result.RawValues = data
		result.RawTimestamps = rawTimestamps
	}
	return &result, nil
}
