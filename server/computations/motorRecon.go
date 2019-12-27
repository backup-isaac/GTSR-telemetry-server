package computations

import (
	"server/datatypes"
	"server/recontool"

	"time"
)

// Velocity is the vehicle's velocity computed from motor RPM
type Velocity struct {
	leftRpm  *datatypes.Datapoint
	rightRpm *datatypes.Datapoint
}

// NewVelocity returns an initialized Velocity
func NewVelocity() *Velocity {
	return &Velocity{
		leftRpm:  nil,
		rightRpm: nil,
	}
}

// GetMetrics returns the Velocity's metrics
func (v *Velocity) GetMetrics() []string {
	return []string{"Left_Wavesculptor_RPM", "Right_Wavesculptor_RPM"}
}

// Update signifies an update when both a left and a right rpm have been received
func (v *Velocity) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Left_Wavesculptor_RPM" {
		v.leftRpm = point
	} else if point.Metric == "Right_Wavesculptor_RPM" {
		v.rightRpm = point
	}
	return v.leftRpm != nil && v.rightRpm != nil
}

// Compute returns the current velocity of the car in m/s
func (v *Velocity) Compute() *datatypes.Datapoint {
	avgRpm := (v.leftRpm.Value + v.rightRpm.Value) / 2
	avgTime := v.leftRpm.Time.Add(v.rightRpm.Time.Sub(v.leftRpm.Time) / 2)
	v.leftRpm = nil
	v.rightRpm = nil
	motorRadius := 0.278 // ideally we can manage this better
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(avgRpm, motorRadius),
		Time:   avgTime,
	}
}

// Acceleration is the vehicle's acceleration computed from ∆RPM_Derived_Velocity/∆t
type Acceleration struct {
	velocities []float64
	times      []time.Time
	idx        int
	size       int
}

// NewAcceleration returns an initialized Acceleration
func NewAcceleration() *Acceleration {
	return &Acceleration{
		velocities: make([]float64, 3),
		times:      make([]time.Time, 3),
		idx:        0,
		size:       0,
	}
}

// GetMetrics returns the Acceleration's metrics
func (a *Acceleration) GetMetrics() []string {
	return []string{"RPM_Derived_Velocity"}
}

// Update signifies an update when there are three velocity points in the queue
// it's time to calculate a_n when we have v_{n-1}, v_n, v_{n+1}
func (a *Acceleration) Update(point *datatypes.Datapoint) bool {
	a.velocities[a.idx] = point.Value
	a.times[a.idx] = point.Time
	a.idx = (a.idx + 1) % 3
	a.size++
	return a.size == 3
}

// Compute computes the current acceleration as
// a_n = (v_{n+1}-v_{n-1})/(t_{n+1}-t_{n-1})
func (a *Acceleration) Compute() *datatypes.Datapoint {
	a.size--
	beforeIndex := a.idx
	nowIndex := (a.idx + 1) % 3
	afterIndex := (a.idx + 2) % 3
	dvdt := (a.velocities[afterIndex] - a.velocities[beforeIndex]) / a.times[afterIndex].Sub(a.times[beforeIndex]).Seconds()
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Acceleration",
		Value:  dvdt,
		Time:   a.times[nowIndex],
	}
}

func init() {
	Register(NewVelocity())
	Register(NewAcceleration())
}
