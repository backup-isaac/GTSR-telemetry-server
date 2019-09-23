package configs

import "testing"

func TestConfigs(t *testing.T) {
	// Test that CAN configs can be loaded without error
	if _, err := LoadConfigs(); err != nil {
		t.Fail()
	}
}
