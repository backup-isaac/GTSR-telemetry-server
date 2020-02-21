package api

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"server/recontool"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTimeRangeParams(t *testing.T) {
	req := timeRangeRequest()
	goodForm := req.Form
	req.Form = nil
	_, err := parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form = goodForm
	req.Form.Del("startDate")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("startDate", "foo")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("startDate", "3133690620000")
	req.Form.Set("endDate", "foo")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("endDate", "3133691220000")
	req.Form.Set("resolution", "foo")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("resolution", "0")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("resolution", "500")
	req.Form.Set("terrain", "500")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("terrain", "false")
	req.Form.Set("Rmot", "foo")
	_, err = parseTimeRangeParams(req)
	assert.Error(t, err)
	req.Form.Set("Rmot", "0.3")
	params, err := parseTimeRangeParams(req)
	assert.NoError(t, err)
	assert.Equal(t, &timeRangeParams{
		start:      time.Unix(3133690620, 0),
		end:        time.Unix(3133691220, 0),
		resolution: 500,
		gps:        false,
		vehicle: &recontool.Vehicle{
			RMot:  0.3,
			M:     320,
			Crr1:  0.007,
			Crr2:  0.0005,
			CDa:   0.14,
			TMax:  100,
			QMax:  30,
			RLine: 0.07,
			VcMax: 4.15,
			VcMin: 2.65,
			VSer:  35,
		},
	}, params)
}

func timeRangeRequest() *http.Request {
	form := url.Values{}
	form.Set("startDate", "3133690620000")
	form.Set("endDate", "3133691220000")
	form.Set("resolution", "500")
	form.Set("terrain", "false")
	form.Set("Rmot", "0.3")
	form.Set("m", "320")
	form.Set("Crr1", "0.007")
	form.Set("Crr2", "0.0005")
	form.Set("CDa", "0.14")
	form.Set("Tmax", "100")
	form.Set("Qmax", "30")
	form.Set("Rline", "0.07")
	form.Set("VcMax", "4.15")
	form.Set("VcMin", "2.65")
	form.Set("Vser", "35")
	return &http.Request{
		Form: form,
	}
}

func TestMakeTimestamps(t *testing.T) {
	start := time.Unix(3133690620, 0)
	end := time.Unix(3133690630, 0)
	resolution := 2000
	actual := makeTimestamps(start, end, resolution)
	assert.Equal(t, []int64{
		3133690620000, 3133690622000, 3133690624000, 3133690626000, 3133690628000,
	}, actual)
}

func TestExtractVehicleForm(t *testing.T) {
	form := &timeRangeRequest().Form
	for _, s := range []string{
		"Rmot",
		"m",
		"Crr1",
		"Crr2",
		"CDa",
		"Tmax",
		"Qmax",
		"Rline",
		"VcMax",
		"VcMin",
		"Vser",
	} {
		goodVal := form.Get(s)
		form.Del(s)
		_, err := extractVehicleForm(form)
		assert.Error(t, err)
		form.Set(s, "foo")
		_, err = extractVehicleForm(form)
		assert.Error(t, err)
		form.Set(s, goodVal)
	}
	form.Set("Vser", "-1")
	_, err := extractVehicleForm(form)
	assert.Error(t, err)
	form.Set("Vser", "35")
	vehicle, err := extractVehicleForm(form)
	assert.NoError(t, err)
	assert.Equal(t, &recontool.Vehicle{
		RMot:  0.3,
		M:     320,
		Crr1:  0.007,
		Crr2:  0.0005,
		CDa:   0.14,
		TMax:  100,
		QMax:  30,
		RLine: 0.07,
		VcMax: 4.15,
		VcMin: 2.65,
		VSer:  35,
	}, vehicle)
}

