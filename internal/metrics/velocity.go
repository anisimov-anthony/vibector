package metrics

import (
	"fmt"
	"time"
)

type VelocityMetrics struct {
	LOCPerMinute float64
}

func CalculateVelocity(loc int64, timeDelta time.Duration) (*VelocityMetrics, error) {
	if timeDelta <= 0 {
		return nil, fmt.Errorf("invalid time delta: %v (must be positive)", timeDelta)
	}

	minutes := timeDelta.Minutes()
	locFloat := float64(loc)

	return &VelocityMetrics{
		LOCPerMinute: locFloat / minutes,
	}, nil
}

func CalculateVelocityPerMinute(loc int64, timeDelta time.Duration) (float64, error) {
	if timeDelta <= 0 {
		return 0, fmt.Errorf("invalid time delta: %v (must be positive)", timeDelta)
	}
	return float64(loc) / timeDelta.Minutes(), nil
}
