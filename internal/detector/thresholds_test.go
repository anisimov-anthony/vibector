package detector

import (
	"testing"
)

func TestThresholds_Validate(t *testing.T) {
	tests := []struct {
		name          string
		thresholds    Thresholds
		expectError   bool
		errorContains string
	}{
		{
			name: "valid thresholds all set",
			thresholds: Thresholds{
				SuspiciousAdditions: 100,
				SuspiciousDeletions: 200,
				MaxAdditionsPerMin:  50.0,
				MaxDeletionsPerMin:  100.0,
				MinTimeDeltaSeconds: 60,
			},
			expectError: false,
		},
		{
			name: "valid thresholds only additions",
			thresholds: Thresholds{
				SuspiciousAdditions: 100,
			},
			expectError: false,
		},
		{
			name: "valid thresholds only deletions",
			thresholds: Thresholds{
				SuspiciousDeletions: 200,
			},
			expectError: false,
		},
		{
			name: "valid thresholds only velocity",
			thresholds: Thresholds{
				MaxAdditionsPerMin: 50.0,
			},
			expectError: false,
		},
		{
			name: "valid thresholds only time delta",
			thresholds: Thresholds{
				MinTimeDeltaSeconds: 30,
			},
			expectError: false,
		},
		{
			name:          "all zeros invalid",
			thresholds:    Thresholds{},
			expectError:   true,
			errorContains: "at least one threshold must be configured",
		},
		{
			name: "negative additions",
			thresholds: Thresholds{
				SuspiciousAdditions: -100,
			},
			expectError:   true,
			errorContains: "SuspiciousAdditions cannot be negative",
		},
		{
			name: "negative deletions",
			thresholds: Thresholds{
				SuspiciousDeletions: -200,
			},
			expectError:   true,
			errorContains: "SuspiciousDeletions cannot be negative",
		},
		{
			name: "negative additions per min",
			thresholds: Thresholds{
				MaxAdditionsPerMin: -50.0,
			},
			expectError:   true,
			errorContains: "MaxAdditionsPerMin cannot be negative",
		},
		{
			name: "negative deletions per min",
			thresholds: Thresholds{
				MaxDeletionsPerMin: -100.0,
			},
			expectError:   true,
			errorContains: "MaxDeletionsPerMin cannot be negative",
		},
		{
			name: "negative time delta",
			thresholds: Thresholds{
				MinTimeDeltaSeconds: -60,
			},
			expectError:   true,
			errorContains: "MinTimeDeltaSeconds cannot be negative",
		},
		{
			name: "zero values with one valid",
			thresholds: Thresholds{
				SuspiciousAdditions: 0,
				SuspiciousDeletions: 0,
				MaxAdditionsPerMin:  100.0,
				MaxDeletionsPerMin:  0,
				MinTimeDeltaSeconds: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.thresholds.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestThresholds_IsZero(t *testing.T) {
	tests := []struct {
		name       string
		thresholds Thresholds
		want       bool
	}{
		{
			name:       "all zeros",
			thresholds: Thresholds{},
			want:       true,
		},
		{
			name: "has additions",
			thresholds: Thresholds{
				SuspiciousAdditions: 100,
			},
			want: false,
		},
		{
			name: "has deletions",
			thresholds: Thresholds{
				SuspiciousDeletions: 200,
			},
			want: false,
		},
		{
			name: "has additions per min",
			thresholds: Thresholds{
				MaxAdditionsPerMin: 50.0,
			},
			want: false,
		},
		{
			name: "has deletions per min",
			thresholds: Thresholds{
				MaxDeletionsPerMin: 100.0,
			},
			want: false,
		},
		{
			name: "has time delta",
			thresholds: Thresholds{
				MinTimeDeltaSeconds: 60,
			},
			want: false,
		},
		{
			name: "all set",
			thresholds: Thresholds{
				SuspiciousAdditions: 100,
				SuspiciousDeletions: 200,
				MaxAdditionsPerMin:  50.0,
				MaxDeletionsPerMin:  100.0,
				MinTimeDeltaSeconds: 60,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.thresholds.IsZero()
			if got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
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
