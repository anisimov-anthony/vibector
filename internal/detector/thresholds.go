package detector

import "fmt"

type Thresholds struct {
	SuspiciousAdditions int64
	SuspiciousDeletions int64

	MaxAdditionsPerMin float64
	MaxDeletionsPerMin float64

	MinTimeDeltaSeconds int64
}

func (t *Thresholds) Validate() error {
	if t.SuspiciousAdditions < 0 {
		return fmt.Errorf("SuspiciousAdditions cannot be negative")
	}

	if t.SuspiciousDeletions < 0 {
		return fmt.Errorf("SuspiciousDeletions cannot be negative")
	}

	if t.MaxAdditionsPerMin < 0 {
		return fmt.Errorf("MaxAdditionsPerMin cannot be negative")
	}

	if t.MaxDeletionsPerMin < 0 {
		return fmt.Errorf("MaxDeletionsPerMin cannot be negative")
	}

	if t.MinTimeDeltaSeconds < 0 {
		return fmt.Errorf("MinTimeDeltaSeconds cannot be negative")
	}

	if t.SuspiciousAdditions == 0 &&
		t.SuspiciousDeletions == 0 &&
		t.MaxAdditionsPerMin == 0 &&
		t.MaxDeletionsPerMin == 0 &&
		t.MinTimeDeltaSeconds == 0 {
		return fmt.Errorf("at least one threshold must be configured")
	}

	return nil
}

func (t *Thresholds) IsZero() bool {
	return t.SuspiciousAdditions == 0 &&
		t.SuspiciousDeletions == 0 &&
		t.MaxAdditionsPerMin == 0 &&
		t.MaxDeletionsPerMin == 0 &&
		t.MinTimeDeltaSeconds == 0
}
