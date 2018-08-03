package canConfigs

import (
	"encoding/json"
	"io/ioutil"
)

// CanConfigType holds CAN configuration information
type CanConfigType struct {
	CanID    int
	Datatype string
	Name     string
	Offset   int
}

// LoadConfigs loads the CAN configs from the config file
func LoadConfigs() (map[int]*CanConfigType, error) {
	rawJSON, err := ioutil.ReadFile("can_config.json")
	if err != nil {
		return nil, err
	}
	var canConfigList []CanConfigType
	err = json.Unmarshal(rawJSON, &canConfigList)
	if err != nil {
		return nil, err
	}
	canDatatypes := make(map[int]*CanConfigType)
	for i := range canConfigList {
		config := &canConfigList[i]
		canDatatypes[config.CanID] = config
	}
	return canDatatypes, nil
}
