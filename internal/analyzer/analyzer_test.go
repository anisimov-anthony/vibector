package analyzer

import (
	"fmt"
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/git"
)

// mockRepository implements the git.Repository interface for testing
type mockRepository struct {
	commits      []*git.Commit
	commitPairs  []*git.CommitPair
	commitErr    error
	pairErr      error
	supportPairs bool
}

func (m *mockRepository) GetCommits(opts *git.CommitOptions) ([]*git.Commit, error) {
	if m.commitErr != nil {
		return nil, m.commitErr
	}
	return m.commits, nil
}

func (m *mockRepository) GetCommitPairs(commits []*git.Commit) ([]*git.CommitPair, error) {
	if !m.supportPairs {
		return nil, fmt.Errorf("repository does not support commit pair creation")
	}
	if m.pairErr != nil {
		return nil, m.pairErr
	}
	return m.commitPairs, nil
}

func (m *mockRepository) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	repo := &mockRepository{}
	analyzer := New(repo)

	if analyzer == nil {
		t.Fatal("New() returned nil")
	}
	if analyzer.repo != repo {
		t.Error("New() did not set repository correctly")
	}
}

func TestAnalyzer_AnalyzeRepository(t *testing.T) {
	now := time.Now()

	t.Run("successful analysis with commits and pairs", func(t *testing.T) {
		commits := []*git.Commit{
			{
				Hash:      "abc123",
				Author:    "John Doe",
				Email:     "john@example.com",
				Timestamp: now,
			},
			{
				Hash:      "def456",
				Author:    "Jane Smith",
				Email:     "jane@example.com",
				Timestamp: now.Add(-10 * time.Minute),
			},
		}

		pairs := []*git.CommitPair{
			{
				Previous:  commits[1],
				Current:   commits[0],
				TimeDelta: 10 * time.Minute,
				Stats: &git.DiffStats{
					Additions: 100,
					Deletions: 50,
				},
			},
		}

		repo := &mockRepository{
			commits:      commits,
			commitPairs:  pairs,
			supportPairs: true,
		}

		analyzer := New(repo)
		result, err := analyzer.AnalyzeRepository(nil)

		if err != nil {
			t.Fatalf("AnalyzeRepository() unexpected error = %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeRepository() returned nil result")
		}
		if result.TotalCommits != 2 {
			t.Errorf("TotalCommits = %d, want 2", result.TotalCommits)
		}
		if len(result.Commits) != 2 {
			t.Errorf("len(Commits) = %d, want 2", len(result.Commits))
		}
		if len(result.CommitPairs) != 1 {
			t.Errorf("len(CommitPairs) = %d, want 1", len(result.CommitPairs))
		}
	})

	t.Run("empty repository", func(t *testing.T) {
		repo := &mockRepository{
			commits:      []*git.Commit{},
			commitPairs:  []*git.CommitPair{},
			supportPairs: true,
		}

		analyzer := New(repo)
		result, err := analyzer.AnalyzeRepository(nil)

		if err != nil {
			t.Fatalf("AnalyzeRepository() unexpected error = %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeRepository() returned nil result")
		}
		if result.TotalCommits != 0 {
			t.Errorf("TotalCommits = %d, want 0", result.TotalCommits)
		}
		if len(result.Commits) != 0 {
			t.Errorf("len(Commits) = %d, want 0", len(result.Commits))
		}
		if len(result.CommitPairs) != 0 {
			t.Errorf("len(CommitPairs) = %d, want 0", len(result.CommitPairs))
		}
	})

	t.Run("error retrieving commits", func(t *testing.T) {
		repo := &mockRepository{
			commitErr: fmt.Errorf("failed to get commits"),
		}

		analyzer := New(repo)
		result, err := analyzer.AnalyzeRepository(nil)

		if err == nil {
			t.Fatal("AnalyzeRepository() expected error but got none")
		}
		if result != nil {
			t.Errorf("AnalyzeRepository() expected nil result on error, got %v", result)
		}
		if !contains(err.Error(), "failed to retrieve commits") {
			t.Errorf("error = %v, want error containing 'failed to retrieve commits'", err)
		}
	})

	t.Run("error creating commit pairs", func(t *testing.T) {
		commits := []*git.Commit{
			{Hash: "abc123"},
			{Hash: "def456"},
		}

		repo := &mockRepository{
			commits:      commits,
			pairErr:      fmt.Errorf("failed to create pairs"),
			supportPairs: true,
		}

		analyzer := New(repo)
		result, err := analyzer.AnalyzeRepository(nil)

		if err == nil {
			t.Fatal("AnalyzeRepository() expected error but got none")
		}
		if result != nil {
			t.Errorf("AnalyzeRepository() expected nil result on error, got %v", result)
		}
		if !contains(err.Error(), "failed to create commit pairs") {
			t.Errorf("error = %v, want error containing 'failed to create commit pairs'", err)
		}
	})

	t.Run("repository does not support pair creation", func(t *testing.T) {
		commits := []*git.Commit{
			{Hash: "abc123"},
			{Hash: "def456"},
		}

		repo := &mockRepository{
			commits:      commits,
			supportPairs: false,
		}

		analyzer := New(repo)
		_, err := analyzer.AnalyzeRepository(nil)

		if err == nil {
			t.Fatal("AnalyzeRepository() expected error but got none")
		}
		if !contains(err.Error(), "does not support commit pair creation") {
			t.Errorf("error = %v, want error containing 'does not support commit pair creation'", err)
		}
	})

	t.Run("with commit options", func(t *testing.T) {
		commits := []*git.Commit{
			{Hash: "abc123"},
		}

		repo := &mockRepository{
			commits:      commits,
			commitPairs:  []*git.CommitPair{},
			supportPairs: true,
		}

		analyzer := New(repo)
		opts := &git.CommitOptions{
			Branch:   "main",
			MaxDepth: 10,
		}
		result, err := analyzer.AnalyzeRepository(opts)

		if err != nil {
			t.Fatalf("AnalyzeRepository() unexpected error = %v", err)
		}
		if result == nil {
			t.Fatal("AnalyzeRepository() returned nil result")
		}
		if result.TotalCommits != 1 {
			t.Errorf("TotalCommits = %d, want 1", result.TotalCommits)
		}
	})

	t.Run("single commit no pairs", func(t *testing.T) {
		commits := []*git.Commit{
			{Hash: "abc123"},
		}

		repo := &mockRepository{
			commits:      commits,
			commitPairs:  []*git.CommitPair{},
			supportPairs: true,
		}

		analyzer := New(repo)
		result, err := analyzer.AnalyzeRepository(nil)

		if err != nil {
			t.Fatalf("AnalyzeRepository() unexpected error = %v", err)
		}
		if result.TotalCommits != 1 {
			t.Errorf("TotalCommits = %d, want 1", result.TotalCommits)
		}
		if len(result.CommitPairs) != 0 {
			t.Errorf("len(CommitPairs) = %d, want 0", len(result.CommitPairs))
		}
	})
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
