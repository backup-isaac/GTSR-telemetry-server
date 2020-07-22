package computations

import (
	"fmt"
	"server/datatypes"
)

// LeftRightSum sums the left and right versions of a metric
type LeftRightSum struct {
	standardComputation
	baseMetric string
}

// NewLeftRightSum returns an initialized LeftRightSum that will
// base itself off of the specified base metric
func NewLeftRightSum(baseMetric string) *LeftRightSum {
	return &LeftRightSum{
		standardComputation: standardComputation{
			values: make(map[string]float64),
			fields: []string{fmt.Sprintf("Left_%s", baseMetric), fmt.Sprintf("Right_%s", baseMetric)},
		},
		baseMetric: baseMetric,
	}
}

// Compute adds Left_[base metric] + Right_[base metric]
func (s *LeftRightSum) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: s.baseMetric,
		Value:  s.values[fmt.Sprintf("Left_%s", s.baseMetric)] + s.values[fmt.Sprintf("Right_%s", s.baseMetric)],
		Time:   s.timestamp,
	}
	s.values = make(map[string]float64)
	return datapoint
}

// LeftRightAverage averages the left and right versions of a metric
type LeftRightAverage struct {
	standardComputation
	// left       *datatypes.Datapoint
	// right      *datatypes.Datapoint
	baseMetric string
}

// NewLeftRightAverage returns an initialized LeftRightAverage that will
// base itself off of the specified metric
func NewLeftRightAverage(baseMetric string) *LeftRightAverage {
	return &LeftRightAverage{
		standardComputation: standardComputation{
			values: make(map[string]float64),
			fields: []string{fmt.Sprintf("Left_%s", baseMetric), fmt.Sprintf("Right_%s", baseMetric)},
		},
		baseMetric: baseMetric,
	}
}

// Compute averages Left_[base metric] with Right_[base metric]
func (a *LeftRightAverage) Compute() *datatypes.Datapoint {
	datapoint := &datatypes.Datapoint{
		Metric: fmt.Sprintf("Average_%s", a.baseMetric),
		Value:  (a.values[fmt.Sprintf("Left_%s", a.baseMetric)] + a.values[fmt.Sprintf("Right_%s", a.baseMetric)]) / 2,
		Time:   a.timestamp,
	}
	a.values = make(map[string]float64)
	return datapoint
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
	c.cumSum += c.currents[0].Value * (c.currents[1].Time.Sub(c.currents[0].Time).Seconds())
	t := c.currents[0].Time
	c.currents[0] = c.currents[1]
	c.idx = 1
	return &datatypes.Datapoint{
		Metric: fmt.Sprintf("%s_Charge_Consumed", c.currentName),
		Value:  c.cumSum,
		Time:   t,
	}
}
