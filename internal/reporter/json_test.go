package reporter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/detector"
	"github.com/anisimov-anthony/vibector/internal/git"
	"github.com/anisimov-anthony/vibector/internal/metrics"
)

func TestJSONReporter_Generate(t *testing.T) {
	now := time.Now()

	t.Run("generates valid JSON with no suspicious commits", func(t *testing.T) {
		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{},
			Stats: &metrics.RepositoryStats{
				TotalCommits:         10,
				TotalCommitPairs:     9,
				UniqueAuthors:        2,
				TimeSpan:             60 * time.Minute,
				TotalLOCAdded:        500,
				TotalLOCDeleted:      200,
				UnfilteredLOCAdded:   600,
				UnfilteredLOCDeleted: 250,
				AverageVelocity:      25.5,
				MedianVelocity:       20.0,
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

		reporter := &JSONReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}
		if output == "" {
			t.Fatal("Generate() returned empty output")
		}

		var result JSONReport
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Generated JSON is invalid: %v", err)
		}

		if result.Statistics.TotalCommits != 10 {
			t.Errorf("TotalCommits = %d, want 10", result.Statistics.TotalCommits)
		}
		if result.Statistics.CommitPairs != 9 {
			t.Errorf("CommitPairs = %d, want 9", result.Statistics.CommitPairs)
		}
		if result.Statistics.UniqueAuthors != 2 {
			t.Errorf("UniqueAuthors = %d, want 2", result.Statistics.UniqueAuthors)
		}
		if result.Statistics.TotalLOCAdded != 500 {
			t.Errorf("TotalLOCAdded = %d, want 500", result.Statistics.TotalLOCAdded)
		}
		if result.Statistics.TotalLOCDeleted != 200 {
			t.Errorf("TotalLOCDeleted = %d, want 200", result.Statistics.TotalLOCDeleted)
		}
		if result.Statistics.UnfilteredLOCAdded != 600 {
			t.Errorf("UnfilteredLOCAdded = %d, want 600", result.Statistics.UnfilteredLOCAdded)
		}
		if result.Statistics.UnfilteredLOCDeleted != 250 {
			t.Errorf("UnfilteredLOCDeleted = %d, want 250", result.Statistics.UnfilteredLOCDeleted)
		}
		if result.Statistics.AverageVelocity != 25.5 {
			t.Errorf("AverageVelocity = %f, want 25.5", result.Statistics.AverageVelocity)
		}
		if result.Statistics.MedianVelocity != 20.0 {
			t.Errorf("MedianVelocity = %f, want 20.0", result.Statistics.MedianVelocity)
		}

		if result.Statistics.VelocityPercentiles == nil {
			t.Fatal("VelocityPercentiles should not be nil")
		}
		if result.Statistics.VelocityPercentiles.P50 != 20.0 {
			t.Errorf("P50 = %f, want 20.0", result.Statistics.VelocityPercentiles.P50)
		}
		if result.Statistics.VelocityPercentiles.P75 != 30.0 {
			t.Errorf("P75 = %f, want 30.0", result.Statistics.VelocityPercentiles.P75)
		}
		if result.Statistics.VelocityPercentiles.P90 != 40.0 {
			t.Errorf("P90 = %f, want 40.0", result.Statistics.VelocityPercentiles.P90)
		}
		if result.Statistics.VelocityPercentiles.P95 != 45.0 {
			t.Errorf("P95 = %f, want 45.0", result.Statistics.VelocityPercentiles.P95)
		}
		if result.Statistics.VelocityPercentiles.P99 != 50.0 {
			t.Errorf("P99 = %f, want 50.0", result.Statistics.VelocityPercentiles.P99)
		}

		if result.Thresholds.SuspiciousAdditions != 100 {
			t.Errorf("SuspiciousAdditions = %d, want 100", result.Thresholds.SuspiciousAdditions)
		}
		if result.Thresholds.SuspiciousDeletions != 200 {
			t.Errorf("SuspiciousDeletions = %d, want 200", result.Thresholds.SuspiciousDeletions)
		}
		if result.Thresholds.MaxAdditionsPerMin != 50.0 {
			t.Errorf("MaxAdditionsPerMin = %f, want 50.0", result.Thresholds.MaxAdditionsPerMin)
		}
		if result.Thresholds.MaxDeletionsPerMin != 100.0 {
			t.Errorf("MaxDeletionsPerMin = %f, want 100.0", result.Thresholds.MaxDeletionsPerMin)
		}
		if result.Thresholds.MinTimeDeltaSeconds != 60 {
			t.Errorf("MinTimeDeltaSeconds = %d, want 60", result.Thresholds.MinTimeDeltaSeconds)
		}

		if result.SuspiciousCount != 0 {
			t.Errorf("SuspiciousCount = %d, want 0", result.SuspiciousCount)
		}
		if len(result.SuspiciousCommits) != 0 {
			t.Errorf("len(SuspiciousCommits) = %d, want 0", len(result.SuspiciousCommits))
		}
	})

	t.Run("generates valid JSON with suspicious commits", func(t *testing.T) {
		commit := &git.Commit{
			Hash:      "abc123456789",
			Author:    "John Doe",
			Email:     "john@example.com",
			Timestamp: now,
			Message:   "Add new feature",
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
						"Suspicious commit size",
						"Addition velocity too high",
					},
				},
			},
			Stats: &metrics.RepositoryStats{
				TotalCommits:     10,
				TotalCommitPairs: 9,
				UniqueAuthors:    2,
				TimeSpan:         60 * time.Minute,
			},
			Thresholds: &detector.Thresholds{
				SuspiciousAdditions: 100,
			},
		}

		reporter := &JSONReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}

		var result JSONReport
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Generated JSON is invalid: %v", err)
		}

		if result.SuspiciousCount != 1 {
			t.Errorf("SuspiciousCount = %d, want 1", result.SuspiciousCount)
		}
		if len(result.SuspiciousCommits) != 1 {
			t.Fatalf("len(SuspiciousCommits) = %d, want 1", len(result.SuspiciousCommits))
		}

		sc := result.SuspiciousCommits[0]
		if sc.Hash != "abc123456789" {
			t.Errorf("Hash = %s, want abc123456789", sc.Hash)
		}
		if sc.Author != "John Doe" {
			t.Errorf("Author = %s, want John Doe", sc.Author)
		}
		if sc.Email != "john@example.com" {
			t.Errorf("Email = %s, want john@example.com", sc.Email)
		}
		if sc.Message != "Add new feature" {
			t.Errorf("Message = %s, want Add new feature", sc.Message)
		}
		if sc.Additions != 500 {
			t.Errorf("Additions = %d, want 500", sc.Additions)
		}
		if sc.Deletions != 50 {
			t.Errorf("Deletions = %d, want 50", sc.Deletions)
		}
		if sc.TotalAdditions != 600 {
			t.Errorf("TotalAdditions = %d, want 600", sc.TotalAdditions)
		}
		if sc.TotalDeletions != 60 {
			t.Errorf("TotalDeletions = %d, want 60", sc.TotalDeletions)
		}
		if sc.FilesChanged != 10 {
			t.Errorf("FilesChanged = %d, want 10", sc.FilesChanged)
		}
		if sc.FilesChangedTotal != 12 {
			t.Errorf("FilesChangedTotal = %d, want 12", sc.FilesChangedTotal)
		}
		if sc.TimeDelta != 300.0 {
			t.Errorf("TimeDelta = %f, want 300.0 (5 minutes)", sc.TimeDelta)
		}
		if sc.AdditionVelocityMin != 100.0 {
			t.Errorf("AdditionVelocityMin = %f, want 100.0", sc.AdditionVelocityMin)
		}
		if sc.DeletionVelocityMin != 10.0 {
			t.Errorf("DeletionVelocityMin = %f, want 10.0", sc.DeletionVelocityMin)
		}
		if len(sc.Reasons) != 2 {
			t.Errorf("len(Reasons) = %d, want 2", len(sc.Reasons))
		}
	})

	t.Run("generates JSON without velocity percentiles", func(t *testing.T) {
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

		reporter := &JSONReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}

		var result JSONReport
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Generated JSON is invalid: %v", err)
		}

		if result.Statistics.VelocityPercentiles != nil {
			t.Error("VelocityPercentiles should be nil when not provided")
		}
	})

	t.Run("generates JSON with nil velocity metrics", func(t *testing.T) {
		commit := &git.Commit{
			Hash:      "abc123",
			Timestamp: now,
		}

		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{
				{
					Pair: &git.CommitPair{
						Previous:  &git.Commit{Hash: "prev123"},
						Current:   commit,
						TimeDelta: 5 * time.Minute,
						Stats: &git.DiffStats{
							Additions: 100,
							Deletions: 50,
						},
					},
					AdditionVelocity: nil,
					DeletionVelocity: nil,
					Reasons:          []string{"Some reason"},
				},
			},
			Stats: &metrics.RepositoryStats{
				TotalCommits:     1,
				TotalCommitPairs: 0,
			},
			Thresholds: &detector.Thresholds{
				MinTimeDeltaSeconds: 60,
			},
		}

		reporter := &JSONReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}

		var result JSONReport
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Generated JSON is invalid: %v", err)
		}

		if len(result.SuspiciousCommits) != 1 {
			t.Fatalf("len(SuspiciousCommits) = %d, want 1", len(result.SuspiciousCommits))
		}

		sc := result.SuspiciousCommits[0]
		if sc.AdditionVelocityMin != 0 {
			t.Errorf("AdditionVelocityMin = %f, want 0 when nil", sc.AdditionVelocityMin)
		}
		if sc.DeletionVelocityMin != 0 {
			t.Errorf("DeletionVelocityMin = %f, want 0 when nil", sc.DeletionVelocityMin)
		}
	})

	t.Run("timestamp formatting", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		commit := &git.Commit{
			Hash:      "abc123",
			Timestamp: testTime,
		}

		data := &ReportData{
			Suspicious: []*detector.SuspiciousCommit{
				{
					Pair: &git.CommitPair{
						Previous:  &git.Commit{Hash: "prev123"},
						Current:   commit,
						TimeDelta: 5 * time.Minute,
						Stats: &git.DiffStats{
							Additions: 100,
						},
					},
					Reasons: []string{"Test"},
				},
			},
			Stats:      &metrics.RepositoryStats{},
			Thresholds: &detector.Thresholds{SuspiciousAdditions: 50},
		}

		reporter := &JSONReporter{}
		output, err := reporter.Generate(data)

		if err != nil {
			t.Fatalf("Generate() unexpected error = %v", err)
		}

		var result JSONReport
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Generated JSON is invalid: %v", err)
		}

		expectedTimestamp := testTime.Format(time.RFC3339)
		if result.SuspiciousCommits[0].Timestamp != expectedTimestamp {
			t.Errorf("Timestamp = %s, want %s", result.SuspiciousCommits[0].Timestamp, expectedTimestamp)
		}
	})
}
