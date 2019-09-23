package configs

import (
	"strings"
	"testing"
)

func TestConfigs(t *testing.T) {
	// Test that CAN configs can be loaded without error
	configs, err := LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	for _, configList := range configs {
		for _, config := range configList {
			if strings.ContainsAny(config.Name, " \n\t\r") {
				t.Errorf("Illegal metric name: %v", config.Name)
			}
		}
	}
}
