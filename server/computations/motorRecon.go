package computations

import (
	"fmt"
	"server/datatypes"
	"server/recontool"

	"time"
)

// Velocity is the vehicle's velocity computed from motor RPM
type Velocity struct {
	avgRpm *datatypes.Datapoint
}

// NewVelocity returns an initialized Velocity
func NewVelocity() *Velocity {
	return &Velocity{}
}

// GetMetrics returns the Velocity's metrics
func (v *Velocity) GetMetrics() []string {
	return []string{"Average_Wavesculptor_RPM"}
}

// Update signifies an update when both a left and a right rpm have been received
func (v *Velocity) Update(point *datatypes.Datapoint) bool {
	v.avgRpm = point
	return true
}

// Compute returns the current velocity of the car in m/s
func (v *Velocity) Compute() *datatypes.Datapoint {
	avgRpm := v.avgRpm.Value
	time := v.avgRpm.Time
	v.avgRpm = nil
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Velocity",
		Value:  recontool.Velocity(avgRpm, sr3.RMot),
		Time:   time,
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
	latest := t.phaseCurrent.Time
	if t.rpm.Time.After(latest) {
		latest = t.rpm.Time
	}
	rpm := t.rpm.Value
	phaseC := t.phaseCurrent.Value
	t.rpm = nil
	t.phaseCurrent = nil
	return &datatypes.Datapoint{
		Metric: fmt.Sprintf("%s_RPM_Derived_Torque", t.motor),
		Value:  recontool.MotorTorque(rpm, phaseC, sr3.TMax),
		Time:   latest,
	}
}

// ModeledMotorForce calculates the magnitude of force that the motors exert
// to cause the car to move
type ModeledMotorForce struct {
	velocity     *datatypes.Datapoint
	acceleration *datatypes.Datapoint
	terrainAngle *datatypes.Datapoint
}

// NewModeledMotorForce returns an initialized ModeledMotorForce
func NewModeledMotorForce() *ModeledMotorForce {
	return &ModeledMotorForce{}
}

// GetMetrics returns the ModeledMotorForce's metrics
func (f *ModeledMotorForce) GetMetrics() []string {
	return []string{"RPM_Derived_Velocity", "RPM_Derived_Acceleration", "Terrain_Angle"}
}

// Update signifies an update when all required metrics have been received
func (f *ModeledMotorForce) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "RPM_Derived_Velocity":
		f.velocity = point
	case "RPM_Derived_Acceleration":
		f.acceleration = point
	case "Terrain_Angle":
		f.terrainAngle = point
	}
	return f.velocity != nil && f.acceleration != nil && f.terrainAngle != nil
}

// Compute returns the modeled motor force in Newtons
func (f *ModeledMotorForce) Compute() *datatypes.Datapoint {
	latest := f.velocity.Time
	if f.acceleration.Time.After(latest) {
		latest = f.acceleration.Time
	}
	if f.terrainAngle.Time.After(latest) {
		latest = f.terrainAngle.Time
	}
	velocity := f.velocity.Value
	accel := f.acceleration.Value
	angle := f.terrainAngle.Value
	f.velocity = nil
	f.acceleration = nil
	f.terrainAngle = nil
	return &datatypes.Datapoint{
		Metric: "Modeled_Motor_Force",
		Value:  recontool.ModeledMotorForce(velocity, accel, angle, sr3),
		Time:   latest,
	}
}

// ModeledMotorTorque calculates motor torque from modeled force
type ModeledMotorTorque struct {
	force *datatypes.Datapoint
}

// NewModeledMotorTorque returns an initialized ModelMotorTorque
func NewModeledMotorTorque() *ModeledMotorTorque {
	return &ModeledMotorTorque{}
}

// GetMetrics returns the ModeledMotorTorque's metrics
func (t *ModeledMotorTorque) GetMetrics() []string {
	return []string{"Modeled_Motor_Force"}
}

// Update signifies an update when a new force point is received
func (t *ModeledMotorTorque) Update(point *datatypes.Datapoint) bool {
	t.force = point
	return true
}

// Compute computes modeled motor torque in Nm
func (t *ModeledMotorTorque) Compute() *datatypes.Datapoint {
	time := t.force.Time
	force := t.force.Value
	t.force = nil
	return &datatypes.Datapoint{
		Metric: "Modeled_Motor_Torque",
		Value:  force * sr3.RMot,
		Time:   time,
	}
}

// MotorEfficiency calculates motor efficiency
type MotorEfficiency struct {
	busVoltage *datatypes.Datapoint
	torque     *datatypes.Datapoint
}

// NewMotorEfficiency returns an initialized MotorEfficiency
func NewMotorEfficiency() *MotorEfficiency {
	return &MotorEfficiency{}
}

// GetMetrics returns the MotorEfficiency's metrics
func (e *MotorEfficiency) GetMetrics() []string {
	return []string{"Average_Bus_Voltage", "RPM_Derived_Torque"}
}

