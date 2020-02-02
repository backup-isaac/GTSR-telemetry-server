package computations

import (
	"fmt"
	"server/datatypes"
)

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
	latest := s.left.Time
	if s.right.Time.After(latest) {
		latest = s.right.Time
	}
	left := s.left.Value
	right := s.right.Value
	s.left = nil
	s.right = nil
	return &datatypes.Datapoint{
		Metric: s.baseMetric,
		Value:  left + right,
		Time:   latest,
	}
}

// LeftRightAverage averages the left and right versions of a metric
type LeftRightAverage struct {
	left       *datatypes.Datapoint
	right      *datatypes.Datapoint
	baseMetric string
}

// NewLeftRightAverage returns an initialized LeftRightAverage that will
// base itself off of the specified metric
func NewLeftRightAverage(baseMetric string) *LeftRightAverage {
	return &LeftRightAverage{
		baseMetric: baseMetric,
	}
}

// GetMetrics returns the LeftRightAverage's metrics
func (a *LeftRightAverage) GetMetrics() []string {
	return []string{fmt.Sprintf("Left_%s", a.baseMetric), fmt.Sprintf("Right_%s", a.baseMetric)}
}

// Update signifies an update when both the left and right versions of the metric have been received
func (a *LeftRightAverage) Update(point *datatypes.Datapoint) bool {
	if point.Metric == fmt.Sprintf("Left_%s", a.baseMetric) {
		a.left = point
	} else if point.Metric == fmt.Sprintf("Right_%s", a.baseMetric) {
		a.right = point
	}
	return a.left != nil && a.right != nil
}

// Compute averages Left_[base metric] with Right_[base metric]
func (a *LeftRightAverage) Compute() *datatypes.Datapoint {
	latest := a.left.Time
	if a.right.Time.After(latest) {
		latest = a.right.Time
	}
	left := a.left.Value
	right := a.right.Value
	a.left = nil
	a.right = nil
	return &datatypes.Datapoint{
		Metric: fmt.Sprintf("Average_%s", a.baseMetric),
		Value:  (left + right) / 2,
		Time:   latest,
	}
}

// ChargeIntegral computes cumsum(current*dt)
// Resets when car goes offline
type ChargeIntegral struct {
	cumSum      float64
	currents    []*datatypes.Datapoint
	idx         int
	currentName string
}

// NewChargeIntegral returns an initialized ChargeIntegral
func NewChargeIntegral(currentName string) *ChargeIntegral {
	return &ChargeIntegral{
		cumSum:      0,
		currents:    make([]*datatypes.Datapoint, 2),
		currentName: currentName,
	}
}

// GetMetrics returns the ChargeIntegral's metrics
func (c *ChargeIntegral) GetMetrics() []string {
	return []string{fmt.Sprintf("%s_Current", c.currentName), "Connection_Status"}
}

// Update signifies an update when two currents have been stored
// so that a ∆time can be computed. A Connection_Status = 0 point
// resets the charge consumed so far
// Unit: coulomb
func (c *ChargeIntegral) Update(point *datatypes.Datapoint) bool {
	if point.Metric == "Connection_Status" {
		if point.Value == 0 {
			c.cumSum = 0
			c.idx = 0
		}
	} else {
		c.currents[c.idx] = point
		c.idx++
	}
	return c.idx == 2
}

// Compute computes charge as cumsum(current * dt)
func (c *ChargeIntegral) Compute() *datatypes.Datapoint {
	c.cumSum += (c.currents[1].Value + c.currents[0].Value) * (c.currents[1].Time.Sub(c.currents[0].Time).Seconds())
	c.currents[0] = c.currents[1]
	c.idx = 1
	return &datatypes.Datapoint{
		Metric: fmt.Sprintf("%s_Charge_Consumed", c.currentName),
		Value:  c.cumSum,
		Time:   c.currents[0].Time,
	}
}
