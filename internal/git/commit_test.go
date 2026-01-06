package git

import (
	"testing"
	"time"
)

func TestCommitStructure(t *testing.T) {
	now := time.Now()

	t.Run("create commit", func(t *testing.T) {
		commit := &Commit{
			Hash:      "abc123",
			Author:    "John Doe",
			Email:     "john@example.com",
			Timestamp: now,
			Message:   "Initial commit",
			Parents:   []string{"parent1", "parent2"},
		}

		if commit.Hash != "abc123" {
			t.Errorf("Hash = %s, want abc123", commit.Hash)
		}
		if commit.Author != "John Doe" {
			t.Errorf("Author = %s, want John Doe", commit.Author)
		}
		if commit.Email != "john@example.com" {
			t.Errorf("Email = %s, want john@example.com", commit.Email)
		}
		if !commit.Timestamp.Equal(now) {
			t.Errorf("Timestamp = %v, want %v", commit.Timestamp, now)
		}
		if commit.Message != "Initial commit" {
			t.Errorf("Message = %s, want Initial commit", commit.Message)
		}
		if len(commit.Parents) != 2 {
			t.Errorf("len(Parents) = %d, want 2", len(commit.Parents))
		}
	})

	t.Run("commit with no parents", func(t *testing.T) {
		commit := &Commit{
			Hash:      "root123",
			Author:    "Jane Doe",
			Email:     "jane@example.com",
			Timestamp: now,
			Message:   "Root commit",
			Parents:   []string{},
		}

		if len(commit.Parents) != 0 {
			t.Errorf("len(Parents) = %d, want 0", len(commit.Parents))
		}
	})
}

func TestCommitPairStructure(t *testing.T) {
	now := time.Now()

	previous := &Commit{
		Hash:      "prev123",
		Author:    "John Doe",
		Email:     "john@example.com",
		Timestamp: now.Add(-10 * time.Minute),
	}

	current := &Commit{
		Hash:      "curr456",
		Author:    "John Doe",
		Email:     "john@example.com",
		Timestamp: now,
	}

	stats := &DiffStats{
		FilesChanged:      5,
		Additions:         100,
		Deletions:         50,
		TotalAdditions:    120,
		TotalDeletions:    60,
		FilesChangedTotal: 6,
	}

	t.Run("create commit pair", func(t *testing.T) {
		pair := &CommitPair{
			Previous:  previous,
			Current:   current,
			TimeDelta: 10 * time.Minute,
			Stats:     stats,
		}

		if pair.Previous != previous {
			t.Error("Previous commit not set correctly")
		}
		if pair.Current != current {
			t.Error("Current commit not set correctly")
		}
		if pair.TimeDelta != 10*time.Minute {
			t.Errorf("TimeDelta = %v, want 10m", pair.TimeDelta)
		}
		if pair.Stats != stats {
			t.Error("Stats not set correctly")
		}
	})
}

func TestDiffStatsStructure(t *testing.T) {
	t.Run("create diff stats", func(t *testing.T) {
		stats := &DiffStats{
			FilesChanged:      5,
			Additions:         100,
			Deletions:         50,
			TotalAdditions:    120,
			TotalDeletions:    60,
			FilesChangedTotal: 6,
		}

		if stats.FilesChanged != 5 {
			t.Errorf("FilesChanged = %d, want 5", stats.FilesChanged)
		}
		if stats.Additions != 100 {
			t.Errorf("Additions = %d, want 100", stats.Additions)
		}
		if stats.Deletions != 50 {
			t.Errorf("Deletions = %d, want 50", stats.Deletions)
		}
		if stats.TotalAdditions != 120 {
			t.Errorf("TotalAdditions = %d, want 120", stats.TotalAdditions)
		}
		if stats.TotalDeletions != 60 {
			t.Errorf("TotalDeletions = %d, want 60", stats.TotalDeletions)
		}
		if stats.FilesChangedTotal != 6 {
			t.Errorf("FilesChangedTotal = %d, want 6", stats.FilesChangedTotal)
		}
	})

	t.Run("zero diff stats", func(t *testing.T) {
		stats := &DiffStats{}

		if stats.FilesChanged != 0 {
			t.Errorf("FilesChanged = %d, want 0", stats.FilesChanged)
		}
		if stats.Additions != 0 {
			t.Errorf("Additions = %d, want 0", stats.Additions)
		}
		if stats.Deletions != 0 {
			t.Errorf("Deletions = %d, want 0", stats.Deletions)
		}
	})
}

func TestCommitOptionsStructure(t *testing.T) {
	t.Run("create commit options", func(t *testing.T) {
		opts := &CommitOptions{
			Branch:   "main",
			MaxDepth: 100,
		}

		if opts.Branch != "main" {
			t.Errorf("Branch = %s, want main", opts.Branch)
		}
		if opts.MaxDepth != 100 {
			t.Errorf("MaxDepth = %d, want 100", opts.MaxDepth)
		}
	})

	t.Run("empty commit options", func(t *testing.T) {
		opts := &CommitOptions{}

		if opts.Branch != "" {
			t.Errorf("Branch = %s, want empty", opts.Branch)
		}
		if opts.MaxDepth != 0 {
			t.Errorf("MaxDepth = %d, want 0", opts.MaxDepth)
		}
	})
}
