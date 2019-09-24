package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
)

// CanConfigType holds CAN configuration information
type CanConfigType struct {
	CanID       int     `json:"can_id"`
	Datatype    string  `json:"datatype"`
	Name        string  `json:"name"`
	Offset      int     `json:"offset"`
	CheckBounds bool    `json:"check_bounds"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	Description string  `json:"description"`
}

// LoadConfigs loads the CAN configs from the config file
func LoadConfigs() (map[int][]*CanConfigType, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("Could not find runtime caller")
	}
	dir := path.Join(path.Dir(filename), "can_configs")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var canConfigList []CanConfigType

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			rawJSON, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}

			var tmpConfigList []CanConfigType
			err = json.Unmarshal(rawJSON, &tmpConfigList)
			if err != nil {
				return nil, err
			}
			canConfigList = append(canConfigList, tmpConfigList...)
		}
	}

	canDatatypes := make(map[int][]*CanConfigType)
	for i := range canConfigList {
		config := &canConfigList[i]
		canDatatypes[config.CanID] = append(canDatatypes[config.CanID], config)
	}
	return canDatatypes, nil
}
