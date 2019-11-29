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
	MotorTorque   []float64            `json:"motor_torque"`
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
	phaseCurrentSeries := SumLeftRight(data["Left_Phase_C_Current"], data["Right_Phase_C_Current"])
	velocitySeries := CalculateVelocitySeries(rpm, vehicle.RMot)
	distanceSeries := RiemannSumIntegrate(velocitySeries, dt)
	accelerationSeries := Gradient(velocitySeries, dt)
	motorTorqueSeries := CalculateMotorTorqueSeries(rpm, phaseCurrentSeries)
	result.VelocityMph = Scale(velocitySeries, MetersPerSecondToMilesPerHour)
	result.DistanceMiles = Scale(distanceSeries, MetersToMiles)
	result.TimeMinutes = Scale(timeSeries, SecondsToMinutes)
	result.Acceleration = accelerationSeries
	result.MotorTorque = motorTorqueSeries
	return &result, nil
}
