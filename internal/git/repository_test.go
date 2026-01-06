package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// createTestRepo creates a test git repository with commits
func createTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config git name: %v", err)
	}

	// Create first commit
	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("line1\nline2\nline3\n"), 0o600); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	cmd = exec.Command("git", "add", "file1.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Create second commit
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("new line 1\nnew line 2\n"), 0o600); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	cmd = exec.Command("git", "add", "file2.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add file2")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Create third commit with deletion
	if err := os.WriteFile(file1, []byte("line1\n"), 0o600); err != nil {
		t.Fatalf("Failed to modify file1: %v", err)
	}

	cmd = exec.Command("git", "add", "file1.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Delete lines from file1")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	return tmpDir
}

func TestOpenRepository(t *testing.T) {
	t.Run("open valid repository", func(t *testing.T) {
		repoPath := createTestRepo(t)

		repo, err := OpenRepository(repoPath, nil)
		if err != nil {
			t.Fatalf("OpenRepository() unexpected error = %v", err)
		}
		if repo == nil {
			t.Fatal("OpenRepository() returned nil")
		}

		defer repo.Close()
	})

	t.Run("empty path returns error", func(t *testing.T) {
		repo, err := OpenRepository("", nil)
		if err == nil {
			t.Fatal("OpenRepository(\"\") expected error but got none")
		}
		if repo != nil {
			t.Errorf("OpenRepository(\"\") expected nil on error, got %v", repo)
		}
		if !contains(err.Error(), "path cannot be empty") {
			t.Errorf("error = %v, want error containing 'path cannot be empty'", err)
		}
	})

	t.Run("non-existent path returns error", func(t *testing.T) {
		repo, err := OpenRepository("/non/existent/path", nil)
		if err == nil {
			t.Fatal("OpenRepository() expected error for non-existent path but got none")
		}
		if repo != nil {
			t.Errorf("OpenRepository() expected nil on error, got %v", repo)
		}
	})

	t.Run("non-git directory returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		repo, err := OpenRepository(tmpDir, nil)
		if err == nil {
			t.Fatal("OpenRepository() expected error for non-git directory but got none")
		}
		if repo != nil {
			t.Errorf("OpenRepository() expected nil on error, got %v", repo)
		}
	})

	t.Run("open with exclude files option", func(t *testing.T) {
		repoPath := createTestRepo(t)

		opts := &RepositoryOptions{
			ExcludeFiles: []string{"*.log", "*.tmp"},
		}
		repo, err := OpenRepository(repoPath, opts)
		if err != nil {
			t.Fatalf("OpenRepository() unexpected error = %v", err)
		}
		if repo == nil {
			t.Fatal("OpenRepository() returned nil")
		}

		defer repo.Close()
	})
}

func TestGitRepository_GetCommits(t *testing.T) {
	repoPath := createTestRepo(t)
	repo, err := OpenRepository(repoPath, nil)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	t.Run("get all commits", func(t *testing.T) {
		commits, err := repo.GetCommits(nil)
		if err != nil {
			t.Fatalf("GetCommits() unexpected error = %v", err)
		}
		if len(commits) != 3 {
			t.Errorf("len(commits) = %d, want 3", len(commits))
		}

		// Verify commits are in reverse chronological order (newest first)
		for i := 0; i < len(commits)-1; i++ {
			if commits[i].Timestamp.Before(commits[i+1].Timestamp) {
				t.Error("Commits should be in reverse chronological order")
			}
		}

		// Verify commit fields
		for _, commit := range commits {
			if commit.Hash == "" {
				t.Error("Commit hash should not be empty")
			}
			if commit.Author == "" {
				t.Error("Commit author should not be empty")
			}
			if commit.Email == "" {
				t.Error("Commit email should not be empty")
			}
			if commit.Message == "" {
				t.Error("Commit message should not be empty")
			}
		}
	})

	t.Run("get commits with max depth", func(t *testing.T) {
		opts := &CommitOptions{
			MaxDepth: 2,
		}
		commits, err := repo.GetCommits(opts)
		if err != nil {
			t.Fatalf("GetCommits() unexpected error = %v", err)
		}
		if len(commits) != 2 {
			t.Errorf("len(commits) = %d, want 2", len(commits))
		}
	})

	t.Run("get commits with max depth 1", func(t *testing.T) {
		opts := &CommitOptions{
			MaxDepth: 1,
		}
		commits, err := repo.GetCommits(opts)
		if err != nil {
			t.Fatalf("GetCommits() unexpected error = %v", err)
		}
		if len(commits) != 1 {
			t.Errorf("len(commits) = %d, want 1", len(commits))
		}
	})

	t.Run("nil options works", func(t *testing.T) {
		commits, err := repo.GetCommits(nil)
		if err != nil {
			t.Fatalf("GetCommits(nil) unexpected error = %v", err)
		}
		if len(commits) == 0 {
			t.Error("GetCommits(nil) should return commits")
		}
	})
}

