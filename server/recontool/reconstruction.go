package recontool

import (
	"fmt"
)

// AnalysisResult contains the results of ReconTool analysis
type AnalysisResult struct {
	MaxTorque              float64              `json:"max_torque"`
	RawValues              map[string][]float64 `json:"raw_values"`
	RawTimestamps          []int64              `json:"raw_timestamps"`
	TimeMinutes            []float64            `json:"time_min"`
	VelocityMph            []float64            `json:"velocity_mph"`
	DistanceMiles          []float64            `json:"distance_mi"`
	Acceleration           []float64            `json:"acceleration"`
	ModelDerivedTorque     []float64            `json:"model_derived_torque"`
	MotorPower             []float64            `json:"motor_power"`
	ModelDerivedMotorPower []float64            `json:"model_derived_power"`
	BusPower               []float64            `json:"bus_power"`
	//MotorTorque   []float64            `json:"motor_torque"`
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
	busVoltage := Average(data["Left_Bus_Voltage"], data["Right_Bus_Voltage"])
	busCurrent := SumLeftRight(data["Left_Bus_Current"], data["Right_Bus_Current"])
	busPowerSeries := CalculateBusPowerSeries(data["Left_Bus_Voltage"], data["Right_Bus_Voltage"], data["Left_Bus_Current"], data["Right_Bus_Current"])
	result.MaxTorque = vehicle.TMax
	// we keep telemetry from phase B as well, should probably use it
	phaseCurrentSeries := SumLeftRight(data["Left_Phase_C_Current"], data["Right_Phase_C_Current"])
	velocitySeries := CalculateVelocitySeries(rpm, vehicle.RMot)
	distanceSeries := RiemannSumIntegrate(velocitySeries, dt)
	accelerationSeries := Gradient(velocitySeries, dt)
	motorTorqueSeries := CalculateMotorTorqueSeries(rpm, phaseCurrentSeries, vehicle.TMax)
	terrainAngleSeries := DeriveTerrainAngleSeries(motorTorqueSeries, velocitySeries, accelerationSeries, vehicle)
	resultantForceSeries := DeriveMotorForceSeries(velocitySeries, accelerationSeries, terrainAngleSeries, vehicle)
	modelDerivedTorqueSeries := Scale(resultantForceSeries, vehicle.RMot)
	motorControllerEfficiencySeries := CalculateMotorControllerEfficiencySeries(phaseCurrentSeries, busVoltage, busPowerSeries)
	packResistance := CalculatePackResistance(busCurrent, busVoltage)
	packEfficiencySeries := CalculatePackEfficiencySeries(busCurrent, busPowerSeries, packResistance)
	motorEfficiencySeries := CalculateMotorEfficiencySeries(busVoltage, motorTorqueSeries)
	drivetrainEfficiencySeries := CalculateDrivetrainEfficiencySeries(motorControllerEfficiencySeries, packEfficiencySeries, motorEfficiencySeries)
	motorPowerSeries := CalculateMotorPowerSeries(motorTorqueSeries, velocitySeries, phaseCurrentSeries, drivetrainEfficiencySeries, vehicle)
	modelDerivedPowerSeries := CalculateModelDerivedPowerSeries(resultantForceSeries, velocitySeries, drivetrainEfficiencySeries)
	result.VelocityMph = Scale(velocitySeries, MetersPerSecondToMilesPerHour)
	result.DistanceMiles = Scale(distanceSeries, MetersToMiles)
	result.TimeMinutes = Scale(timeSeries, SecondsToMinutes)
	result.Acceleration = accelerationSeries
	//result.MotorTorque = motorTorqueSeries
	result.ModelDerivedTorque = modelDerivedTorqueSeries
	result.MotorPower = motorPowerSeries
	result.ModelDerivedMotorPower = modelDerivedPowerSeries
	result.BusPower = busPowerSeries
	return &result, nil
}
