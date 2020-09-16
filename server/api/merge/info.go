package merge

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"time"
)

const configFileName = "merge_info_config.json"

// Model mirrors the structure of the merge config file. Used to read and edit
// that config file's contents.
type Model struct {
	LastJobStartTimestamp time.Time `json:"lastJobStartTimestamp"`
	LastJobEndTimestamp   time.Time `json:"lastJobEndTimestamp"`
	LastJobBlockNumber    int       `json:"lastJobBlockNumber"`
	DidLastJobFinish      bool      `json:"didLastJobFinish"`
}

var infoPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	infoPath = path.Join(dir, configFileName)
}

// ReadMergeInfoModel marshals the info in the merge config file into a Model
// object.
func ReadMergeInfoModel() (*Model, error) {
	configFile, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return &Model{}, nil
	}
	m := &Model{}
	err = json.Unmarshal(configFile, m)
	return m, err
}

// Commit commits changes to the merge config file, finalizing a transaction.
func (m *Model) Commit() error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(infoPath, bytes, 0644)
}
