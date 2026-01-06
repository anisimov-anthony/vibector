package reporter

import (
	"encoding/json"
	"time"
)

type JSONReporter struct{}

type JSONReport struct {
	Statistics        JSONStats              `json:"statistics"`
	Thresholds        JSONThresholds         `json:"thresholds"`
	SuspiciousCount   int                    `json:"suspicious_count"`
	SuspiciousCommits []JSONSuspiciousCommit `json:"suspicious_commits"`
}

type JSONStats struct {
	TotalCommits         int              `json:"total_commits"`
	CommitPairs          int              `json:"commit_pairs"`
	UniqueAuthors        int              `json:"unique_authors"`
	TimeSpanSeconds      float64          `json:"time_span_seconds"`
	TotalLOCAdded        int64            `json:"total_loc_added_filtered"`
	TotalLOCDeleted      int64            `json:"total_loc_deleted_filtered"`
	UnfilteredLOCAdded   int64            `json:"total_loc_added_unfiltered"`
	UnfilteredLOCDeleted int64            `json:"total_loc_deleted_unfiltered"`
	AverageVelocity      float64          `json:"average_velocity_loc_per_min"`
	MedianVelocity       float64          `json:"median_velocity_loc_per_min"`
	VelocityPercentiles  *JSONPercentiles `json:"velocity_percentiles,omitempty"`
}

type JSONPercentiles struct {
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

type JSONThresholds struct {
	SuspiciousAdditions int64   `json:"suspicious_additions"`
	SuspiciousDeletions int64   `json:"suspicious_deletions"`
	MaxAdditionsPerMin  float64 `json:"max_additions_per_min"`
	MaxDeletionsPerMin  float64 `json:"max_deletions_per_min"`
	MinTimeDeltaSeconds int64   `json:"min_time_delta_seconds"`
}

type JSONSuspiciousCommit struct {
	Hash                string   `json:"hash"`
	Author              string   `json:"author"`
	Email               string   `json:"email"`
	Timestamp           string   `json:"timestamp"`
	Message             string   `json:"message"`
	Additions           int64    `json:"additions_filtered"`
	Deletions           int64    `json:"deletions_filtered"`
	TotalAdditions      int64    `json:"additions_total"`
	TotalDeletions      int64    `json:"deletions_total"`
	FilesChanged        int      `json:"files_changed_filtered"`
	FilesChangedTotal   int      `json:"files_changed_total"`
	TimeDelta           float64  `json:"time_delta_seconds"`
	AdditionVelocityMin float64  `json:"addition_velocity_per_min"`
	DeletionVelocityMin float64  `json:"deletion_velocity_per_min"`
	Reasons             []string `json:"reasons"`
}

func (r *JSONReporter) Generate(data *ReportData) (string, error) {
	report := JSONReport{
		Statistics: JSONStats{
			TotalCommits:         data.Stats.TotalCommits,
			CommitPairs:          data.Stats.TotalCommitPairs,
			UniqueAuthors:        data.Stats.UniqueAuthors,
			TimeSpanSeconds:      data.Stats.TimeSpan.Seconds(),
			TotalLOCAdded:        data.Stats.TotalLOCAdded,
			TotalLOCDeleted:      data.Stats.TotalLOCDeleted,
			UnfilteredLOCAdded:   data.Stats.UnfilteredLOCAdded,
			UnfilteredLOCDeleted: data.Stats.UnfilteredLOCDeleted,
			AverageVelocity:      data.Stats.AverageVelocity,
			MedianVelocity:       data.Stats.MedianVelocity,
		},
		Thresholds: JSONThresholds{
			SuspiciousAdditions: data.Thresholds.SuspiciousAdditions,
			SuspiciousDeletions: data.Thresholds.SuspiciousDeletions,
			MaxAdditionsPerMin:  data.Thresholds.MaxAdditionsPerMin,
			MaxDeletionsPerMin:  data.Thresholds.MaxDeletionsPerMin,
			MinTimeDeltaSeconds: data.Thresholds.MinTimeDeltaSeconds,
		},
		SuspiciousCount:   len(data.Suspicious),
		SuspiciousCommits: make([]JSONSuspiciousCommit, len(data.Suspicious)),
	}

	if data.Stats.VelocityPercentile != nil {
		report.Statistics.VelocityPercentiles = &JSONPercentiles{
			P50: data.Stats.VelocityPercentile.P50,
			P75: data.Stats.VelocityPercentile.P75,
			P90: data.Stats.VelocityPercentile.P90,
			P95: data.Stats.VelocityPercentile.P95,
			P99: data.Stats.VelocityPercentile.P99,
		}
	}

	for i, s := range data.Suspicious {
		commit := JSONSuspiciousCommit{
			Hash:              s.Pair.Current.Hash,
			Author:            s.Pair.Current.Author,
			Email:             s.Pair.Current.Email,
			Timestamp:         s.Pair.Current.Timestamp.Format(time.RFC3339),
			Message:           s.Pair.Current.Message,
			Additions:         s.Pair.Stats.Additions,
			Deletions:         s.Pair.Stats.Deletions,
			TotalAdditions:    s.Pair.Stats.TotalAdditions,
			TotalDeletions:    s.Pair.Stats.TotalDeletions,
			FilesChanged:      s.Pair.Stats.FilesChanged,
			FilesChangedTotal: s.Pair.Stats.FilesChangedTotal,
			TimeDelta:         s.Pair.TimeDelta.Seconds(),
			Reasons:           s.Reasons,
		}
		if s.AdditionVelocity != nil {
			commit.AdditionVelocityMin = s.AdditionVelocity.LOCPerMinute
		}
		if s.DeletionVelocity != nil {
			commit.DeletionVelocityMin = s.DeletionVelocity.LOCPerMinute
		}
		report.SuspiciousCommits[i] = commit
	}

	bytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
