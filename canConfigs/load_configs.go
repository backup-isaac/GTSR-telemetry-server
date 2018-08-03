package canConfigs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
)

// CanConfigType holds CAN configuration information
type CanConfigType struct {
	CanID       int
	Datatype    string
	Name        string
	Offset      int
	CheckBounds bool
	MinValue    float64
	MaxValue    float64
}

// LoadConfigs loads the CAN configs from the config file
func LoadConfigs() (map[int][]*CanConfigType, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("Could not find runtime caller")
	}
	rawJSON, err := ioutil.ReadFile(path.Join(path.Dir(filename), "can_config.json"))
	if err != nil {
		return nil, err
	}
	var canConfigList []CanConfigType
	err = json.Unmarshal(rawJSON, &canConfigList)
	if err != nil {
		return nil, err
	}
	canDatatypes := make(map[int][]*CanConfigType)
	for i := range canConfigList {
		config := &canConfigList[i]
		canDatatypes[config.CanID] = append(canDatatypes[config.CanID], config)
	}
	return canDatatypes, nil
}
