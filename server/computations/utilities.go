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
