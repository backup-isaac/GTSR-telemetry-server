package recontool

import (
	"fmt"
)

// AnalysisResult contains the results of ReconTool analysis
type AnalysisResult struct {
	MaxTorque              float64              `json:"max_torque"`
	PackCapacity           float64              `json:"pack_capacity"`
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
	SimulatedTotalCharge   []float64            `json:"simulated_total_charge"`
	SimulatedNetCharge     []float64            `json:"simulated_net_charge"`
	MeasuredTotalCharge    []float64            `json:"measured_total_charge"`
	MeasuredNetCharge      []float64            `json:"measured_net_charge"`
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
	busPowerSeries := CalculateSeries(func(params ...float64) float64 {
		return BusPower(params[0], params[1], params[2], params[3])
	}, data["Left_Bus_Voltage"], data["Right_Bus_Voltage"], data["Left_Bus_Current"], data["Right_Bus_Current"])
	result.MaxTorque = vehicle.TMax
	result.PackCapacity = vehicle.QMax
	// we keep telemetry from phase B as well, should probably use it
	phaseCurrentSeries := SumLeftRight(data["Left_Phase_C_Current"], data["Right_Phase_C_Current"])
	velocitySeries := CalculateSeries(func(params ...float64) float64 {
		return Velocity(params[0], vehicle.RMot)
	}, rpm)
	distanceSeries := RiemannSumIntegrate(velocitySeries, dt)
	accelerationSeries := Gradient(velocitySeries, dt)
	motorTorqueSeries := CalculateSeries(func(params ...float64) float64 {
		return MotorTorque(params[0], params[1], vehicle.TMax)
	}, rpm, phaseCurrentSeries)
	terrainAngleSeries := DeriveTerrainAngleSeries(motorTorqueSeries, velocitySeries, accelerationSeries, vehicle)
	resultantForceSeries := CalculateSeries(func(params ...float64) float64 {
		return ModeledMotorForce(params[0], params[1], params[2], vehicle)
	}, velocitySeries, accelerationSeries, terrainAngleSeries)
	modelDerivedTorqueSeries := Scale(resultantForceSeries, vehicle.RMot)
	motorControllerEfficiencySeries := CalculateSeries(func(params ...float64) float64 {
		return MotorControllerEfficiency(params[0], params[1], meanIf(busPowerSeries, func(p float64) bool { return p > 0 }))
	}, phaseCurrentSeries, busVoltage)
	packResistance := PackResistance(busCurrent, busVoltage)
	packEfficiencySeries := CalculateSeries(func(params ...float64) float64 {
		return PackEfficiency(params[0], params[1], packResistance)
	}, busCurrent, busPowerSeries)
	motorEfficiencySeries := CalculateSeries(func(params ...float64) float64 {
		return MotorEfficiency(params[0], params[1])
	}, busVoltage, motorTorqueSeries)
	drivetrainEfficiencySeries := CalculateSeries(func(params ...float64) float64 {
		return DrivetrainEfficiency(params[0], params[1], params[2])
	}, motorControllerEfficiencySeries, packEfficiencySeries, motorEfficiencySeries)
	motorPowerSeries := CalculateSeries(func(params ...float64) float64 {
		return MotorPower(params[0], params[1], params[2], params[3], vehicle)
	}, motorTorqueSeries, velocitySeries, phaseCurrentSeries, drivetrainEfficiencySeries)
	modelDerivedPowerSeries := CalculateSeries(func(params ...float64) float64 {
		return ModelDerivedPower(params[0], params[1], params[2])
	}, resultantForceSeries, velocitySeries, drivetrainEfficiencySeries)
	modelDerivedCurrentSeries := CalculateSeries(func(params ...float64) float64 {
		if params[1] == 0 {
			return 0
		}
		return params[0] / params[1]
	}, modelDerivedPowerSeries, busVoltage)
	simulatedTotalChargeSeries := RiemannSumIntegrate(modelDerivedCurrentSeries, dt/3600)
	simulatedNetChargeSeries := RiemannSumIntegrate(data["BMS_Current"], dt/3600)
	measuredTotalChargeSeries := RiemannSumIntegrate(busCurrent, dt/3600)
	result.VelocityMph = Scale(velocitySeries, MetersPerSecondToMilesPerHour)
	result.DistanceMiles = Scale(distanceSeries, MetersToMiles)
	result.TimeMinutes = Scale(timeSeries, SecondsToMinutes)
	//result.TimeMinutes = floats.ScaleTo(make([]float64, timeSeries), SecondsToMinutes, timeSeries)
	result.Acceleration = accelerationSeries
	//result.MotorTorque = motorTorqueSeries
	result.ModelDerivedTorque = modelDerivedTorqueSeries
	result.MotorPower = motorPowerSeries
	result.ModelDerivedMotorPower = modelDerivedPowerSeries
	result.BusPower = busPowerSeries
	result.SimulatedTotalCharge = simulatedTotalChargeSeries
	result.SimulatedNetCharge = simulatedNetChargeSeries
	result.MeasuredTotalCharge = measuredTotalChargeSeries
	result.MeasuredNetCharge = SumLeftRight(data["Left_Charge_Consumed"], data["Right_Charge_Consumed"])
	return &result, nil
}
