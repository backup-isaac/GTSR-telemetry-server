package track

import (
	"os"
	"reflect"
	"testing"
)

func TestTrackInfoModel(t *testing.T) {
	infoPath = "track_info_config_TEST.json"
	defer os.Remove(infoPath)
	m := &Model{
		IsTrackInfoNew:      true,
		IsTrackInfoUploaded: false,
		PointNumber:         5,
	}
	err := m.Commit()
	if err != nil {
		t.Fatalf("Error commiting initial model: %+v", err)
	}
	m1, err := ReadTrackInfoModel()
	if err != nil {
		t.Fatalf("Error reading model: %+v", err)
	}
	if !reflect.DeepEqual(m, m1) {
		t.Errorf("Read model does not match written: want %+v, got %+v", m, m1)
	}
	m.IsTrackInfoNew = false
	m.PointNumber = 10
	m.Commit()
	m2, err := ReadTrackInfoModel()
	if err != nil {
		t.Fatalf("Error reading model: %+v", err)
	}
	if !reflect.DeepEqual(m, m2) {
		t.Errorf("Read model does not match written: want %+v, got %+v", m, m2)
	}
}
