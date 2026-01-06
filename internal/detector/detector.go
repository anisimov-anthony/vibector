package detector

import (
	"fmt"
	"time"

	"github.com/anisimov-anthony/vibector/internal/git"
	"github.com/anisimov-anthony/vibector/internal/metrics"
)

type SuspiciousCommit struct {
	Pair             *git.CommitPair
	AdditionVelocity *metrics.VelocityMetrics
	DeletionVelocity *metrics.VelocityMetrics
	Reasons          []string
}

type Detector struct {
	thresholds *Thresholds
}

func New(thresholds *Thresholds) (*Detector, error) {
	if err := thresholds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid thresholds: %w", err)
	}

	return &Detector{
		thresholds: thresholds,
	}, nil
}

func (d *Detector) DetectSuspicious(pairs []*git.CommitPair, repoStats *metrics.RepositoryStats) []*SuspiciousCommit {
	if pairs == nil {
		return []*SuspiciousCommit{}
	}

	suspicious := make([]*SuspiciousCommit, 0)

	for _, pair := range pairs {
		if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
			continue
		}

		reasons := make([]string, 0)

		var additionVelocity, deletionVelocity *metrics.VelocityMetrics
		if d.thresholds.MaxAdditionsPerMin > 0 || d.thresholds.MaxDeletionsPerMin > 0 {
			var err error
			additionVelocity, err = metrics.CalculateVelocity(pair.Stats.Additions, pair.TimeDelta)
			if err != nil {
				continue
			}
			deletionVelocity, err = metrics.CalculateVelocity(pair.Stats.Deletions, pair.TimeDelta)
			if err != nil {
				continue
			}
		}

		if d.thresholds.MinTimeDeltaSeconds > 0 {
			if pair.TimeDelta.Seconds() < float64(d.thresholds.MinTimeDeltaSeconds) {
				reasons = append(reasons, fmt.Sprintf(
					"Time between commits too short: %.1f seconds (threshold: %d seconds)",
					pair.TimeDelta.Seconds(),
					d.thresholds.MinTimeDeltaSeconds,
				))
			}
		}

		if d.thresholds.SuspiciousAdditions > 0 {
			if pair.Stats.Additions > d.thresholds.SuspiciousAdditions {
				reasons = append(reasons, fmt.Sprintf(
					"Suspicious commit size: %d additions (threshold: %d lines)",
					pair.Stats.Additions,
					d.thresholds.SuspiciousAdditions,
				))
			}
		}

		if d.thresholds.SuspiciousDeletions > 0 {
			if pair.Stats.Deletions > d.thresholds.SuspiciousDeletions {
				reasons = append(reasons, fmt.Sprintf(
					"Suspicious commit size: %d deletions (threshold: %d lines)",
					pair.Stats.Deletions,
					d.thresholds.SuspiciousDeletions,
				))
			}
		}

		if d.thresholds.MaxAdditionsPerMin > 0 && additionVelocity != nil {
			if additionVelocity.LOCPerMinute > d.thresholds.MaxAdditionsPerMin {
				reasons = append(reasons, fmt.Sprintf(
					"Addition velocity too high: %.1f additions/min (threshold: %.1f additions/min)",
					additionVelocity.LOCPerMinute,
					d.thresholds.MaxAdditionsPerMin,
				))
			}
		}

		if d.thresholds.MaxDeletionsPerMin > 0 && deletionVelocity != nil {
			if deletionVelocity.LOCPerMinute > d.thresholds.MaxDeletionsPerMin {
				reasons = append(reasons, fmt.Sprintf(
					"Deletion velocity too high: %.1f deletions/min (threshold: %.1f deletions/min)",
					deletionVelocity.LOCPerMinute,
					d.thresholds.MaxDeletionsPerMin,
				))
			}
		}

		if len(reasons) > 0 {
			suspicious = append(suspicious, &SuspiciousCommit{
				Pair:             pair,
				AdditionVelocity: additionVelocity,
				DeletionVelocity: deletionVelocity,
				Reasons:          reasons,
			})
		}
	}

	return suspicious
}

func FormatTimeDelta(d time.Duration) string {
	minutes := d.Minutes()
	if minutes < 1 {
		return fmt.Sprintf("%.2f minutes", minutes)
	}
	return fmt.Sprintf("%.1f minutes", minutes)
}
