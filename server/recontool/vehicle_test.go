package recontool

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDragForce(t *testing.T) {
	assert.InDelta(t, 2.45, sr3.DragForce(5), fd(2.45))
	assert.InDelta(t, 2.45, sr3.DragForce(-5), fd(2.45))
}

func TestRollingFrictionalForce(t *testing.T) {
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(5, math.Pi/6), fd(34.50452))
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(-5, math.Pi/6), fd(34.50452))
	assert.InDelta(t, 34.50452, sr3.RollingFrictionalForce(5, math.Pi/-6), fd(34.50452))
}
