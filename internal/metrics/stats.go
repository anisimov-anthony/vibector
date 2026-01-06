package metrics

import (
	"math"
	"sort"
	"time"

	"github.com/anisimov-anthony/vibector/internal/git"
)

type RepositoryStats struct {
	TotalCommits         int
	TotalCommitPairs     int
	UniqueAuthors        int
	Authors              map[string]*AuthorStats
	TimeSpan             time.Duration
	FirstCommit          time.Time
	LastCommit           time.Time
	TotalLOCAdded        int64
	TotalLOCDeleted      int64
	UnfilteredLOCAdded   int64
	UnfilteredLOCDeleted int64
	AverageVelocity      float64
	MedianVelocity       float64
	VelocityPercentile   *Percentiles
}

type AuthorStats struct {
	Name        string
	Email       string
	CommitCount int
	LOCAdded    int64
	LOCDeleted  int64
	AvgVelocity float64
	MaxVelocity float64
	FirstCommit time.Time
	LastCommit  time.Time
}

type Percentiles struct {
	P50 float64
	P75 float64
	P90 float64
	P95 float64
	P99 float64
}

func CalculateStats(commits []*git.Commit, pairs []*git.CommitPair) *RepositoryStats {
	if commits == nil {
		commits = []*git.Commit{}
	}
	if pairs == nil {
		pairs = []*git.CommitPair{}
	}

	stats := &RepositoryStats{
		TotalCommits:     len(commits),
		TotalCommitPairs: len(pairs),
		Authors:          make(map[string]*AuthorStats),
	}

	if len(commits) == 0 {
		return stats
	}

	authorSet := make(map[string]bool)
	for _, commit := range commits {
		authorKey := commit.Email
		authorSet[authorKey] = true

		if stats.FirstCommit.IsZero() || commit.Timestamp.Before(stats.FirstCommit) {
			stats.FirstCommit = commit.Timestamp
		}
		if stats.LastCommit.IsZero() || commit.Timestamp.After(stats.LastCommit) {
			stats.LastCommit = commit.Timestamp
		}

		if _, exists := stats.Authors[authorKey]; !exists {
			stats.Authors[authorKey] = &AuthorStats{
				Name:  commit.Author,
				Email: commit.Email,
			}
		}
	}

	stats.UniqueAuthors = len(authorSet)
	stats.TimeSpan = stats.LastCommit.Sub(stats.FirstCommit)

	velocities := make([]float64, 0, len(pairs))

	for _, pair := range pairs {
		stats.TotalLOCAdded += pair.Stats.Additions
		stats.TotalLOCDeleted += pair.Stats.Deletions

		stats.UnfilteredLOCAdded += pair.Stats.TotalAdditions
		stats.UnfilteredLOCDeleted += pair.Stats.TotalDeletions

		hasFilteredChanges := pair.Stats.Additions > 0 || pair.Stats.Deletions > 0

		if hasFilteredChanges {
			velocity, err := CalculateVelocityPerMinute(pair.Stats.Additions, pair.TimeDelta)
			if err == nil && !math.IsNaN(velocity) && velocity >= 0 {
				velocities = append(velocities, velocity)
			}

			authorKey := pair.Current.Email
			authorStats, exists := stats.Authors[authorKey]
			if !exists {
				authorStats = &AuthorStats{
					Name:  pair.Current.Author,
					Email: pair.Current.Email,
				}
				stats.Authors[authorKey] = authorStats
			}

			authorStats.CommitCount++
			authorStats.LOCAdded += pair.Stats.Additions
			authorStats.LOCDeleted += pair.Stats.Deletions

			if authorStats.FirstCommit.IsZero() || pair.Current.Timestamp.Before(authorStats.FirstCommit) {
				authorStats.FirstCommit = pair.Current.Timestamp
			}
			if authorStats.LastCommit.IsZero() || pair.Current.Timestamp.After(authorStats.LastCommit) {
				authorStats.LastCommit = pair.Current.Timestamp
			}

			if err == nil && !math.IsNaN(velocity) && velocity >= 0 {
				if velocity > authorStats.MaxVelocity {
					authorStats.MaxVelocity = velocity
				}
			}
		}
	}

	if len(velocities) > 0 {
		sum := 0.0
		for _, v := range velocities {
			sum += v
		}
		stats.AverageVelocity = sum / float64(len(velocities))
		stats.MedianVelocity = calculateMedian(velocities)
		stats.VelocityPercentile = calculatePercentiles(velocities)
	}

	for email, authorStats := range stats.Authors {
		if authorStats.CommitCount > 0 {
			authorVelocities := make([]float64, 0)
			for _, pair := range pairs {
				if pair.Current.Email == email {
					if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
						continue
					}

					velocity, err := CalculateVelocityPerMinute(pair.Stats.Additions, pair.TimeDelta)
					if err == nil && !math.IsNaN(velocity) && velocity >= 0 {
						authorVelocities = append(authorVelocities, velocity)
					}
				}
			}

			if len(authorVelocities) > 0 {
				sum := 0.0
				for _, v := range authorVelocities {
					sum += v
				}
				authorStats.AvgVelocity = sum / float64(len(authorVelocities))
			}
		}
	}

	return stats
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func calculatePercentiles(values []float64) *Percentiles {
	if len(values) == 0 {
		return &Percentiles{}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	return &Percentiles{
		P50: percentile(sorted, 50),
		P75: percentile(sorted, 75),
		P90: percentile(sorted, 90),
		P95: percentile(sorted, 95),
		P99: percentile(sorted, 99),
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	rank := (p / 100.0) * float64(len(sorted)-1)
	lowerIndex := int(math.Floor(rank))
	upperIndex := int(math.Ceil(rank))

	if lowerIndex == upperIndex {
		return sorted[lowerIndex]
	}

	weight := rank - float64(lowerIndex)
	return sorted[lowerIndex]*(1-weight) + sorted[upperIndex]*weight
}
