package reporter

import (
	"testing"
)

func TestNewReporter(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		wantType    string
		expectError bool
	}{
		{
			name:        "text reporter",
			format:      "text",
			wantType:    "*reporter.TextReporter",
			expectError: false,
		},
		{
			name:        "json reporter",
			format:      "json",
			wantType:    "*reporter.JSONReporter",
			expectError: false,
		},
		{
			name:        "invalid format",
			format:      "xml",
			expectError: true,
		},
		{
			name:        "empty format",
			format:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReporter(tt.format)
			if tt.expectError {
				if err == nil {
					t.Errorf("NewReporter() expected error but got none")
				}
				if got != nil {
					t.Errorf("NewReporter() expected nil on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Errorf("NewReporter() unexpected error = %v", err)
				return
			}
			if got == nil {
				t.Error("NewReporter() returned nil")
			}
		})
	}
}