func TestParseCsvParams(t *testing.T) {
	req := csvRequest()
	goodForm := req.MultipartForm
	req.MultipartForm = nil
	_, err := parseCsvParams(req)
	assert.Error(t, err)
	req.MultipartForm = goodForm
	for _, s := range []string{"terrain", "autoPlots", "compileFiles"} {
		goodVal := req.MultipartForm.Value[s]
		delete(req.MultipartForm.Value, s)
		_, err = parseCsvParams(req)
		assert.Error(t, err)
		req.MultipartForm.Value[s] = []string{"foo"}
		_, err = parseCsvParams(req)
		assert.Error(t, err)
		req.MultipartForm.Value[s] = goodVal
	}
	goodFile := req.MultipartForm.File
	req.MultipartForm.File = map[string][]*multipart.FileHeader{}
	_, err = parseCsvParams(req)
	assert.Error(t, err)
	req.MultipartForm.File = goodFile
	req.MultipartForm.Value["Vser"] = []string{"foo"}
	_, err = parseCsvParams(req)
	assert.Error(t, err)
	req.MultipartForm.Value["Vser"] = []string{"35"}
	params, err := parseCsvParams(req)
	assert.NoError(t, err)
	assert.Equal(t, &csvParams{
		gps:          false,
		plotAll:      false,
		combineFiles: false,
		numCsvs:      1,
		vehicle: &recontool.Vehicle{
			RMot:  0.3,
			M:     320,
			Crr1:  0.007,
			Crr2:  0.0005,
			CDa:   0.14,
			TMax:  100,
			QMax:  30,
			RLine: 0.07,
			VcMax: 4.15,
			VcMin: 2.65,
			VSer:  35,
		},
	}, params)
}

func csvRequest() *http.Request {
	return &http.Request{
		MultipartForm: &multipart.Form{
			Value: map[string][]string{
				"terrain":      {"false"},
				"autoPlots":    {"false"},
				"compileFiles": {"false"},
				"Rmot":         {"0.3"},
				"m":            {"320"},
				"Crr1":         {"0.007"},
				"Crr2":         {"0.0005"},
				"CDa":          {"0.14"},
				"Tmax":         {"100"},
				"Qmax":         {"30"},
				"Rline":        {"0.07"},
				"VcMax":        {"4.15"},
				"VcMin":        {"2.65"},
				"Vser":         {"35"},
			},
			File: map[string][]*multipart.FileHeader{
				"": {nil},
			},
		},
	}
}

func TestExtractVehicleMultipart(t *testing.T) {
	form := csvRequest().MultipartForm
	for _, s := range []string{
		"Rmot",
		"m",
		"Crr1",
		"Crr2",
		"CDa",
		"Tmax",
		"Qmax",
		"Rline",
		"VcMax",
		"VcMin",
		"Vser",
	} {
		goodVal := form.Value[s]
		delete(form.Value, s)
		_, err := extractVehicleMultipart(form)
		assert.Error(t, err)
		form.Value[s] = []string{"foo"}
		_, err = extractVehicleMultipart(form)
		assert.Error(t, err)
		form.Value[s] = goodVal
	}
	form.Value["Vser"] = []string{"-1"}
	_, err := extractVehicleMultipart(form)
	assert.Error(t, err)
	form.Value["Vser"] = []string{"35"}
	vehicle, err := extractVehicleMultipart(form)
	assert.NoError(t, err)
	assert.Equal(t, &recontool.Vehicle{
		RMot:  0.3,
		M:     320,
		Crr1:  0.007,
		Crr2:  0.0005,
		CDa:   0.14,
		TMax:  100,
		QMax:  30,
		RLine: 0.07,
		VcMax: 4.15,
		VcMin: 2.65,
		VSer:  35,
	}, vehicle)
}

