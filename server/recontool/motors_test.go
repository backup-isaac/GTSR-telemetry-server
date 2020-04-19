package recontool

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelDerivedPower(t *testing.T) {
	assert.InDelta(t, 2424.2424, ModelDerivedPower(600, 4, 0.99), fd(2424.2424))
}

func TestMotorEfficiency(t *testing.T) {
	assert.InDelta(t, 0.994232069, MotorEfficiency(120, 30), fd(0.994232069))
	assert.Equal(t, 1.0, MotorEfficiency(120, -1))
}

func TestMotorPower(t *testing.T) {
	assert.InDelta(t, 804.16051, MotorPower(20, 10, 4, 0.9, sr3), fd(804.16051))
}

func TestModeledMotorForce(t *testing.T) {
	assert.InDelta(t, 72.035364, ModeledMotorForce(5, 5, math.Pi/-6, sr3), fd(72.035364))
}

func TestMotorTorque(t *testing.T) {
	assert.InDelta(t, 8.4852, MotorTorque(50, 6, 40), fd(8.4852))
	assert.Equal(t, 5.0, MotorTorque(50, 6, 5.0))
}

func TestVelocity(t *testing.T) {
	assert.InDelta(t, 3.14159265, Velocity(100, 0.3), fd(3.14159265))
}
