package configs

import (
	"server/storage"
	"testing"
)

func TestConfigs(t *testing.T) {
	// Test that CAN configs can be loaded without error
	configs, err := LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	existingNames := make(map[string]bool)
	// Check that there are no illegal metric names in the config
	for _, configList := range configs {
		for _, config := range configList {
			if !storage.ValidMetric(config.Name) {
				t.Errorf("Illegal metric name: %v", config.Name)
			}
			if existingNames[config.Name] {
				t.Errorf("Found duplicate metric name in CAN configs")
			}
			existingNames[config.Name] = true
		}
	}
}
