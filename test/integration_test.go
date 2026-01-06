package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/anisimov-anthony/vibector/internal/analyzer"
	"github.com/anisimov-anthony/vibector/internal/detector"
	"github.com/anisimov-anthony/vibector/internal/git"
	"github.com/anisimov-anthony/vibector/internal/metrics"
	"github.com/anisimov-anthony/vibector/internal/reporter"
)

// createIntegrationTestRepo creates a test git repository with realistic commit patterns
func createIntegrationTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "human@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Human Developer")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config git name: %v", err)
	}

	// Create normal human-speed commit
	file1 := filepath.Join(tmpDir, "main.go")
	content1 := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	if err := os.WriteFile(file1, []byte(content1), 0o600); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	cmd = exec.Command("git", "add", "main.go")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Create suspiciously large commit (AI-like)
	file2 := filepath.Join(tmpDir, "generated.go")
	largeContent := `package main

import (
	"fmt"
	"time"
)

`
	// Add 200 lines to simulate AI generation
	for i := 0; i < 200; i++ {
		largeContent += fmt.Sprintf("func Function%d() { fmt.Println(\"Function %d\") }\n", i, i)
	}

	if err := os.WriteFile(file2, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	cmd = exec.Command("git", "add", "generated.go")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add generated functions")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Create another normal commit
	file3 := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(file3, []byte("# Test Project\n"), 0o600); err != nil {
		t.Fatalf("Failed to write file3: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add README")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	return tmpDir
}

func TestFullWorkflow(t *testing.T) {
	repoPath := createIntegrationTestRepo(t)

	t.Run("complete analysis workflow", func(t *testing.T) {
		// Step 1: Open repository
		repo, err := git.OpenRepository(repoPath, nil)
		if err != nil {
			t.Fatalf("Failed to open repository: %v", err)
		}
		defer repo.Close()

		// Step 2: Analyze repository
		a := analyzer.New(repo)
		result, err := a.AnalyzeRepository(nil)
		if err != nil {
			t.Fatalf("Failed to analyze repository: %v", err)
		}

		if result.TotalCommits != 3 {
			t.Errorf("TotalCommits = %d, want 3", result.TotalCommits)
		}

		// Step 3: Calculate statistics
		stats := metrics.CalculateStats(result.Commits, result.CommitPairs)
		if stats.TotalCommits != 3 {
			t.Errorf("Stats TotalCommits = %d, want 3", stats.TotalCommits)
		}
		if stats.UniqueAuthors != 1 {
			t.Errorf("UniqueAuthors = %d, want 1", stats.UniqueAuthors)
		}

		// Step 4: Detect suspicious commits
		thresholds := &detector.Thresholds{
			SuspiciousAdditions: 100,
			MaxAdditionsPerMin:  50.0,
		}
		d, err := detector.New(thresholds)
		if err != nil {
			t.Fatalf("Failed to create detector: %v", err)
		}

		suspicious := d.DetectSuspicious(result.CommitPairs, stats)
		if len(suspicious) == 0 {
			t.Error("Expected to detect at least one suspicious commit")
		}

		// Step 5: Generate reports
		reportData := &reporter.ReportData{
			Suspicious: suspicious,
			Stats:      stats,
			Thresholds: thresholds,
		}

		// Test text report
		textReporter, err := reporter.NewReporter("text")
		if err != nil {
			t.Fatalf("Failed to create text reporter: %v", err)
		}
		textOutput, err := textReporter.Generate(reportData)
		if err != nil {
			t.Fatalf("Failed to generate text report: %v", err)
		}
		if textOutput == "" {
			t.Error("Text report should not be empty")
		}

		// Test JSON report
		jsonReporter, err := reporter.NewReporter("json")
		if err != nil {
			t.Fatalf("Failed to create JSON reporter: %v", err)
		}
		jsonOutput, err := jsonReporter.Generate(reportData)
		if err != nil {
			t.Fatalf("Failed to generate JSON report: %v", err)
		}
		if jsonOutput == "" {
			t.Error("JSON report should not be empty")
		}
	})

	t.Run("workflow with file exclusions", func(t *testing.T) {
		// Create test repo with log files
		tmpDir := t.TempDir()
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		// Create initial commit
		file1 := filepath.Join(tmpDir, "code.go")
		if err := os.WriteFile(file1, []byte("package main\n"), 0o600); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "Initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		time.Sleep(2 * time.Second)

		// Add log file (should be excluded)
		logFile := filepath.Join(tmpDir, "debug.log")
		logContent := ""
		for i := 0; i < 1000; i++ {
			logContent += "Log line\n"
		}
		if err := os.WriteFile(logFile, []byte(logContent), 0o600); err != nil {
			t.Fatalf("Failed to write log file: %v", err)
		}

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "Add logs")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		// Open with exclusions
		opts := &git.RepositoryOptions{
			ExcludeFiles: []string{"*.log"},
		}
		repo, err := git.OpenRepository(tmpDir, opts)
		if err != nil {
			t.Fatalf("Failed to open repository: %v", err)
		}
		defer repo.Close()

		a := analyzer.New(repo)
		result, err := a.AnalyzeRepository(nil)
		if err != nil {
			t.Fatalf("Failed to analyze repository: %v", err)
		}

		stats := metrics.CalculateStats(result.Commits, result.CommitPairs)

		// Filtered stats should exclude log file additions
		if stats.TotalLOCAdded >= 1000 {
			t.Errorf("TotalLOCAdded = %d, should be < 1000 (log file excluded)", stats.TotalLOCAdded)
		}

		// Unfiltered stats should include log file
		if stats.UnfilteredLOCAdded < 1000 {
			t.Errorf("UnfilteredLOCAdded = %d, should be >= 1000 (log file included)", stats.UnfilteredLOCAdded)
		}
	})
}

