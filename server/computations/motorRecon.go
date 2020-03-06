package computations

import (
	"fmt"
	"server/datatypes"
	"server/recontool"

	"time"
)

// Velocity is the vehicle's velocity computed from motor RPM
type Velocity struct {
	standardComputation
}

// NewVelocity returns an initialized Velocity
func NewVelocity() *Velocity {
	return &Velocity{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Average_Wavesculptor_RPM"},
		},
	}
}

// Compute returns the current velocity of the car in m/s
func (v *Velocity) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(v.values["Average_Wavesculptor_RPM"], sr3.RMot),
		Time:   v.timestamp,
	}
	v.values = make(map[string]float64)
	return datapoint
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
	return []string{"RPM_Derived_Velocity", "Connection_Status"}
}

// Update signifies an update when there are three velocity points in the queue
// it's time to calculate a_n when we have v_{n-1}, v_n, v_{n+1}
func (a *Acceleration) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Connection_Status" {
		if point.Value == 0 {
			a.idx = 0
			a.size = 0
		}
		return false
	}
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
// Resets when car goes offline
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
	d.cumSum += d.velocities[0].Value * (d.velocities[1].Time.Sub(d.velocities[0].Time).Seconds())
	t := d.velocities[0].Time
	d.velocities[0] = d.velocities[1]
	d.idx = 1
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Distance",
		Value:  d.cumSum,
		Time:   t,
	}
}

// EmpiricalTorque computes a motor torque empirically from phase current and RPM
type EmpiricalTorque struct {
	standardComputation
	motor string
}

// NewEmpiricalTorque returns an initialized EmpiricalTorque that will
// base itself off of the specified motor
func NewEmpiricalTorque(motor string) *EmpiricalTorque {
	return &EmpiricalTorque{
		standardComputation: standardComputation{
			values: make(map[string]float64),
			fields: []string{fmt.Sprintf("%s_Phase_C_Current", motor), fmt.Sprintf("%s_Wavesculptor_RPM", motor)},
		},
		motor: motor,
	}
}

// Compute returns the motor's torque in Nm
func (t *EmpiricalTorque) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: fmt.Sprintf("%s_RPM_Derived_Torque", t.motor),
		Value: recontool.MotorTorque(
			t.values[fmt.Sprintf("%s_Wavesculptor_RPM", t.motor)],
			t.values[fmt.Sprintf("%s_Phase_C_Current", t.motor)],
			sr3.TMax,
		),
		Time: t.timestamp,
	}
	t.values = make(map[string]float64)
	return datapoint
}

// ModeledMotorForce calculates the magnitude of force that the motors exert
// to cause the car to move
type ModeledMotorForce struct {
	standardComputation
}

// NewModeledMotorForce returns an initialized ModeledMotorForce
func NewModeledMotorForce() *ModeledMotorForce {
	return &ModeledMotorForce{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"RPM_Derived_Velocity", "RPM_Derived_Acceleration", "Terrain_Angle"},
		},
	}
}

// Compute returns the modeled motor force in Newtons
func (f *ModeledMotorForce) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Modeled_Motor_Force",
		Value: recontool.ModeledMotorForce(
			f.values["RPM_Derived_Velocity"],
			f.values["RPM_Derived_Acceleration"],
			f.values["Terrain_Angle"],
			sr3,
		),
		Time: f.timestamp,
	}
	f.values = make(map[string]float64)
	return datapoint
}

// ModeledMotorTorque calculates motor torque from modeled force
type ModeledMotorTorque struct {
	standardComputation
}

// NewModeledMotorTorque returns an initialized ModelMotorTorque
func NewModeledMotorTorque() *ModeledMotorTorque {
	return &ModeledMotorTorque{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Modeled_Motor_Force"},
		},
	}
}

// Compute computes modeled motor torque in Nm
func (t *ModeledMotorTorque) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Modeled_Motor_Torque",
		Value:  t.values["Modeled_Motor_Force"] * sr3.RMot,
		Time:   t.timestamp,
	}
	t.values = make(map[string]float64)
	return datapoint
}

