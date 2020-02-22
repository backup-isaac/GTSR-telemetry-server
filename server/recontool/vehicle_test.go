package recontool

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sr3 = &Vehicle{
	RMot:  0.278,
	M:     362.874,
	Crr1:  0.006,
	Crr2:  0.0009,
	CDa:   0.16,
	TMax:  80,
	QMax:  36,
	RLine: 0.1,
	VcMax: 4.2,
	VcMin: 2.5,
	VSer:  35,
}

func TestDragForce(t *testing.T) {
	assert.InDelta(t, 2.45, sr3.DragForce(5), fd(2.45))
	assert.InDelta(t, 2.45, sr3.DragForce(-5), fd(2.45))
}

func TestRollingFrictionalForce(t *testing.T) {
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(5, math.Pi/6), fd(34.50452))
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(-5, math.Pi/6), fd(34.50452))
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(5, math.Pi/-6), fd(34.50452))
}

func fd(f float64) float64 {
	return f * 1e-6
}
