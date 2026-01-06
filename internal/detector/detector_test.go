package detector

import (
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/git"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		thresholds  *Thresholds
		expectError bool
	}{
		{
			name: "valid thresholds",
			thresholds: &Thresholds{
				SuspiciousAdditions: 100,
				MaxAdditionsPerMin:  50.0,
			},
			expectError: false,
		},
		{
			name:        "invalid thresholds - all zeros",
			thresholds:  &Thresholds{},
			expectError: true,
		},
		{
			name: "invalid thresholds - negative value",
			thresholds: &Thresholds{
				SuspiciousAdditions: -100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.thresholds)
			if tt.expectError {
				if err == nil {
					t.Errorf("New() expected error but got none")
				}
				if got != nil {
					t.Errorf("New() expected nil detector on error, got %v", got)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if got == nil {
					t.Errorf("New() returned nil detector")
				}
			}
		})
	}
}

func TestDetector_DetectSuspicious(t *testing.T) {
	now := time.Now()

	t.Run("nil pairs returns empty", func(t *testing.T) {
		d, _ := New(&Thresholds{SuspiciousAdditions: 100})
		result := d.DetectSuspicious(nil, nil)
		if len(result) != 0 {
			t.Errorf("DetectSuspicious(nil) returned %d results, want 0", len(result))
		}
	})

	t.Run("empty pairs returns empty", func(t *testing.T) {
		d, _ := New(&Thresholds{SuspiciousAdditions: 100})
		result := d.DetectSuspicious([]*git.CommitPair{}, nil)
		if len(result) != 0 {
			t.Errorf("DetectSuspicious([]) returned %d results, want 0", len(result))
		}
	})

	t.Run("detects suspicious additions by count", func(t *testing.T) {
		d, _ := New(&Thresholds{SuspiciousAdditions: 100})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Author:    "John Doe",
					Email:     "john@example.com",
					Timestamp: now,
				},
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 200,
					Deletions: 50,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
		if len(result[0].Reasons) != 1 {
			t.Errorf("Expected 1 reason, got %d", len(result[0].Reasons))
		}
	})

	t.Run("detects suspicious deletions by count", func(t *testing.T) {
		d, _ := New(&Thresholds{SuspiciousDeletions: 100})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 50,
					Deletions: 200,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
	})

	t.Run("detects high addition velocity", func(t *testing.T) {
		d, _ := New(&Thresholds{MaxAdditionsPerMin: 50.0})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 1 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 100,
					Deletions: 10,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
		if result[0].AdditionVelocity == nil {
			t.Errorf("AdditionVelocity should not be nil")
		}
		if result[0].AdditionVelocity.LOCPerMinute <= 50.0 {
			t.Errorf("AdditionVelocity = %f, want > 50.0", result[0].AdditionVelocity.LOCPerMinute)
		}
	})

	t.Run("detects high deletion velocity", func(t *testing.T) {
		d, _ := New(&Thresholds{MaxDeletionsPerMin: 100.0})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 1 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 10,
					Deletions: 200,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
		if result[0].DeletionVelocity == nil {
			t.Errorf("DeletionVelocity should not be nil")
		}
		if result[0].DeletionVelocity.LOCPerMinute <= 100.0 {
			t.Errorf("DeletionVelocity = %f, want > 100.0", result[0].DeletionVelocity.LOCPerMinute)
		}
	})

	t.Run("detects short time delta", func(t *testing.T) {
		d, _ := New(&Thresholds{MinTimeDeltaSeconds: 60})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 30 * time.Second,
				Stats: &git.DiffStats{
					Additions: 10,
					Deletions: 5,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
	})

	t.Run("detects multiple violations", func(t *testing.T) {
		d, _ := New(&Thresholds{
			SuspiciousAdditions: 100,
			MaxAdditionsPerMin:  50.0,
			MinTimeDeltaSeconds: 60,
		})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 30 * time.Second,
				Stats: &git.DiffStats{
					Additions: 200,
					Deletions: 10,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 1 {
			t.Fatalf("DetectSuspicious() returned %d results, want 1", len(result))
		}
		if len(result[0].Reasons) != 3 {
			t.Errorf("Expected 3 reasons, got %d: %v", len(result[0].Reasons), result[0].Reasons)
		}
	})

	t.Run("ignores commits with no changes", func(t *testing.T) {
		d, _ := New(&Thresholds{SuspiciousAdditions: 10})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 0,
					Deletions: 0,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 0 {
			t.Errorf("DetectSuspicious() returned %d results, want 0 (no changes)", len(result))
		}
	})

	t.Run("does not detect when below thresholds", func(t *testing.T) {
		d, _ := New(&Thresholds{
			SuspiciousAdditions: 100,
			SuspiciousDeletions: 100,
			MaxAdditionsPerMin:  50.0,
			MaxDeletionsPerMin:  100.0,
			MinTimeDeltaSeconds: 60,
		})
		pairs := []*git.CommitPair{
			{
				Previous: &git.Commit{Hash: "abc123"},
				Current: &git.Commit{
					Hash:      "def456",
					Timestamp: now,
				},
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 50,
					Deletions: 50,
				},
			},
		}

		result := d.DetectSuspicious(pairs, nil)
		if len(result) != 0 {
			t.Errorf("DetectSuspicious() returned %d results, want 0", len(result))
		}
	})
}

func TestFormatTimeDelta(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "less than one minute",
			duration: 30 * time.Second,
			want:     "0.50 minutes",
		},
		{
			name:     "exactly one minute",
			duration: 1 * time.Minute,
			want:     "1.0 minutes",
		},
		{
			name:     "more than one minute",
			duration: 5 * time.Minute,
			want:     "5.0 minutes",
		},
		{
			name:     "fractional minutes above 1",
			duration: 5*time.Minute + 30*time.Second,
			want:     "5.5 minutes",
		},
		{
			name:     "very short duration",
			duration: 5 * time.Second,
			want:     "0.08 minutes",
		},
		{
			name:     "long duration",
			duration: 120 * time.Minute,
			want:     "120.0 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeDelta(tt.duration)
			if got != tt.want {
				t.Errorf("FormatTimeDelta() = %v, want %v", got, tt.want)
			}
		})
	}
}