// MotorEfficiency calculates motor efficiency
type MotorEfficiency struct {
	standardComputation
}

// NewMotorEfficiency returns an initialized MotorEfficiency
func NewMotorEfficiency() *MotorEfficiency {
	return &MotorEfficiency{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Average_Bus_Voltage", "RPM_Derived_Torque"},
		},
	}
}

// Compute computes motor efficiency
func (e *MotorEfficiency) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Motor_Efficiency",
		Value:  recontool.MotorEfficiency(e.values["Average_Bus_Voltage"], e.values["RPM_Derived_Torque"]),
		Time:   e.timestamp,
	}
	e.values = make(map[string]float64)
	return datapoint
}

// EmpiricalMotorPower computes motor power from torque, velocity,
// and drivetrain characteristics
type EmpiricalMotorPower struct {
	standardComputation
}

// NewEmpiricalMotorPower returns an initialized EmpiricalMotorPower
func NewEmpiricalMotorPower() *EmpiricalMotorPower {
	return &EmpiricalMotorPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"RPM_Derived_Torque", "RPM_Derived_Velocity", "Phase_C_Current", "Drivetrain_Efficiency"},
		},
	}
}

// Update signifies an update when all required metrics have been received
func (p *EmpiricalMotorPower) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Drivetrain_Efficiency" && point.Value == 0 {
		p.values = make(map[string]float64)
		return false
	}
	p.values[point.Metric] = point.Value
	if point.Time.After(p.timestamp) {
		p.timestamp = point.Time
	}
	return len(p.values) >= len(p.fields)
}

// Compute computes empirical motor power in Watts
func (p *EmpiricalMotorPower) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "RPM_Derived_Motor_Power",
		Value: recontool.MotorPower(
			p.values["RPM_Derived_Torque"],
			p.values["RPM_Derived_Velocity"],
			p.values["Phase_C_Current"],
			p.values["Drivetrain_Efficiency"],
			sr3,
		),
		Time: p.timestamp,
	}
	p.values = make(map[string]float64)
	return datapoint
}

// ModeledMotorPower computes motor power from modeled force, velocity,
// and drivetrain efficiency
type ModeledMotorPower struct {
	standardComputation
}

// NewModeledMotorPower returns an initialized EmpiricalMotorPower
func NewModeledMotorPower() *ModeledMotorPower {
	return &ModeledMotorPower{
		standardComputation{
			values: make(map[string]float64),
			fields: []string{"Modeled_Motor_Force", "RPM_Derived_Velocity", "Drivetrain_Efficiency"},
		},
	}
}

// Update signifies an update when all required metrics have been received
func (p *ModeledMotorPower) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Drivetrain_Efficiency" && point.Value == 0 {
		p.values = make(map[string]float64)
		return false
	}
	p.values[point.Metric] = point.Value
	if point.Time.After(p.timestamp) {
		p.timestamp = point.Time
	}
	return len(p.values) >= len(p.fields)
}

// Compute computes modeled motor power in Watts
func (p *ModeledMotorPower) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: "Modeled_Motor_Power",
		Value: recontool.ModelDerivedPower(
			p.values["Modeled_Motor_Force"],
			p.values["RPM_Derived_Velocity"],
			p.values["Drivetrain_Efficiency"],
		),
		Time: p.timestamp,
	}
	p.values = make(map[string]float64)
	return datapoint
}

func init() {
	Register(NewLeftRightAverage("Wavesculptor_RPM"))
	Register(NewVelocity())
	Register(NewAcceleration())
	Register(NewDistance())
	Register(NewEmpiricalTorque("Left"))
	Register(NewEmpiricalTorque("Right"))
	Register(NewLeftRightSum("RPM_Derived_Torque"))
	Register(NewLeftRightSum("Phase_C_Current"))
	Register(NewModeledMotorForce())
	Register(NewModeledMotorTorque())
	Register(NewMotorEfficiency())
	Register(NewEmpiricalMotorPower())
	Register(NewModeledMotorPower())
}
