package recontool

// MetricNames contains the names of metrics that ReconTool needs to use
var MetricNames = []string{
	"BMS_Current",
	"Cell_Voltage_1",
	"Cell_Voltage_2",
	"Cell_Voltage_3",
	"Cell_Voltage_4",
	"Cell_Voltage_5",
	"Cell_Voltage_6",
	"Cell_Voltage_7",
	"Cell_Voltage_8",
	"Cell_Voltage_9",
	"Cell_Voltage_10",
	"Cell_Voltage_11",
	"Cell_Voltage_12",
	"Cell_Voltage_13",
	"Cell_Voltage_14",
	"Cell_Voltage_15",
	"Cell_Voltage_16",
	"Cell_Voltage_17",
	"Cell_Voltage_18",
	"Cell_Voltage_19",
	"Cell_Voltage_20",
	"Cell_Voltage_21",
	"Cell_Voltage_22",
	"Cell_Voltage_23",
	"Cell_Voltage_24",
	"Cell_Voltage_25",
	"Cell_Voltage_26",
	"Cell_Voltage_27",
	"Cell_Voltage_28",
	"Cell_Voltage_29",
	"Cell_Voltage_30",
	"Cell_Voltage_31",
	"Cell_Voltage_32",
	"Cell_Voltage_33",
	"Cell_Voltage_34",
	"Cell_Voltage_35",
	"Right_Bus_Current",
	"Right_Bus_Voltage",
	"Right_Wavesculptor_RPM",
	"Right_Phase_C_Current",
	"Right_Charge_Consumed",
	"Left_Bus_Current",
	"Left_Bus_Voltage",
	"Left_Wavesculptor_RPM",
	"Left_Phase_C_Current",
	"Left_Charge_Consumed",
	"Throttle",
	"GPS_Latitude",
	"GPS_Longitude",
	"Photon_Channel_0_Array_Current",
	"Photon_Channel_0_Array_Voltage",
	"Photon_Channel_1_Array_Current",
	"Photon_Channel_1_Array_Voltage",
}

// Vehicle contains parameters of the vehicle whose data is being analyzed
type Vehicle struct {
	// Motor radius (m)
	RMot float64
	// Mass (kg)
	M float64
	// Crr1 rolling resistance
	Crr1 float64
	// Crr2 dynamic rolling resistance (s/m)
	Crr2 float64
	// Area drag coefficient (m^2)
	CDa float64
	// Maximum motor torque (N-m)
	TMax float64
	// Battery charge capacity (A-hr)
	QMax float64
	// Phase line resistance (Ω)
	RLine float64
	// Maximum battery module voltage (V)
	VcMax float64
	// Minimum battery module voltage (V)
	VcMin float64
	// Number of battery modules in series
	VSer uint
}