// Update signifies an update when both required metrics have been received
func (e *MotorEfficiency) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Average_Bus_Voltage" {
		e.busVoltage = point
	} else if point.Metric == "RPM_Derived_Torque" {
		e.torque = point
	}
	return e.busVoltage != nil && e.torque != nil
}

// Compute computes motor efficiency
func (e *MotorEfficiency) Compute() *datatypes.Datapoint {
	latest := e.busVoltage.Time
	if e.torque.Time.After(latest) {
		latest = e.torque.Time
	}
	vBus := e.busVoltage.Value
	torque := e.torque.Value
	e.busVoltage = nil
	e.torque = nil
	return &datatypes.Datapoint{
		Metric: "Motor_Efficiency",
		Value:  recontool.MotorEfficiency(vBus, torque),
		Time:   latest,
	}
}

// EmpiricalMotorPower computes motor power from torque, velocity,
// and drivetrain characteristics
type EmpiricalMotorPower struct {
	torque               *datatypes.Datapoint
	velocity             *datatypes.Datapoint
	phaseCCurrent        *datatypes.Datapoint
	drivetrainEfficiency *datatypes.Datapoint
}

// NewEmpiricalMotorPower returns an initialized EmpiricalMotorPower
func NewEmpiricalMotorPower() *EmpiricalMotorPower {
	return &EmpiricalMotorPower{}
}

// GetMetrics returns the EmpiricalMotorPower's metrics
func (p *EmpiricalMotorPower) GetMetrics() []string {
	return []string{"RPM_Derived_Torque", "RPM_Derived_Velocity", "Phase_C_Current", "Drivetrain_Efficiency"}
}

// Update signifies an update when all required metrics have been received
func (p *EmpiricalMotorPower) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "RPM_Derived_Torque":
		p.torque = point
	case "RPM_Derived_Velocity":
		p.velocity = point
	case "Phase_C_Current":
		p.phaseCCurrent = point
	case "Drivetrain_Efficiency":
		p.drivetrainEfficiency = point
	}
	return p.torque != nil && p.velocity != nil && p.phaseCCurrent != nil && p.drivetrainEfficiency != nil
}

// Compute computes empirical motor power in Watts
func (p *EmpiricalMotorPower) Compute() *datatypes.Datapoint {
	latest := p.torque.Time
	if p.velocity.Time.After(latest) {
		latest = p.velocity.Time
	}
	if p.phaseCCurrent.Time.After(latest) {
		latest = p.phaseCCurrent.Time
	}
	if p.drivetrainEfficiency.Time.After(latest) {
		latest = p.drivetrainEfficiency.Time
	}
	torque := p.torque.Value
	velocity := p.velocity.Value
	iPhaseC := p.phaseCCurrent.Value
	effDt := p.drivetrainEfficiency.Value
	p.torque = nil
	p.velocity = nil
	p.phaseCCurrent = nil
	p.drivetrainEfficiency = nil
	return &datatypes.Datapoint{
		Metric: "RPM_Derived_Motor_Power",
		Value:  recontool.MotorPower(torque, velocity, iPhaseC, effDt, sr3),
		Time:   latest,
	}
}

// ModeledMotorPower computes motor power from modeled force, velocity,
// and drivetrain efficiency
type ModeledMotorPower struct {
	force                *datatypes.Datapoint
	velocity             *datatypes.Datapoint
	drivetrainEfficiency *datatypes.Datapoint
}

// NewModeledMotorPower returns an initialized EmpiricalMotorPower
func NewModeledMotorPower() *ModeledMotorPower {
	return &ModeledMotorPower{}
}

// GetMetrics returns the ModeledMotorPower's metrics
func (p *ModeledMotorPower) GetMetrics() []string {
	return []string{"Modeled_Motor_Force", "RPM_Derived_Velocity", "Drivetrain_Efficiency"}
}

// Update signifies an update when all required metrics have been received
func (p *ModeledMotorPower) Update(point *datatypes.Datapoint) bool {
	switch point.Metric {
	case "Modeled_Motor_Force":
		p.force = point
	case "RPM_Derived_Velocity":
		p.velocity = point
	case "Drivetrain_Efficiency":
		p.drivetrainEfficiency = point
	}
	return p.force != nil && p.velocity != nil && p.drivetrainEfficiency != nil
}

// Compute computes modeled motor power in Watts
func (p *ModeledMotorPower) Compute() *datatypes.Datapoint {
	latest := p.force.Time
	if p.velocity.Time.After(latest) {
		latest = p.velocity.Time
	}
	if p.drivetrainEfficiency.Time.After(latest) {
		latest = p.drivetrainEfficiency.Time
	}
	force := p.force.Value
	velocity := p.velocity.Value
	effDt := p.drivetrainEfficiency.Value
	p.force = nil
	p.velocity = nil
	p.drivetrainEfficiency = nil
	return &datatypes.Datapoint{
		Metric: "Modeled_Motor_Power",
		Value:  recontool.ModelDerivedPower(force, velocity, effDt),
		Time:   latest,
	}
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
