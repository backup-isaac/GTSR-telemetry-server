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
// Unit: m/s^2
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

// Distance is the vehicle's distance traveled, computed as cumsum(RPM_Derived_Velocity*dt)
// TODO reset when car goes offline
type Distance struct {
	cumSum     float64
	velocities []*datatypes.Datapoint
	idx        int
}

// NewDistance returns an initialized Distance
func NewDistance() *Distance {
	return &Distance{
		cumSum:     0,
		velocities: make([]*datatypes.Datapoint, 2),
	}
}

// GetMetrics returns the metrics that Distance depends upon
func (d *Distance) GetMetrics() []string {
	return []string{"RPM_Derived_Velocity", "Connection_Status"}
}

// Update signifies an update when two velocities have been stored
// so that a ∆time can be computed. A Connection_Status = 0 point
// resets the distance traveled so far
// Unit: meter
func (d *Distance) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "RPM_Derived_Velocity" {
		d.velocities[d.idx] = point
		d.idx++
	} else if point.Value == 0 {
		d.cumSum = 0
		d.idx = 0
	}
	return d.idx == 2
}

// Compute computes distance as cumsum(RPM_Derived_Velocity * dt)
func (d *Distance) Compute() *datatypes.Datapoint {
	d.cumSum += (d.velocities[1].Value + d.velocities[0].Value) * (d.velocities[1].Time.Sub(d.velocities[0].Time).Seconds()) / 2
	d.velocities[0] = d.velocities[1]
	d.idx = 1
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  d.cumSum,
		Time:   d.velocities[0].Time,
	}
}

func init() {
	Register(NewVelocity())
	Register(NewAcceleration())
	Register(NewDistance())
}
