package recontool

// MetricHeaderNames maps uploaded CSV headers to server metrics
var MetricHeaderNames = map[string]string{
	"BMS Current":              "BMS_Current",
	"BMS Voltage 1":            "Cell_Voltage_1",
	"BMS Voltage 2":            "Cell_Voltage_2",
	"BMS Voltage 3":            "Cell_Voltage_3",
	"BMS Voltage 4":            "Cell_Voltage_4",
	"BMS Voltage 5":            "Cell_Voltage_5",
	"BMS Voltage 6":            "Cell_Voltage_6",
	"BMS Voltage 7":            "Cell_Voltage_7",
	"BMS Voltage 8":            "Cell_Voltage_8",
	"BMS Voltage 9":            "Cell_Voltage_9",
	"BMS Voltage 10":           "Cell_Voltage_10",
	"BMS Voltage 11":           "Cell_Voltage_11",
	"BMS Voltage 12":           "Cell_Voltage_12",
	"BMS Voltage 13":           "Cell_Voltage_13",
	"BMS Voltage 14":           "Cell_Voltage_14",
	"BMS Voltage 15":           "Cell_Voltage_15",
	"BMS Voltage 16":           "Cell_Voltage_16",
	"BMS Voltage 17":           "Cell_Voltage_17",
	"BMS Voltage 18":           "Cell_Voltage_18",
	"BMS Voltage 19":           "Cell_Voltage_19",
	"BMS Voltage 20":           "Cell_Voltage_20",
	"BMS Voltage 21":           "Cell_Voltage_21",
	"BMS Voltage 22":           "Cell_Voltage_22",
	"BMS Voltage 23":           "Cell_Voltage_23",
	"BMS Voltage 24":           "Cell_Voltage_24",
	"BMS Voltage 25":           "Cell_Voltage_25",
	"BMS Voltage 26":           "Cell_Voltage_26",
	"BMS Voltage 27":           "Cell_Voltage_27",
	"BMS Voltage 28":           "Cell_Voltage_28",
	"BMS Voltage 29":           "Cell_Voltage_29",
	"BMS Voltage 30":           "Cell_Voltage_30",
	"BMS Voltage 31":           "Cell_Voltage_31",
	"BMS Voltage 32":           "Cell_Voltage_32",
	"BMS Voltage 33":           "Cell_Voltage_33",
	"BMS Voltage 34":           "Cell_Voltage_34",
	"BMS Voltage 35":           "Cell_Voltage_35",
	"Right WS Current":         "Right_Bus_Current",
	"Right WS Voltage":         "Right_Bus_Voltage",
	"Right WS RPM":             "Right_Wavesculptor_RPM",
	"Right WS Phase C Current": "Right_Phase_C_Current",
	"Right WS Charge Consumed": "Right_Charge_Consumed",
	"Left WS Current":          "Left_Bus_Current",
	"Left WS Voltage":          "Left_Bus_Voltage",
	"Left WS RPM":              "Left_Wavesculptor_RPM",
	"Left WS Phase C Current":  "Left_Phase_C_Current",
	"Left WS Charge Consumed":  "Left_Charge_Consumed",
	"Throttle":                 "Throttle",
	"GPS Latitude":             "GPS_Latitude",
	"GPS Longitude":            "GPS_Longitude",
	"Left MPPT Current":        "Photon_Channel_0_Array_Current",
	"Left MPPT Voltage":        "Photon_Channel_0_Array_Voltage",
	"Right MPPT Current":       "Photon_Channel_1_Array_Current",
	"Right MPPT Voltage":       "Photon_Channel_1_Array_Voltage",
}

// MetricNames holds the names of metrics ReconTool needs to query from the server
var MetricNames []string

func init() {
	MetricNames := make([]string, len(MetricHeaderNames))
	i := 0
	for _, v := range MetricHeaderNames {
		MetricNames[i] = v
		i++
	}
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