func TestGitRepository_GetCommitPairs(t *testing.T) {
	repoPath := createTestRepo(t)
	gitRepo, err := OpenRepository(repoPath, nil)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}
	defer gitRepo.Close()

	repo := gitRepo.(*gitRepository)

	t.Run("get commit pairs from commits", func(t *testing.T) {
		commits, err := repo.GetCommits(nil)
		if err != nil {
			t.Fatalf("GetCommits() error = %v", err)
		}

		pairs, err := repo.GetCommitPairs(commits)
		if err != nil {
			t.Fatalf("GetCommitPairs() unexpected error = %v", err)
		}

		if len(pairs) != 2 {
			t.Errorf("len(pairs) = %d, want 2", len(pairs))
		}

		for _, pair := range pairs {
			if pair.Previous == nil {
				t.Error("Previous commit should not be nil")
			}
			if pair.Current == nil {
				t.Error("Current commit should not be nil")
			}
			if pair.Stats == nil {
				t.Error("Stats should not be nil")
			}
			if pair.TimeDelta <= 0 {
				t.Error("TimeDelta should be positive")
			}
		}
	})

	t.Run("empty commits returns empty pairs", func(t *testing.T) {
		pairs, err := repo.GetCommitPairs([]*Commit{})
		if err != nil {
			t.Fatalf("GetCommitPairs() unexpected error = %v", err)
		}
		if len(pairs) != 0 {
			t.Errorf("len(pairs) = %d, want 0", len(pairs))
		}
	})

	t.Run("single commit returns empty pairs", func(t *testing.T) {
		commits := []*Commit{
			{Hash: "abc123", Timestamp: time.Now()},
		}
		pairs, err := repo.GetCommitPairs(commits)
		if err != nil {
			t.Fatalf("GetCommitPairs() unexpected error = %v", err)
		}
		if len(pairs) != 0 {
			t.Errorf("len(pairs) = %d, want 0", len(pairs))
		}
	})
}

func TestGitRepository_ShouldExcludeFile(t *testing.T) {
	repoPath := createTestRepo(t)

	t.Run("no exclusions", func(t *testing.T) {
		gitRepo, _ := OpenRepository(repoPath, nil)
		defer gitRepo.Close()
		repo := gitRepo.(*gitRepository)

		if repo.shouldExcludeFile("test.log") {
			t.Error("shouldExcludeFile() should return false when no exclusions")
		}
	})

	t.Run("with exclusion patterns", func(t *testing.T) {
		opts := &RepositoryOptions{
			ExcludeFiles: []string{"*.log", "*.tmp", "package-lock.json"},
		}
		gitRepo, _ := OpenRepository(repoPath, opts)
		defer gitRepo.Close()
		repo := gitRepo.(*gitRepository)

		tests := []struct {
			name     string
			filePath string
			want     bool
		}{
			{
				name:     "matches log pattern",
				filePath: "debug.log",
				want:     true,
			},
			{
				name:     "matches tmp pattern",
				filePath: "cache.tmp",
				want:     true,
			},
			{
				name:     "matches exact filename",
				filePath: "package-lock.json",
				want:     true,
			},
			{
				name:     "no match",
				filePath: "main.go",
				want:     false,
			},
			{
				name:     "matches full path",
				filePath: "logs/error.log",
				want:     true,
			},
			{
				name:     "matches basename only",
				filePath: "dir/subdir/test.tmp",
				want:     true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := repo.shouldExcludeFile(tt.filePath)
				if got != tt.want {
					t.Errorf("shouldExcludeFile(%q) = %v, want %v", tt.filePath, got, tt.want)
				}
			})
		}
	})
}

func TestGitRepository_Close(t *testing.T) {
	repoPath := createTestRepo(t)
	repo, err := OpenRepository(repoPath, nil)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Close() unexpected error = %v", err)
	}
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
