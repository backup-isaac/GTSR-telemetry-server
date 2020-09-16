package merge

import (
	"os"
	"reflect"
	"testing"
	"time"
)

// TestMergeInfoModel tests reading and writing functionality to a JSON config
// file like merge_info_config.json through the funcs on the Model type
// defined in info.go.
func TestMergeInfoModel(t *testing.T) {
	// Overwrite the value of infoPath after it's initialized in info.go's
	// init() func.
	//
	// infoPath is a global variable in the merge package. That's why this
	// works.
	//
	// Having infoPath be a global variable isn't a great way to organize
	// things, but it's a quick & dirty temporary solution.
	infoPath = "merge_info_config_TEST.json"

	defer os.Remove(infoPath)

	startTimestamp, _ := time.Parse("2006-01-02 15:04:05", "2009-11-07T23:00:00-05:00")
	endTimestamp, _ := time.Parse("2006-01-02 15:04:05", "2009-11-08T00:00:00-05:00")
	m := &Model{
		LastJobStartTimestamp: startTimestamp,
		LastJobEndTimestamp:   endTimestamp,
		LastJobBlockNumber:    42,
		DidLastJobFinish:      false,
	}
	err := m.Commit()
	if err != nil {
		t.Fatalf("Error committing initial model: %+v", err)
	}

	m1, err := ReadMergeInfoModel()
	if err != nil {
		t.Fatalf("Error reading model: %+v", err)
	}
	if !reflect.DeepEqual(m, m1) {
		t.Errorf("Read model does not match written:\nwant: %+v\ngot: %+v", m, m1)
	}

	m.LastJobBlockNumber = 69
	m.DidLastJobFinish = true
	m.Commit()

	m2, err := ReadMergeInfoModel()
	if err != nil {
		t.Fatalf("Error reading model: %+v", err)
	}
	if !reflect.DeepEqual(m, m2) {
		t.Errorf("Read model does not match written:\nwant: %+v\ngot: %+v", m, m2)
	}
}