func TestEndToEndScenarios(t *testing.T) {
	t.Run("detect AI-like burst of commits", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "config", "user.email", "ai@example.com")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "AI Assistant")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		// Create rapid succession of commits (AI-like)
		for i := 0; i < 3; i++ {
			file := filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i))
			content := ""
			for j := 0; j < 100; j++ {
				content += fmt.Sprintf("line %d\n", j)
			}
			_ = os.WriteFile(file, []byte(content), 0o600)

			cmd = exec.Command("git", "add", ".")
			cmd.Dir = tmpDir
			_ = cmd.Run()

			cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Commit %d", i))
			cmd.Dir = tmpDir
			_ = cmd.Run()

			// Very short time between commits (but still measurable)
			time.Sleep(500 * time.Millisecond)
		}

		repo, err := git.OpenRepository(tmpDir, nil)
		if err != nil {
			t.Fatalf("Failed to open repository: %v", err)
		}
		defer repo.Close()

		a := analyzer.New(repo)
		result, err := a.AnalyzeRepository(nil)
		if err != nil {
			t.Fatalf("Failed to analyze: %v", err)
		}

		stats := metrics.CalculateStats(result.Commits, result.CommitPairs)

		// Should detect high velocity
		if stats.AverageVelocity == 0 {
			t.Error("Average velocity should not be zero")
		}

		thresholds := &detector.Thresholds{
			SuspiciousAdditions: 50,
			MaxAdditionsPerMin:  100.0,
		}
		d, _ := detector.New(thresholds)
		suspicious := d.DetectSuspicious(result.CommitPairs, stats)

		if len(suspicious) == 0 {
			t.Error("Should detect suspicious commits in AI-like burst")
		}
	})

	t.Run("no false positives for normal commits", func(t *testing.T) {
		t.Skip("Skipping due to long execution time")
		tmpDir := t.TempDir()
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "config", "user.email", "human@example.com")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		cmd = exec.Command("git", "config", "user.name", "Human")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		// Create normal small commits with reasonable sizes and timing
		for i := 0; i < 3; i++ {
			file := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
			content := ""
			// Normal-sized commits (20-50 lines)
			for j := 0; j < 50; j++ {
				content += fmt.Sprintf("Line %d in file %d\n", j, i)
			}
			_ = os.WriteFile(file, []byte(content), 0o600)

			cmd = exec.Command("git", "add", ".")
			cmd.Dir = tmpDir
			_ = cmd.Run()

			cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Small commit %d", i))
			cmd.Dir = tmpDir
			_ = cmd.Run()

			// Reasonable time between commits (2 minutes for 50 lines = 25 LOC/min)
			time.Sleep(2 * time.Minute)
		}

		repo, err := git.OpenRepository(tmpDir, nil)
		if err != nil {
			t.Fatalf("Failed to open repository: %v", err)
		}
		defer repo.Close()

		a := analyzer.New(repo)
		result, err := a.AnalyzeRepository(nil)
		if err != nil {
			t.Fatalf("Failed to analyze: %v", err)
		}

		stats := metrics.CalculateStats(result.Commits, result.CommitPairs)

		// Thresholds that shouldn't trigger for normal development
		thresholds := &detector.Thresholds{
			SuspiciousAdditions: 150,   // More than our 50 lines
			MaxAdditionsPerMin:  100.0, // More than our 25 LOC/min
			MinTimeDeltaSeconds: 10,    // Less than our 2 minutes
		}
		d, _ := detector.New(thresholds)
		suspicious := d.DetectSuspicious(result.CommitPairs, stats)

		if len(suspicious) > 0 {
			t.Errorf("Should not detect suspicious commits for normal activity, but found %d", len(suspicious))
			for _, s := range suspicious {
				t.Logf("  Suspicious: %v", s.Reasons)
			}
		}
	})
}
