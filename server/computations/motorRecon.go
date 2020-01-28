package computations

import (
	"fmt"
	"server/datatypes"
	"server/recontool"

	"time"
)

var sr3 *recontool.Vehicle

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
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(avgRpm, sr3.RMot),
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

// EmpiricalTorque computes a motor torque empirically from phase current and RPM
type EmpiricalTorque struct {
	phaseCurrent *datatypes.Datapoint
	rpm          *datatypes.Datapoint
	motor        string
}

// NewEmpiricalTorque returns an initialized EmpiricalTorque that will
// base itself off of the specified motor
func NewEmpiricalTorque(motor string) *EmpiricalTorque {
	return &EmpiricalTorque{
		motor: motor,
	}
}

// GetMetrics returns the EmpiricalTorque's metrics
func (t *EmpiricalTorque) GetMetrics() []string {
	return []string{fmt.Sprintf("%s_Phase_C_Current", t.motor), fmt.Sprintf("%s_Wavesculptor_RPM", t.motor)}
}

// Update signifies an update when both a phase current and an RPM have been received
func (t *EmpiricalTorque) Update(point *datatypes.Datapoint) bool {
	if point.Metric == fmt.Sprintf("%s_Phase_C_Current", t.motor) {
		t.phaseCurrent = point
	} else if point.Metric == fmt.Sprintf("%s_Wavesculptor_RPM", t.motor) {
		t.rpm = point
	}
	return t.phaseCurrent != nil && t.rpm != nil
}

// Compute returns the motor's torque in Nm
func (t *EmpiricalTorque) Compute() *datatypes.Datapoint {
	avgTime := t.phaseCurrent.Time.Add(t.phaseCurrent.Time.Sub(t.rpm.Time) / 2)
	rpm := t.rpm.Value
	phaseC := t.phaseCurrent.Value
	t.rpm = nil
	t.phaseCurrent = nil
	return &datatypes.Datapoint{
		Metric: fmt.Sprintf("%s_RPM_Derived_Torque", t.motor),
		Value:  recontool.MotorTorque(rpm, phaseC, sr3.TMax),
		Time:   avgTime,
	}
}

// LeftRightSum sums the left and right versions of a metric
type LeftRightSum struct {
	left       *datatypes.Datapoint
	right      *datatypes.Datapoint
	baseMetric string
}

// NewLeftRightSum returns an initialized LeftRightSum that will
// base itself off of the specified base metric
func NewLeftRightSum(baseMetric string) *LeftRightSum {
	return &LeftRightSum{
		baseMetric: baseMetric,
	}
}

// GetMetrics returns the LeftRightSum's metrics
func (s *LeftRightSum) GetMetrics() []string {
	return []string{fmt.Sprintf("Left_%s", s.baseMetric), fmt.Sprintf("Right_%s", s.baseMetric)}
}

// Update signifies an update when both the left and right versions of the metric have been received
func (s *LeftRightSum) Update(point *datatypes.Datapoint) bool {
	if point.Metric == fmt.Sprintf("Left_%s", s.baseMetric) {
		s.left = point
	} else if point.Metric == fmt.Sprintf("Right_%s", s.baseMetric) {
		s.right = point
	}
	return s.left != nil && s.right != nil
}

// Compute adds Left_[base metric] + Right_[base metric]
func (s *LeftRightSum) Compute() *datatypes.Datapoint {
	avgTime := s.left.Time.Add(s.left.Time.Sub(s.right.Time) / 2)
	left := s.left.Value
	right := s.right.Value
	s.left = nil
	s.right = nil
	return &datatypes.Datapoint{
		Metric: s.baseMetric,
		Value:  left + right,
		Time:   avgTime,
	}
}

// TerrainAngle computes the angle of the terrain that the vehicle is driving on
// by deriving the amount of gravitational force that is accelerating the vehicle
type TerrainAngle struct {
	torque       *datatypes.Datapoint
	velocity     *datatypes.Datapoint
	acceleration *datatypes.Datapoint
}

// NewTerrainAngle returns an initialized TerrainAngle
func NewTerrainAngle() *TerrainAngle {
	return &TerrainAngle{}
}

// GetMetrics returns the TerrainAngle's metrics
func (t *TerrainAngle) GetMetrics() []string {
	return []string{"RPM_Derived_Torque", "RPM_Derived_Velocity", "RPM_Derived_Acceleration"}
}

// Update signifies an update when all three required metrics have been received
func (t *TerrainAngle) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "RPM_Derived_Torque":
		t.torque = point
	case "RPM_Derived_Velocity":
		t.velocity = point
	case "RPM_Derived_Acceleration":
		t.acceleration = point
	}
	return t.torque != nil && t.velocity != nil && t.acceleration != nil
}

// Compute returns the terrain angle in radians
func (t *TerrainAngle) Compute() *datatypes.Datapoint {
	latest := t.torque.Time
	if t.velocity.Time.After(latest) {
		latest = t.velocity.Time
	}
	if t.acceleration.Time.After(latest) {
		latest = t.acceleration.Time
	}
	torque := t.torque.Value
	velocity := t.velocity.Value
	accel := t.acceleration.Value
	t.torque = nil
	t.velocity = nil
	t.acceleration = nil
	return &datatypes.Datapoint{
		Metric: "Terrain_Angle",
		Value:  recontool.DeriveTerrainAngle(torque, velocity, accel, sr3),
		Time:   latest,
	}
}

func init() {
	sr3 = &recontool.Vehicle{
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
	Register(NewVelocity())
	Register(NewAcceleration())
	Register(NewDistance())
	Register(NewEmpiricalTorque("Left"))
	Register(NewEmpiricalTorque("Right"))
	Register(NewLeftRightSum("RPM_Derived_Torque"))
	Register(NewTerrainAngle())
}
