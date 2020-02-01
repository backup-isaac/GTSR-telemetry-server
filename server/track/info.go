package track

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"runtime"
)

// Model mirrors the structure of track_info_config.json. Used to edit
// track_info_config's contents
type Model struct {
	IsTrackInfoNew      bool `json:"isTrackInfoNew"`
	IsTrackInfoUploaded bool `json:"isTrackInfoUploaded"`
	PointNumber         int  `json:"pointNumber"`
}

var infoPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	infoPath = path.Join(dir, "track_info_config.json")
}

// ReadTrackInfoModel the current track info model
func ReadTrackInfoModel() (*Model, error) {
	configFile, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return &Model{}, nil
	}
	m := &Model{}
	err = json.Unmarshal(configFile, m)
	return m, err
}

// Commit commits the track info model transaction
func (m *Model) Commit() error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(infoPath, bytes, 0644)
}
