package metrics

import (
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/git"
)

func TestCalculateStats(t *testing.T) {
	now := time.Now()

	t.Run("empty commits and pairs", func(t *testing.T) {
		stats := CalculateStats(nil, nil)
		if stats.TotalCommits != 0 {
			t.Errorf("TotalCommits = %d, want 0", stats.TotalCommits)
		}
		if stats.TotalCommitPairs != 0 {
			t.Errorf("TotalCommitPairs = %d, want 0", stats.TotalCommitPairs)
		}
		if stats.UniqueAuthors != 0 {
			t.Errorf("UniqueAuthors = %d, want 0", stats.UniqueAuthors)
		}
	})

	t.Run("single commit no pairs", func(t *testing.T) {
		commits := []*git.Commit{
			{
				Hash:      "abc123",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now,
				Message:   "Initial commit",
			},
		}
		stats := CalculateStats(commits, nil)
		if stats.TotalCommits != 1 {
			t.Errorf("TotalCommits = %d, want 1", stats.TotalCommits)
		}
		if stats.UniqueAuthors != 1 {
			t.Errorf("UniqueAuthors = %d, want 1", stats.UniqueAuthors)
		}
		if stats.TotalCommitPairs != 0 {
			t.Errorf("TotalCommitPairs = %d, want 0", stats.TotalCommitPairs)
		}
	})

	t.Run("multiple commits with pairs", func(t *testing.T) {
		commits := []*git.Commit{
			{
				Hash:      "abc123",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now.Add(-20 * time.Minute),
			},
			{
				Hash:      "def456",
				Author:    "Jane Smith",
				Email:     "jane@example.com",
				Timestamp: now.Add(-10 * time.Minute),
			},
			{
				Hash:      "ghi789",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now,
			},
		}

		pairs := []*git.CommitPair{
			{
				Previous:  commits[0],
				Current:   commits[1],
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions:         100,
					Deletions:         50,
					TotalAdditions:    120,
					TotalDeletions:    60,
					FilesChanged:      5,
					FilesChangedTotal: 6,
				},
			},
			{
				Previous:  commits[1],
				Current:   commits[2],
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions:         200,
					Deletions:         100,
					TotalAdditions:    250,
					TotalDeletions:    120,
					FilesChanged:      8,
					FilesChangedTotal: 10,
				},
			},
		}

		stats := CalculateStats(commits, pairs)

		if stats.TotalCommits != 3 {
			t.Errorf("TotalCommits = %d, want 3", stats.TotalCommits)
		}
		if stats.TotalCommitPairs != 2 {
			t.Errorf("TotalCommitPairs = %d, want 2", stats.TotalCommitPairs)
		}
		if stats.UniqueAuthors != 2 {
			t.Errorf("UniqueAuthors = %d, want 2", stats.UniqueAuthors)
		}
		if stats.TotalLOCAdded != 300 {
			t.Errorf("TotalLOCAdded = %d, want 300", stats.TotalLOCAdded)
		}
		if stats.TotalLOCDeleted != 150 {
			t.Errorf("TotalLOCDeleted = %d, want 150", stats.TotalLOCDeleted)
		}
		if stats.UnfilteredLOCAdded != 370 {
			t.Errorf("UnfilteredLOCAdded = %d, want 370", stats.UnfilteredLOCAdded)
		}
		if stats.UnfilteredLOCDeleted != 180 {
			t.Errorf("UnfilteredLOCDeleted = %d, want 180", stats.UnfilteredLOCDeleted)
		}

		expectedTimeSpan := 20 * time.Minute
		if stats.TimeSpan != expectedTimeSpan {
			t.Errorf("TimeSpan = %v, want %v", stats.TimeSpan, expectedTimeSpan)
		}

		if stats.AverageVelocity == 0 {
			t.Errorf("AverageVelocity should not be 0")
		}
		if stats.MedianVelocity == 0 {
			t.Errorf("MedianVelocity should not be 0")
		}
		if stats.VelocityPercentile == nil {
			t.Errorf("VelocityPercentile should not be nil")
		}
	})

	t.Run("author statistics", func(t *testing.T) {
		commits := []*git.Commit{
			{
				Hash:      "abc123",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now.Add(-10 * time.Minute),
			},
			{
				Hash:      "def456",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now,
			},
		}

		pairs := []*git.CommitPair{
			{
				Previous:  commits[0],
				Current:   commits[1],
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions:    100,
					Deletions:    50,
					FilesChanged: 5,
				},
			},
		}

		stats := CalculateStats(commits, pairs)

		authorStats, exists := stats.Authors["john@example.com"]
		if !exists {
			t.Fatal("Author john@example.com not found")
		}

		if authorStats.Name != "John Doe" {
			t.Errorf("Author name = %s, want John Doe", authorStats.Name)
		}
		if authorStats.CommitCount != 1 {
			t.Errorf("CommitCount = %d, want 1", authorStats.CommitCount)
		}
		if authorStats.LOCAdded != 100 {
			t.Errorf("LOCAdded = %d, want 100", authorStats.LOCAdded)
		}
		if authorStats.LOCDeleted != 50 {
			t.Errorf("LOCDeleted = %d, want 50", authorStats.LOCDeleted)
		}
		if authorStats.AvgVelocity == 0 {
			t.Errorf("AvgVelocity should not be 0")
		}
		if authorStats.MaxVelocity == 0 {
			t.Errorf("MaxVelocity should not be 0")
		}
	})

	t.Run("filters out commits with no changes", func(t *testing.T) {
		commits := []*git.Commit{
			{
				Hash:      "abc123",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now,
			},
			{
				Hash:      "def456",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now.Add(10 * time.Minute),
			},
		}

		pairs := []*git.CommitPair{
			{
				Previous:  commits[0],
				Current:   commits[1],
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions:    0,
					Deletions:    0,
					FilesChanged: 0,
				},
			},
		}

		stats := CalculateStats(commits, pairs)

		if stats.AverageVelocity != 0 {
			t.Errorf("AverageVelocity = %f, want 0 (should skip zero-change commits)", stats.AverageVelocity)
		}
		if stats.MedianVelocity != 0 {
			t.Errorf("MedianVelocity = %f, want 0 (should skip zero-change commits)", stats.MedianVelocity)
		}
	})
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{
			name:   "empty slice",
			values: []float64{},
			want:   0,
		},
		{
			name:   "single value",
			values: []float64{5.0},
			want:   5.0,
		},
		{
			name:   "odd number of values",
			values: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			want:   3.0,
		},
		{
			name:   "even number of values",
			values: []float64{1.0, 2.0, 3.0, 4.0},
			want:   2.5,
		},
		{
			name:   "unsorted values odd",
			values: []float64{5.0, 1.0, 3.0, 4.0, 2.0},
			want:   3.0,
		},
		{
			name:   "unsorted values even",
			values: []float64{4.0, 1.0, 3.0, 2.0},
			want:   2.5,
		},
		{
			name:   "duplicate values",
			values: []float64{2.0, 2.0, 2.0, 2.0},
			want:   2.0,
		},
		{
			name:   "negative values",
			values: []float64{-5.0, -3.0, -1.0, 1.0, 3.0},
			want:   -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMedian(tt.values)
			if got != tt.want {
				t.Errorf("calculateMedian() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculatePercentiles(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   *Percentiles
	}{
		{
			name:   "empty slice",
			values: []float64{},
			want: &Percentiles{
				P50: 0,
				P75: 0,
				P90: 0,
				P95: 0,
				P99: 0,
			},
		},
		{
			name:   "single value",
			values: []float64{10.0},
			want: &Percentiles{
				P50: 10.0,
				P75: 10.0,
				P90: 10.0,
				P95: 10.0,
				P99: 10.0,
			},
		},
		{
			name:   "sequential values 1-10",
			values: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			want: &Percentiles{
				P50: 5.5,
				P75: 7.75,
				P90: 9.1,
				P95: 9.55,
				P99: 9.91,
			},
		},
		{
			name:   "two values",
			values: []float64{1.0, 10.0},
			want: &Percentiles{
				P50: 5.5,
				P75: 7.75,
				P90: 9.1,
				P95: 9.55,
				P99: 9.91,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePercentiles(tt.values)
			if got == nil {
				t.Fatal("calculatePercentiles() returned nil")
			}

			const epsilon = 0.01
			if abs(got.P50-tt.want.P50) > epsilon {
				t.Errorf("P50 = %v, want %v", got.P50, tt.want.P50)
			}
			if abs(got.P75-tt.want.P75) > epsilon {
				t.Errorf("P75 = %v, want %v", got.P75, tt.want.P75)
			}
			if abs(got.P90-tt.want.P90) > epsilon {
				t.Errorf("P90 = %v, want %v", got.P90, tt.want.P90)
			}
			if abs(got.P95-tt.want.P95) > epsilon {
				t.Errorf("P95 = %v, want %v", got.P95, tt.want.P95)
			}
			if abs(got.P99-tt.want.P99) > epsilon {
				t.Errorf("P99 = %v, want %v", got.P99, tt.want.P99)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []float64
		p      float64
		want   float64
	}{
		{
			name:   "empty slice",
			sorted: []float64{},
			p:      50,
			want:   0,
		},
		{
			name:   "50th percentile of 1-10",
			sorted: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			p:      50,
			want:   5.5,
		},
		{
			name:   "75th percentile of 1-10",
			sorted: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			p:      75,
			want:   7.75,
		},
		{
			name:   "0th percentile (min)",
			sorted: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:      0,
			want:   1.0,
		},
		{
			name:   "100th percentile (max)",
			sorted: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:      100,
			want:   5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(tt.sorted, tt.p)
			const epsilon = 0.01
			if abs(got-tt.want) > epsilon {
				t.Errorf("percentile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
