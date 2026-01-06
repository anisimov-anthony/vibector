package git

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type RepositoryOptions struct {
	ExcludeFiles []string
}

type Repository interface {
	GetCommits(opts *CommitOptions) ([]*Commit, error)
	Close() error
}

type gitRepository struct {
	repo         *git.Repository
	path         string
	excludeFiles []string
}

func OpenRepository(path string, opts *RepositoryOptions) (Repository, error) {
	if path == "" {
		return nil, fmt.Errorf("repository path cannot be empty")
	}

	if opts == nil {
		opts = &RepositoryOptions{}
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open local repository: %w", err)
	}

	return &gitRepository{
		repo:         r,
		path:         path,
		excludeFiles: opts.ExcludeFiles,
	}, nil
}

func (r *gitRepository) GetCommits(opts *CommitOptions) ([]*Commit, error) {
	if opts == nil {
		opts = &CommitOptions{}
	}

	var ref *plumbing.Reference
	var err error

	if opts.Branch != "" {
		ref, err = r.repo.Reference(plumbing.ReferenceName("refs/heads/"+opts.Branch), true)
		if err != nil {
			return nil, fmt.Errorf("failed to get branch reference: %w", err)
		}
	} else {
		ref, err = r.repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", err)
		}
	}

	commitIter, err := r.repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit iterator: %w", err)
	}
	defer commitIter.Close()

	commits := make([]*Commit, 0)
	count := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		if opts.MaxDepth > 0 && count >= opts.MaxDepth {
			return io.EOF
		}

		parents := make([]string, len(c.ParentHashes))
		for i, p := range c.ParentHashes {
			parents[i] = p.String()
		}

		commits = append(commits, &Commit{
			Hash:      c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
			Message:   c.Message,
			Parents:   parents,
		})

		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	return commits, nil
}

func (r *gitRepository) GetCommitPairs(commits []*Commit) ([]*CommitPair, error) {
	if len(commits) < 2 {
		return []*CommitPair{}, nil
	}

	pairs := make([]*CommitPair, 0)

	for i := 0; i < len(commits)-1; i++ {
		current := commits[i]
		previous := commits[i+1]

		if len(current.Parents) > 1 {
			continue
		}

		timeDelta := current.Timestamp.Sub(previous.Timestamp)
		if timeDelta <= 0 {
			continue
		}

		stats, err := r.getDiffStats(previous.Hash, current.Hash)
		if err != nil {
			continue
		}

		pairs = append(pairs, &CommitPair{
			Previous:  previous,
			Current:   current,
			TimeDelta: timeDelta,
			Stats:     stats,
		})
	}

	return pairs, nil
}

func (r *gitRepository) shouldExcludeFile(filePath string) bool {
	if len(r.excludeFiles) == 0 {
		return false
	}

	for _, pattern := range r.excludeFiles {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}

		matched, err = filepath.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (r *gitRepository) getDiffStats(fromHash, toHash string) (*DiffStats, error) {
	fromCommit, err := r.repo.CommitObject(plumbing.NewHash(fromHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get from commit: %w", err)
	}

	toCommit, err := r.repo.CommitObject(plumbing.NewHash(toHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get to commit: %w", err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get from tree: %w", err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get to tree: %w", err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	stats := &DiffStats{}
	filesChanged := make(map[string]bool)
	filesChangedTotal := make(map[string]bool)

	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}

		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()

			var filePath string
			if to != nil {
				filePath = to.Path()
			} else if from != nil {
				filePath = from.Path()
			}

			if from != nil {
				filesChangedTotal[from.Path()] = true
			}
			if to != nil {
				filesChangedTotal[to.Path()] = true
			}

			isExcluded := r.shouldExcludeFile(filePath)

			if !isExcluded {
				if from != nil {
					filesChanged[from.Path()] = true
				}
				if to != nil {
					filesChanged[to.Path()] = true
				}
			}

			chunks := filePatch.Chunks()
			for _, chunk := range chunks {
				lines := strings.Split(chunk.Content(), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					switch chunk.Type() {
					case diff.Add:
						stats.TotalAdditions++
						if !isExcluded {
							stats.Additions++
						}
					case diff.Delete:
						stats.TotalDeletions++
						if !isExcluded {
							stats.Deletions++
						}
					}
				}
			}
		}
	}

	stats.FilesChanged = len(filesChanged)
	stats.FilesChangedTotal = len(filesChangedTotal)

	return stats, nil
}

func (r *gitRepository) Close() error {
	return nil
}
