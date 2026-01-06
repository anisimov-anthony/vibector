package reporter

import (
	"fmt"
	"strings"
	"time"

	"github.com/anisimov-anthony/vibector/internal/detector"
)

type TextReporter struct{}

func (r *TextReporter) Generate(data *ReportData) (string, error) {
	var sb strings.Builder

	sb.WriteString("----------------------------------------------\n")
	sb.WriteString("|            VIBECTOR ANALYSIS REPORT        |\n")
	sb.WriteString("----------------------------------------------\n\n")

	sb.WriteString("REPOSITORY STATISTICS\n")
	sb.WriteString("---------------------\n")
	sb.WriteString(fmt.Sprintf("Total Commits:         %d\n", data.Stats.TotalCommits))
	sb.WriteString(fmt.Sprintf("Commit Pairs:          %d\n", data.Stats.TotalCommitPairs))
	sb.WriteString(fmt.Sprintf("Unique Authors:        %d\n", data.Stats.UniqueAuthors))
	sb.WriteString(fmt.Sprintf("Time Span:             %s\n\n", formatDuration(data.Stats.TimeSpan)))

	sb.WriteString("Lines of Code (Filtered - Used for Detection):\n")
	sb.WriteString(fmt.Sprintf("  Additions:           %d lines\n", data.Stats.TotalLOCAdded))
	sb.WriteString(fmt.Sprintf("  Deletions:           %d lines\n\n", data.Stats.TotalLOCDeleted))

	sb.WriteString("Lines of Code (Total - Including Excluded Files):\n")
	sb.WriteString(fmt.Sprintf("  Additions:           %d lines\n", data.Stats.UnfilteredLOCAdded))
	sb.WriteString(fmt.Sprintf("  Deletions:           %d lines\n\n", data.Stats.UnfilteredLOCDeleted))

	sb.WriteString("VELOCITY STATISTICS\n")
	sb.WriteString("-------------------\n")
	sb.WriteString(fmt.Sprintf("Average Velocity:   %.2f LOC/min\n", data.Stats.AverageVelocity))
	sb.WriteString(fmt.Sprintf("Median Velocity:    %.2f LOC/min\n\n", data.Stats.MedianVelocity))

	if data.Stats.VelocityPercentile != nil {
		sb.WriteString("Velocity Percentiles:\n")
		sb.WriteString(fmt.Sprintf("  50th:   %.2f LOC/min\n", data.Stats.VelocityPercentile.P50))
		sb.WriteString(fmt.Sprintf("  75th:   %.2f LOC/min\n", data.Stats.VelocityPercentile.P75))
		sb.WriteString(fmt.Sprintf("  90th:   %.2f LOC/min\n", data.Stats.VelocityPercentile.P90))
		sb.WriteString(fmt.Sprintf("  95th:   %.2f LOC/min\n", data.Stats.VelocityPercentile.P95))
		sb.WriteString(fmt.Sprintf("  99th:   %.2f LOC/min\n\n", data.Stats.VelocityPercentile.P99))
	}

	sb.WriteString("CONFIGURED THRESHOLDS\n")
	sb.WriteString("---------------------\n")
	sb.WriteString(fmt.Sprintf("Suspicious Additions:   %d lines (0 = disabled)\n", data.Thresholds.SuspiciousAdditions))
	sb.WriteString(fmt.Sprintf("Suspicious Deletions:   %d lines (0 = disabled)\n", data.Thresholds.SuspiciousDeletions))
	sb.WriteString(fmt.Sprintf("Max Additions/min:      %.2f additions/min (0 = disabled)\n", data.Thresholds.MaxAdditionsPerMin))
	sb.WriteString(fmt.Sprintf("Max Deletions/min:      %.2f deletions/min (0 = disabled)\n", data.Thresholds.MaxDeletionsPerMin))
	sb.WriteString(fmt.Sprintf("Min Time Delta:         %d seconds (0 = disabled)\n", data.Thresholds.MinTimeDeltaSeconds))
	sb.WriteString("\n")

	sb.WriteString("SUSPICIOUS COMMITS\n")
	sb.WriteString("!!!!!!!!!!!!!!!!!!\n")

	if len(data.Suspicious) == 0 {
		sb.WriteString("No suspicious commits detected.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found %d suspicious commit(s):\n\n", len(data.Suspicious)))

		for i, s := range data.Suspicious {
			sb.WriteString(fmt.Sprintf("[%d] Commit: %s\n", i+1, s.Pair.Current.Hash[:7]))
			sb.WriteString(fmt.Sprintf("    Author:          %s <%s>\n", s.Pair.Current.Author, s.Pair.Current.Email))
			sb.WriteString(fmt.Sprintf("    Date:            %s\n", s.Pair.Current.Timestamp.Format(time.RFC3339)))
			sb.WriteString(fmt.Sprintf("    Additions:       %d lines (filtered) / %d lines (total)\n", s.Pair.Stats.Additions, s.Pair.Stats.TotalAdditions))
			sb.WriteString(fmt.Sprintf("    Deletions:       %d lines (filtered) / %d lines (total)\n", s.Pair.Stats.Deletions, s.Pair.Stats.TotalDeletions))
			sb.WriteString(fmt.Sprintf("    Files Changed:   %d (filtered) / %d (total)\n", s.Pair.Stats.FilesChanged, s.Pair.Stats.FilesChangedTotal))
			sb.WriteString(fmt.Sprintf("    Time Delta:      %s\n", detector.FormatTimeDelta(s.Pair.TimeDelta)))
			if s.AdditionVelocity != nil {
				sb.WriteString(fmt.Sprintf("    Add Velocity:    %.2f additions/min\n", s.AdditionVelocity.LOCPerMinute))
			}
			if s.DeletionVelocity != nil {
				sb.WriteString(fmt.Sprintf("    Del Velocity:    %.2f deletions/min\n", s.DeletionVelocity.LOCPerMinute))
			}
			sb.WriteString(fmt.Sprintf("    Message:         %s\n", truncate(s.Pair.Current.Message, 60)))
			sb.WriteString("    Reasons:\n")
			for _, reason := range s.Reasons {
				sb.WriteString(fmt.Sprintf("      - %s\n", reason))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func formatDuration(d time.Duration) string {
	minutes := d.Minutes()
	return fmt.Sprintf("%.0f minutes", minutes)
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
