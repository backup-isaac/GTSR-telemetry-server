package recontool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveTerrainAngle(t *testing.T) {
	assert.InDelta(t, -0.017214401, DeriveTerrainAngle(10, 1, 0.2, sr3), fd(0.011813825))
}
