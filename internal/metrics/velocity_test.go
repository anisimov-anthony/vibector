package metrics

import (
	"testing"
	"time"
)

func TestCalculateVelocity(t *testing.T) {
	tests := []struct {
		name          string
		loc           int64
		timeDelta     time.Duration
		want          float64
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal velocity calculation",
			loc:         100,
			timeDelta:   10 * time.Minute,
			want:        10.0,
			expectError: false,
		},
		{
			name:        "high velocity",
			loc:         500,
			timeDelta:   5 * time.Minute,
			want:        100.0,
			expectError: false,
		},
		{
			name:        "low velocity",
			loc:         10,
			timeDelta:   60 * time.Minute,
			want:        0.16666666666666666,
			expectError: false,
		},
		{
			name:        "one second time delta",
			loc:         10,
			timeDelta:   1 * time.Second,
			want:        600.0,
			expectError: false,
		},
		{
			name:          "zero time delta",
			loc:           100,
			timeDelta:     0,
			expectError:   true,
			errorContains: "invalid time delta",
		},
		{
			name:          "negative time delta",
			loc:           100,
			timeDelta:     -10 * time.Minute,
			expectError:   true,
			errorContains: "invalid time delta",
		},
		{
			name:        "zero LOC",
			loc:         0,
			timeDelta:   10 * time.Minute,
			want:        0.0,
			expectError: false,
		},
		{
			name:        "fractional minutes",
			loc:         50,
			timeDelta:   90 * time.Second,
			want:        33.333333333333336,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateVelocity(tt.loc, tt.timeDelta)
			if tt.expectError {
				if err == nil {
					t.Errorf("CalculateVelocity() expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("CalculateVelocity() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateVelocity() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("CalculateVelocity() returned nil")
				return
			}

			if got.LOCPerMinute != tt.want {
				t.Errorf("CalculateVelocity() LOCPerMinute = %v, want %v", got.LOCPerMinute, tt.want)
			}
		})
	}
}

func TestCalculateVelocityPerMinute(t *testing.T) {
	tests := []struct {
		name          string
		loc           int64
		timeDelta     time.Duration
		want          float64
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal calculation",
			loc:         200,
			timeDelta:   20 * time.Minute,
			want:        10.0,
			expectError: false,
		},
		{
			name:        "very high velocity",
			loc:         1000,
			timeDelta:   1 * time.Minute,
			want:        1000.0,
			expectError: false,
		},
		{
			name:          "zero time delta",
			loc:           100,
			timeDelta:     0,
			expectError:   true,
			errorContains: "invalid time delta",
		},
		{
			name:          "negative time delta",
			loc:           50,
			timeDelta:     -5 * time.Minute,
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:        "zero LOC returns zero velocity",
			loc:         0,
			timeDelta:   10 * time.Minute,
			want:        0.0,
			expectError: false,
		},
		{
			name:        "large numbers",
			loc:         100000,
			timeDelta:   100 * time.Minute,
			want:        1000.0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateVelocityPerMinute(tt.loc, tt.timeDelta)
			if tt.expectError {
				if err == nil {
					t.Errorf("CalculateVelocityPerMinute() expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("CalculateVelocityPerMinute() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateVelocityPerMinute() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("CalculateVelocityPerMinute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
