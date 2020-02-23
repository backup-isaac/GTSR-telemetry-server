package recontool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBusPower(t *testing.T) {
	assert.InDelta(t, 100, BusPower(100, 100, 2, -1), fd(100))
}

func TestMotorControllerEfficiency(t *testing.T) {
	assert.InDelta(t, 0.9916353, MotorControllerEfficiency(2, 100, 10), fd(0.9916353))
	assert.Equal(t, 1.0, MotorControllerEfficiency(2, 100, -1))
}

func TestDrivetrainEfficiency(t *testing.T) {
	assert.InDelta(t, 0.9410934, DrivetrainEfficiency(0.99, 0.98, 0.97), fd(0.9410934))
}
