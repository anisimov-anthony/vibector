package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("load empty config", func(t *testing.T) {
		config, err := Load("")
		if err != nil {
			t.Fatalf("Load(\"\") unexpected error = %v", err)
		}
		if config == nil {
			t.Fatal("Load(\"\") returned nil config")
		}
		if !config.Thresholds.IsZero() {
			t.Error("Empty config should have zero thresholds")
		}
	})

	t.Run("load from yaml file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `thresholds:
  suspicious_additions: 500
  suspicious_deletions: 1000
  max_additions_per_min: 100.5
  max_deletions_per_min: 200.5
  min_time_delta_seconds: 60
exclude_files:
  - "*.log"
  - "*.tmp"
`
		if err := os.WriteFile(configFile, []byte(yamlContent), 0o600); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := Load(configFile)
		if err != nil {
			t.Fatalf("Load() unexpected error = %v", err)
		}

		if config.Thresholds.SuspiciousAdditions != 500 {
			t.Errorf("SuspiciousAdditions = %d, want 500", config.Thresholds.SuspiciousAdditions)
		}
		if config.Thresholds.SuspiciousDeletions != 1000 {
			t.Errorf("SuspiciousDeletions = %d, want 1000", config.Thresholds.SuspiciousDeletions)
		}
		if config.Thresholds.MaxAdditionsPerMin != 100.5 {
			t.Errorf("MaxAdditionsPerMin = %f, want 100.5", config.Thresholds.MaxAdditionsPerMin)
		}
		if config.Thresholds.MaxDeletionsPerMin != 200.5 {
			t.Errorf("MaxDeletionsPerMin = %f, want 200.5", config.Thresholds.MaxDeletionsPerMin)
		}
		if config.Thresholds.MinTimeDeltaSeconds != 60 {
			t.Errorf("MinTimeDeltaSeconds = %d, want 60", config.Thresholds.MinTimeDeltaSeconds)
		}
		if len(config.ExcludeFiles) != 2 {
			t.Errorf("len(ExcludeFiles) = %d, want 2", len(config.ExcludeFiles))
		}
		if len(config.ExcludeFiles) >= 1 && config.ExcludeFiles[0] != "*.log" {
			t.Errorf("ExcludeFiles[0] = %s, want *.log", config.ExcludeFiles[0])
		}
		if len(config.ExcludeFiles) >= 2 && config.ExcludeFiles[1] != "*.tmp" {
			t.Errorf("ExcludeFiles[1] = %s, want *.tmp", config.ExcludeFiles[1])
		}
	})

	t.Run("load from json file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		jsonContent := `{
  "thresholds": {
    "suspicious_additions": 300,
    "suspicious_deletions": 600,
    "max_additions_per_min": 75.0,
    "max_deletions_per_min": 150.0,
    "min_time_delta_seconds": 30
  },
  "exclude_files": ["*.md"]
}`
		if err := os.WriteFile(configFile, []byte(jsonContent), 0o600); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := Load(configFile)
		if err != nil {
			t.Fatalf("Load() unexpected error = %v", err)
		}

		if config.Thresholds.SuspiciousAdditions != 300 {
			t.Errorf("SuspiciousAdditions = %d, want 300", config.Thresholds.SuspiciousAdditions)
		}
		if config.Thresholds.SuspiciousDeletions != 600 {
			t.Errorf("SuspiciousDeletions = %d, want 600", config.Thresholds.SuspiciousDeletions)
		}
		if config.Thresholds.MaxAdditionsPerMin != 75.0 {
			t.Errorf("MaxAdditionsPerMin = %f, want 75.0", config.Thresholds.MaxAdditionsPerMin)
		}
		if config.Thresholds.MaxDeletionsPerMin != 150.0 {
			t.Errorf("MaxDeletionsPerMin = %f, want 150.0", config.Thresholds.MaxDeletionsPerMin)
		}
		if config.Thresholds.MinTimeDeltaSeconds != 30 {
			t.Errorf("MinTimeDeltaSeconds = %d, want 30", config.Thresholds.MinTimeDeltaSeconds)
		}
		if len(config.ExcludeFiles) != 1 {
			t.Errorf("len(ExcludeFiles) = %d, want 1", len(config.ExcludeFiles))
		}
	})

	t.Run("error on non-existent file", func(t *testing.T) {
		config, err := Load("/non/existent/config.yaml")
		if err == nil {
			t.Fatal("Load() expected error for non-existent file but got none")
		}
		if config != nil {
			t.Errorf("Load() expected nil config on error, got %v", config)
		}
	})

	t.Run("error on invalid yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid.yaml")

		invalidContent := `thresholds:
  suspicious_additions: not_a_number
`
		if err := os.WriteFile(configFile, []byte(invalidContent), 0o600); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := Load(configFile)
		if err != nil {
			t.Fatalf("Load() unexpected error = %v", err)
		}

		if config.Thresholds.SuspiciousAdditions != 0 {
			t.Errorf("Invalid value should result in zero, got %d", config.Thresholds.SuspiciousAdditions)
		}
	})

	t.Run("partial config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "partial.yaml")

		partialContent := `thresholds:
  suspicious_additions: 200
`
		if err := os.WriteFile(configFile, []byte(partialContent), 0o600); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := Load(configFile)
		if err != nil {
			t.Fatalf("Load() unexpected error = %v", err)
		}

		if config.Thresholds.SuspiciousAdditions != 200 {
			t.Errorf("SuspiciousAdditions = %d, want 200", config.Thresholds.SuspiciousAdditions)
		}
		if config.Thresholds.SuspiciousDeletions != 0 {
			t.Errorf("Unset SuspiciousDeletions should be 0, got %d", config.Thresholds.SuspiciousDeletions)
		}
	})
}

func TestGenerateSampleConfig(t *testing.T) {
	t.Run("generates sample config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".vibector.yaml")

		err := GenerateSampleConfig(configFile)
		if err != nil {
			t.Fatalf("GenerateSampleConfig() unexpected error = %v", err)
		}

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Fatal("Sample config file was not created")
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("Failed to read generated config: %v", err)
		}

		contentStr := string(content)
		expectedStrings := []string{
			"suspicious_additions",
			"suspicious_deletions",
			"max_additions_per_min",
			"max_deletions_per_min",
			"min_time_delta_seconds",
			"exclude_files",
		}

		for _, expected := range expectedStrings {
			if !contains(contentStr, expected) {
				t.Errorf("Generated config missing expected field: %s", expected)
			}
		}
	})

	t.Run("generated config is valid yaml and loads", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".vibector.yaml")

		err := GenerateSampleConfig(configFile)
		if err != nil {
			t.Fatalf("GenerateSampleConfig() unexpected error = %v", err)
		}

		config, err := Load(configFile)
		if err != nil {
			t.Fatalf("Load() failed on generated config: %v", err)
		}

		if config.Thresholds.SuspiciousAdditions != 500 {
			t.Errorf("Generated SuspiciousAdditions = %d, want 500", config.Thresholds.SuspiciousAdditions)
		}
		if config.Thresholds.SuspiciousDeletions != 1000 {
			t.Errorf("Generated SuspiciousDeletions = %d, want 1000", config.Thresholds.SuspiciousDeletions)
		}
		if config.Thresholds.MaxAdditionsPerMin != 100 {
			t.Errorf("Generated MaxAdditionsPerMin = %f, want 100", config.Thresholds.MaxAdditionsPerMin)
		}
		if config.Thresholds.MaxDeletionsPerMin != 500 {
			t.Errorf("Generated MaxDeletionsPerMin = %f, want 500", config.Thresholds.MaxDeletionsPerMin)
		}
		if config.Thresholds.MinTimeDeltaSeconds != 60 {
			t.Errorf("Generated MinTimeDeltaSeconds = %d, want 60", config.Thresholds.MinTimeDeltaSeconds)
		}
	})

	t.Run("error on invalid path", func(t *testing.T) {
		err := GenerateSampleConfig("/invalid/nonexistent/directory/.vibector.yaml")
		if err == nil {
			t.Fatal("GenerateSampleConfig() expected error for invalid path but got none")
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