func TestReadUploadedCsv(t *testing.T) {
	assert.NoError(t, populateTestCsvs())
	channel := make(chan csvParse)
	emptyFile, err := os.Open("empty.csv")
	defer emptyFile.Close()
	assert.NoError(t, err)
	go readUploadedCsv(emptyFile, 0, false, channel)
	csvParse := <-channel
	assert.Error(t, csvParse.err)
	badHeaders, err := os.Open("missing_headers_only.csv")
	defer badHeaders.Close()
	assert.NoError(t, err)
	go readUploadedCsv(badHeaders, 880, false, channel)
	csvParse = <-channel
	assert.Error(t, csvParse.err)
	headersOnly, err := os.Open("required_headers_only.csv")
	defer headersOnly.Close()
	assert.NoError(t, err)
	go readUploadedCsv(headersOnly, 906, false, channel)
	csvParse = <-channel
	assert.NoError(t, csvParse.err)
	assert.Equal(t, []int64{}, csvParse.timestamps)
	assert.Equal(t, csvColumnsOf([]float64{}, []string{}, map[string]float64{}), csvParse.csvData)
	invalidTime, err := os.Open("invalid_timestamp.csv")
	defer invalidTime.Close()
	assert.NoError(t, err)
	go readUploadedCsv(invalidTime, 1025, false, channel)
	csvParse = <-channel
	assert.Error(t, csvParse.err)
	invalidData, err := os.Open("invalid_data.csv")
	defer invalidData.Close()
	assert.NoError(t, err)
	go readUploadedCsv(invalidData, 1025, false, channel)
	csvParse = <-channel
	assert.Error(t, csvParse.err)
	valid, err := os.Open("valid_data.csv")
	defer valid.Close()
	assert.NoError(t, err)
	go readUploadedCsv(valid, 1603, false, channel)
	csvParse = <-channel
	assert.NoError(t, csvParse.err)
	assert.Equal(t, []int64{0, 0, 0, 0, 0, 0}, csvParse.timestamps)
	assert.Equal(t, csvColumnsOf([]float64{0, 0, 0, 0, 0, 0}, []string{}, map[string]float64{}), csvParse.csvData)
	uneven, err := os.Open("uneven_data.csv")
	defer uneven.Close()
	assert.NoError(t, err)
	go readUploadedCsv(uneven, 1606, false, channel)
	csvParse = <-channel
	assert.Error(t, csvParse.err)
}

