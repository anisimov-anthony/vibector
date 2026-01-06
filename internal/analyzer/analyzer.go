package analyzer

import (
	"fmt"

	"github.com/anisimov-anthony/vibector/internal/git"
)

type Analyzer struct {
	repo git.Repository
}

type AnalysisResult struct {
	Commits      []*git.Commit
	CommitPairs  []*git.CommitPair
	TotalCommits int
}

func New(repo git.Repository) *Analyzer {
	return &Analyzer{
		repo: repo,
	}
}

func (a *Analyzer) AnalyzeRepository(opts *git.CommitOptions) (*AnalysisResult, error) {
	if opts == nil {
		opts = &git.CommitOptions{}
	}

	commits, err := a.repo.GetCommits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commits: %w", err)
	}

	if len(commits) == 0 {
		return &AnalysisResult{
			Commits:      commits,
			CommitPairs:  []*git.CommitPair{},
			TotalCommits: 0,
		}, nil
	}

	pairs, err := a.createCommitPairs(commits)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit pairs: %w", err)
	}

	return &AnalysisResult{
		Commits:      commits,
		CommitPairs:  pairs,
		TotalCommits: len(commits),
	}, nil
}

func (a *Analyzer) createCommitPairs(commits []*git.Commit) ([]*git.CommitPair, error) {
	if repo, ok := a.repo.(interface {
		GetCommitPairs([]*git.Commit) ([]*git.CommitPair, error)
	}); ok {
		return repo.GetCommitPairs(commits)
	}

	return nil, fmt.Errorf("repository does not support commit pair creation")
}
