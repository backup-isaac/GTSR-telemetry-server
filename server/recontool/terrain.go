package recontool

import (
	"math"

	"gonum.org/v1/gonum/unit/constant"
)

// DeriveTerrainAngle computes the terrain angle based on the proportion of empirically-derived motor force that's opposing gravity
func DeriveTerrainAngle(tMot, v, a float64, vehicle *Vehicle) float64 {
	// should take into account that as the varies it affects friction too
	// jackson didn't, hence vehicle.RollingFrictionalForce(v, 0) for now
	// ehh, cos(x) ≈ 1 for small x
	fgsintheta := tMot/vehicle.RMot - vehicle.M*a - vehicle.DragForce(v) - vehicle.RollingFrictionalForce(v, 0)
	return math.Asin(fgsintheta / (vehicle.M * float64(constant.StandardGravity)))
}

// DeriveTerrainAngleSeries computes terrain angle for a series of points
func DeriveTerrainAngleSeries(tMot, dxdt, dvdt []float64, vehicle *Vehicle) []float64 {
	thetaSeries := make([]float64, len(tMot))
	for i, t := range tMot {
		thetaSeries[i] = DeriveTerrainAngle(t, dxdt[i], dvdt[i], vehicle)
	}
	return thetaSeries
}
