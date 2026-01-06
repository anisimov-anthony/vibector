package reporter

import (
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/detector"
	"github.com/anisimov-anthony/vibector/internal/git"
	"github.com/anisimov-anthony/vibector/internal/metrics"
)

func TestTextReporter_Generate(t *testing.T) {
	now := time.Now()

	t.Run("generates report with no suspicious commits", func(t *testing.T) {
		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{},
			Stats: &metrics.RepositoryStats{
				TotalCommits:     10,
				TotalCommitPairs: 9,
				UniqueAuthors:    2,
				TimeSpan:         60 * time.Minute,
				TotalLOCAdded:    500,
				TotalLOCDeleted:  200,
				AverageVelocity:  25.5,
				MedianVelocity:   20.0,
				VelocityPercentile: &metrics.Percentiles{
					P50: 20.0,
					P75: 30.0,
					P90: 40.0,
					P95: 45.0,
					P99: 50.0,
				},
			},
			Thresholds: &detector.Thresholds{
				SuspiciousAdditions: 100,
				SuspiciousDeletions: 200,
				MaxAdditionsPerMin:  50.0,
				MaxDeletionsPerMin:  100.0,
				MinTimeDeltaSeconds: 60,
			},
		}

		reporter := &TextReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}
		if output == "" {
			t.Fatal("Generate() returned empty output")
		}

		expectedStrings := []string{
			"VIBECTOR ANALYSIS REPORT",
			"REPOSITORY STATISTICS",
			"Total Commits:         10",
			"Unique Authors:        2",
			"VELOCITY STATISTICS",
			"Average Velocity:   25.50 LOC/min",
			"Median Velocity:    20.00 LOC/min",
			"CONFIGURED THRESHOLDS",
			"SUSPICIOUS COMMITS",
			"No suspicious commits detected",
		}

		for _, expected := range expectedStrings {
			if !contains(output, expected) {
				t.Errorf("Output missing expected string: %s", expected)
			}
		}
	})

	t.Run("generates report with suspicious commits", func(t *testing.T) {
		commit := &git.Commit{
			Hash:      "abc123456789",
			Author:    "John Doe",
			Email:     "john@example.com",
			Timestamp: now,
			Message:   "Add new feature with lots of code",
		}

		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{
				{
					Pair: &git.CommitPair{
						Previous:  &git.Commit{Hash: "previous123"},
						Current:   commit,
						TimeDelta: 5 * time.Minute,
						Stats: &git.DiffStats{
							Additions:         500,
							Deletions:         50,
							TotalAdditions:    600,
							TotalDeletions:    60,
							FilesChanged:      10,
							FilesChangedTotal: 12,
						},
					},
					AdditionVelocity: &metrics.VelocityMetrics{
						LOCPerMinute: 100.0,
					},
					DeletionVelocity: &metrics.VelocityMetrics{
						LOCPerMinute: 10.0,
					},
					Reasons: []string{
						"Suspicious commit size: 500 additions (threshold: 100 lines)",
						"Addition velocity too high: 100.0 additions/min (threshold: 50.0 additions/min)",
					},
				},
			},
			Stats: &metrics.RepositoryStats{
				TotalCommits:     10,
				TotalCommitPairs: 9,
				UniqueAuthors:    2,
				TimeSpan:         60 * time.Minute,
				TotalLOCAdded:    500,
				TotalLOCDeleted:  200,
				AverageVelocity:  25.5,
				MedianVelocity:   20.0,
			},
			Thresholds: &detector.Thresholds{
				SuspiciousAdditions: 100,
				MaxAdditionsPerMin:  50.0,
			},
		}

		reporter := &TextReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}

		expectedStrings := []string{
			"Found 1 suspicious commit(s)",
			"[1] Commit: abc1234",
			"John Doe",
			"john@example.com",
			"Additions:       500 lines",
			"Deletions:       50 lines",
			"Files Changed:   10",
			"Add Velocity:    100.00 additions/min",
			"Del Velocity:    10.00 deletions/min",
			"Suspicious commit size",
			"Addition velocity too high",
		}

		for _, expected := range expectedStrings {
			if !contains(output, expected) {
				t.Errorf("Output missing expected string: %s", expected)
			}
		}
	})

	t.Run("generates report without velocity percentiles", func(t *testing.T) {
		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{},
			Stats: &metrics.RepositoryStats{
				TotalCommits:       5,
				TotalCommitPairs:   4,
				UniqueAuthors:      1,
				TimeSpan:           30 * time.Minute,
				VelocityPercentile: nil,
			},
			Thresholds: &detector.Thresholds{
				SuspiciousAdditions: 100,
			},
		}

		reporter := &TextReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}
		if output == "" {
			t.Fatal("Generate() returned empty output")
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string no truncation",
			input:  "Hello",
			maxLen: 10,
			want:   "Hello",
		},
		{
			name:   "exact length no truncation",
			input:  "Hello",
			maxLen: 5,
			want:   "Hello",
		},
		{
			name:   "long string truncated",
			input:  "This is a very long string that needs truncation",
			maxLen: 20,
			want:   "This is a very lo...",
		},
		{
			name:   "string with newlines",
			input:  "Line 1\nLine 2\nLine 3",
			maxLen: 50,
			want:   "Line 1 Line 2 Line 3",
		},
		{
			name:   "string with leading/trailing whitespace",
			input:  "  Hello World  ",
			maxLen: 50,
			want:   "Hello World",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "whitespace only",
			input:  "   \n  \n  ",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero duration",
			duration: 0,
			want:     "0 minutes",
		},
		{
			name:     "one minute",
			duration: 1 * time.Minute,
			want:     "1 minutes",
		},
		{
			name:     "fractional minutes rounded",
			duration: 90 * time.Second,
			want:     "2 minutes",
		},
		{
			name:     "many minutes",
			duration: 120 * time.Minute,
			want:     "120 minutes",
		},
		{
			name:     "less than one minute",
			duration: 30 * time.Second,
			want:     "0 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
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
