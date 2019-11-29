package recontool

import (
	"fmt"
)

// AnalysisResult contains the results of ReconTool analysis
type AnalysisResult struct {
	RawValues     map[string][]float64 `json:"raw_values"`
	RawTimestamps []int64              `json:"raw_timestamps"`
	TimeMinutes   []float64            `json:"time_min"`
	VelocityMph   []float64            `json:"velocity_mph"`
	DistanceMiles []float64            `json:"distance_mi"`
	Acceleration  []float64            `json:"acceleration"`
}

// RunReconTool runs ReconTool on data provided as a mapping of metrics to
// time series of their values and returns computed values
func RunReconTool(data map[string][]float64, rawTimestamps []int64, vehicle *Vehicle, gpsTerrain, plotAll bool) (*AnalysisResult, error) {
	result := AnalysisResult{}
	if plotAll {
		result.RawValues = data
		result.RawTimestamps = rawTimestamps
	}
	timeSeries := RemoveTimeOffsets(rawTimestamps)
	dt := timeSeries[1] - timeSeries[0]
	if dt == 0 {
		return nil, fmt.Errorf("Invalid time data %v", rawTimestamps)
	}
	rpm := Average(data["Left_Wavesculptor_RPM"], data["Right_Wavesculptor_RPM"])
	if len(rpm) < 2 {
		return nil, fmt.Errorf("At least 2 data points required")
	}
	velocitySeries := CalculateVelocitySeries(rpm, vehicle.RMot)
	distanceSeries := CalculateDistanceSeries(velocitySeries, dt)
	accelerationSeries := CalculateAccelerationSeries(velocitySeries, dt)
	result.VelocityMph = MetersPerSecondToMilesPerHour(velocitySeries)
	result.DistanceMiles = MetersToMiles(distanceSeries)
	result.TimeMinutes = SecondsToMinutes(timeSeries)
	result.Acceleration = accelerationSeries
	return &result, nil
}