func populateTestCsvs() error {
	goodHeaders := "BMS Current,BMS Voltage 1,BMS Voltage 10,BMS Voltage 11,BMS Voltage 12,BMS Voltage 13,BMS Voltage 14,BMS Voltage 15,BMS Voltage 16,BMS Voltage 17,BMS Voltage 18,BMS Voltage 19,BMS Voltage 2,BMS Voltage 20,BMS Voltage 21,BMS Voltage 22,BMS Voltage 23,BMS Voltage 24,BMS Voltage 25,BMS Voltage 26,BMS Voltage 27,BMS Voltage 28,BMS Voltage 29,BMS Voltage 3,BMS Voltage 30,BMS Voltage 31,BMS Voltage 32,BMS Voltage 33,BMS Voltage 34,BMS Voltage 35,BMS Voltage 36,BMS Voltage 38,BMS Voltage 4,BMS Voltage 5,BMS Voltage 6,BMS Voltage 7,BMS Voltage 8,BMS Voltage 9,GPS Latitude,GPS Longitude,Left Battery Voltage,Left MPPT Current,Left MPPT Voltage,Left WS Charge Consumed,Left WS Current,Left WS Phase C Current,Left WS RPM,Left WS Voltage,Millis,Right Battery Voltage,Right MPPT Current,Right MPPT Voltage,Right WS Charge Consumed,Right WS Current,Right WS Phase C Current,Right WS RPM,Right WS Voltage,Throttle"
	invalidTimestamps := goodHeaders + "\n"
	invalidData := goodHeaders + "\n"
	validData := goodHeaders + "\n"
	zeroRow := ""
	for _, s := range strings.Split(goodHeaders, ",") {
		if s == "Millis" {
			invalidTimestamps += "foo,"
		} else {
			invalidTimestamps += "0,"
		}
		if s == "Left WS Charge Consumed" {
			invalidData += "foo,"
		} else {
			invalidData += "0,"
		}
		zeroRow += "0,"
	}
	for i := 0; i < 6; i++ {
		validData += zeroRow[:len(zeroRow)-1] + "\n"
	}
	testCsvContents := map[string]string{
		"empty":                 "",
		"required_headers_only": goodHeaders,
		"missing_headers_only":  "BMS Voltage 10,BMS Voltage 11,BMS Voltage 12,BMS Voltage 13,BMS Voltage 14,BMS Voltage 15,BMS Voltage 16,BMS Voltage 17,BMS Voltage 18,BMS Voltage 19,BMS Voltage 2,BMS Voltage 20,BMS Voltage 21,BMS Voltage 22,BMS Voltage 23,BMS Voltage 24,BMS Voltage 25,BMS Voltage 26,BMS Voltage 27,BMS Voltage 28,BMS Voltage 29,BMS Voltage 3,BMS Voltage 30,BMS Voltage 31,BMS Voltage 32,BMS Voltage 33,BMS Voltage 34,BMS Voltage 35,BMS Voltage 36,BMS Voltage 38,BMS Voltage 4,BMS Voltage 5,BMS Voltage 6,BMS Voltage 7,BMS Voltage 8,BMS Voltage 9,GPS Latitude,GPS Longitude,Left Battery Voltage,Left MPPT Current,Left MPPT Voltage,Left WS Charge Consumed,Left WS Current,Left WS Phase C Current,Left WS RPM,Left WS Voltage,Millis,Right Battery Voltage,Right MPPT Current,Right MPPT Voltage,Right WS Charge Consumed,Right WS Current,Right WS Phase C Current,Right WS RPM,Right WS Voltage,Throttle",
		"invalid_timestamp":     invalidTimestamps,
		"invalid_data":          invalidData,
		"valid_data":            validData,
		"uneven_data":           validData + "0,0",
	}
	for name, contents := range testCsvContents {
		err := writeFile(name, contents)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(name, contents string) error {
	file, err := os.Create(fmt.Sprintf("%s.csv", name))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(contents)
	return err
}

func csvColumnsOf(col []float64, additionalCols []string, additionalElements map[string]float64) map[string][]float64 {
	ret := map[string][]float64{
		"BMS_Current":                    col,
		"Cell_Voltage_1":                 col,
		"Cell_Voltage_10":                col,
		"Cell_Voltage_11":                col,
		"Cell_Voltage_12":                col,
		"Cell_Voltage_13":                col,
		"Cell_Voltage_14":                col,
		"Cell_Voltage_15":                col,
		"Cell_Voltage_16":                col,
		"Cell_Voltage_17":                col,
		"Cell_Voltage_18":                col,
		"Cell_Voltage_19":                col,
		"Cell_Voltage_2":                 col,
		"Cell_Voltage_20":                col,
		"Cell_Voltage_21":                col,
		"Cell_Voltage_22":                col,
		"Cell_Voltage_23":                col,
		"Cell_Voltage_24":                col,
		"Cell_Voltage_25":                col,
		"Cell_Voltage_26":                col,
		"Cell_Voltage_27":                col,
		"Cell_Voltage_28":                col,
		"Cell_Voltage_29":                col,
		"Cell_Voltage_3":                 col,
		"Cell_Voltage_30":                col,
		"Cell_Voltage_31":                col,
		"Cell_Voltage_32":                col,
		"Cell_Voltage_33":                col,
		"Cell_Voltage_34":                col,
		"Cell_Voltage_35":                col,
		"Cell_Voltage_4":                 col,
		"Cell_Voltage_5":                 col,
		"Cell_Voltage_6":                 col,
		"Cell_Voltage_7":                 col,
		"Cell_Voltage_8":                 col,
		"Cell_Voltage_9":                 col,
		"GPS_Latitude":                   col,
		"GPS_Longitude":                  col,
		"Photon_Channel_0_Array_Current": col,
		"Photon_Channel_0_Array_Voltage": col,
		"Left_Charge_Consumed":           col,
		"Left_Bus_Current":               col,
		"Left_Phase_C_Current":           col,
		"Left_Wavesculptor_RPM":          col,
		"Left_Bus_Voltage":               col,
		"Photon_Channel_1_Array_Current": col,
		"Photon_Channel_1_Array_Voltage": col,
		"Right_Charge_Consumed":          col,
		"Right_Bus_Current":              col,
		"Right_Phase_C_Current":          col,
		"Right_Wavesculptor_RPM":         col,
		"Right_Bus_Voltage":              col,
		"Throttle":                       col,
	}
	for _, s := range additionalCols {
		ret[s] = col
	}
	for k, v := range additionalElements {
		ret[k] = append(ret[k], v)
	}
	return ret
}

func TestParseColumnNames(t *testing.T) {
	parseColumnNamesRunner(t, true, map[string]string{}, map[string]bool{}, false)
	parseColumnNamesRunner(t, true, map[string]string{
		"Test 1": "    ",
		"Test 2": " ",
	}, map[string]bool{}, false)
	parseColumnNamesRunner(t, true, map[string]string{
		"Millis": "          ",
	}, map[string]bool{}, false)
	parseColumnNamesRunner(t, false, map[string]string{}, map[string]bool{"Millis": true}, false)
	parseColumnNamesRunner(t, false, map[string]string{
		"Millis": "   ",
	}, map[string]bool{"Millis": true}, false)
	parseColumnNamesRunner(t, false, map[string]string{}, map[string]bool{"Test 0": true}, false)
	parseColumnNamesRunner(t, true, map[string]string{}, map[string]bool{
		"    ": false,
	}, false)
	parseColumnNamesRunner(t, true, map[string]string{}, map[string]bool{
		"    ": false,
	}, true)
	parseColumnNamesRunner(t, true, map[string]string{}, map[string]bool{
		"Test 3": false,
	}, true)
	parseColumnNamesRunner(t, true, map[string]string{}, map[string]bool{
		"Test 3": false,
	}, false)
	parseColumnNamesRunner(t, false, map[string]string{
		"Test 1": "delete",
	}, map[string]bool{}, false)
	parseColumnNamesRunner(t, false, map[string]string{
		"Test 1": "Foo",
	}, map[string]bool{}, false)
	parseColumnNamesRunner(t, false, map[string]string{
		"Test 1": "delete",
	}, map[string]bool{
		"Test 0": true,
	}, false)
	parseColumnNamesRunner(t, false, map[string]string{
		"Millis": "delete",
	}, map[string]bool{}, false)
}

func parseColumnNamesRunner(t *testing.T, expectedSuccess bool, toPrepend map[string]string, extra map[string]bool, plotAll bool) {
	loggerToServerMapping := map[string]string{
		"Test 0": "Test_0",
		"Test 1": "Test_1",
		"Test 2": "Test_2",
	}
	columns, expectedParse := createColumnNames(toPrepend, extra, plotAll)
	actualParse, err := parseColumnNames(columns, plotAll, loggerToServerMapping, "Millis")
	if expectedSuccess {
		assert.NoError(t, err)
		assert.Equal(t, expectedParse, actualParse)
	} else {
		assert.Error(t, err)
	}
}

func createColumnNames(toPrepend map[string]string, extra map[string]bool, plotAll bool) ([]string, map[string]int) {
	names := []string{"Test 1", "Test 2", "Test 0"}
	indMap := map[string]int{}
	for i, s := range names {
		prep, ok := toPrepend[s]
		if ok {
			if prep == "delete" {
				copy(names[i:], names[i+1:])
				names = names[:len(names)-1]
			} else {
				names[i] = prep + s
			}
		}
		indMap[strings.Replace(s, " ", "_", 1)] = i
	}
	indMap["time"] = len(names)
	timeName := "Millis"
	prep, ok := toPrepend[timeName]
	if ok {
		timeName = prep + timeName
	}
	names = append(names, timeName)
	duplicates := false
	for s, dup := range extra {
		if plotAll && len(strings.TrimLeft(s, " ")) > 0 {
			indMap[s] = len(names)
		}
		names = append(names, s)
		if dup {
			duplicates = true
		}
	}
	if duplicates {
		return names, nil
	}
	return names, indMap
}

func TestMergeParsedCsvs(t *testing.T) {
	mergeCsvRunner(t, []int{4}, []int{})
	mergeCsvRunner(t, []int{0, 0, 0}, []int{})
	mergeCsvRunner(t, []int{10, 10, 0}, []int{})
	mergeCsvRunner(t, []int{12, 12, 12}, []int{})
	mergeCsvRunner(t, []int{15, 15, 1}, []int{})
	mergeCsvRunner(t, []int{20, 17, 27}, []int{})
	mergeCsvRunner(t, []int{24, 24, 24}, []int{2})
	mergeCsvRunner(t, []int{24, 24, 24}, []int{0, 1, 2})
}

func mergeCsvRunner(t *testing.T, columnLengths []int, csvsWithExtraColumns []int) {
	inputCsvs, inputTimestamps, expectedCsv, expectedTimestamps := createTestParsedCsvs(columnLengths, csvsWithExtraColumns)
	actualCsv, actualTimestamps := mergeParsedCsvs(inputCsvs, inputTimestamps)
	assert.Equal(t, expectedTimestamps, actualTimestamps)
	assert.Equal(t, expectedCsv, actualCsv)
}

func createTestParsedCsvs(columnLengths []int, csvsWithExtraColumns []int) ([]map[string][]float64, [][]int64, map[string][]float64, []int64) {
	inputCsvs := make([]map[string][]float64, len(columnLengths))
	inputTimestamps := make([][]int64, len(columnLengths))
	totalTimestamps := 0
	for _, length := range columnLengths {
		totalTimestamps += length
	}
	expectedTimestamps := make([]int64, totalTimestamps)
	expectedCsv := map[string][]float64{
		"Test 0": make([]float64, totalTimestamps),
		"Test 1": make([]float64, totalTimestamps),
	}
	ttIndex := 0
	for i := 0; i < len(columnLengths); i++ {
		inputCsvs[i] = map[string][]float64{
			"Test 0": make([]float64, columnLengths[i]),
			"Test 1": make([]float64, columnLengths[i]),
		}
		inputTimestamps[i] = make([]int64, columnLengths[i])
		for j := 0; j < columnLengths[i]; j++ {
			inputTimestamps[i][j] = int64(i + 100*j)
			expectedTimestamps[ttIndex] = int64(i + 100*j)
			test0 := 1.2*float64(j) + 0.1*float64(i)
			inputCsvs[i]["Test 0"][j] = test0
			expectedCsv["Test 0"][ttIndex] = test0
			test1 := 1.3*float64(j) + 0.1*float64(i)
			inputCsvs[i]["Test 1"][j] = test1
			expectedCsv["Test 1"][ttIndex] = test1
			ttIndex++
		}
	}
	extraCsvCounter := 2
	for _, c := range csvsWithExtraColumns {
		extraColumn := make([]float64, columnLengths[c])
		for i := 0; i < len(extraColumn); i++ {
			extraColumn[i] = float64(extraCsvCounter * i)
		}
		colName := fmt.Sprintf("Test %d", extraCsvCounter)
		inputCsvs[c][colName] = extraColumn
		expectedCsv[colName] = extraColumn
		extraCsvCounter++
	}
	for _, metricValues := range expectedCsv {
		sort.Slice(metricValues, func(i, j int) bool {
			return metricValues[i]-metricValues[j] < 0
		})
	}
	sort.Slice(expectedTimestamps, func(i, j int) bool {
		return expectedTimestamps[i]-expectedTimestamps[j] < 0
	})
	return inputCsvs, inputTimestamps, expectedCsv, expectedTimestamps
}
