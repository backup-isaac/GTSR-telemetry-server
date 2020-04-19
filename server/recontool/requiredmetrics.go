package recontool

import "fmt"

// LoggerMetricHeaders maps uploaded CSV headers to server metrics
var LoggerMetricHeaders = map[string]string{
	"BMS Current":              "BMS_Current",
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

// TimeHeaderName is the CSV header for time
const TimeHeaderName = "Millis"

// MetricNames holds the names of metrics ReconTool needs to query from the server
var MetricNames = []string{
	"BMS_Current",
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

func init() {
	for i := 1; i <= 35; i++ {
		loggerName := fmt.Sprintf("BMS Voltage %d", i)
		metricName := fmt.Sprintf("Cell_Voltage_%d", i)
		LoggerMetricHeaders[loggerName] = metricName
		MetricNames = append(MetricNames, metricName)
	}
}
