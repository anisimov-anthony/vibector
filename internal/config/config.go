package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/anisimov-anthony/vibector/internal/detector"
)

type Config struct {
	Thresholds   detector.Thresholds
	ExcludeFiles []string
}

func Load(configFile string) (*Config, error) {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvPrefix("VIBECTOR")
	v.AutomaticEnv()

	config := &Config{}

	config.Thresholds.SuspiciousAdditions = v.GetInt64("thresholds.suspicious_additions")
	config.Thresholds.SuspiciousDeletions = v.GetInt64("thresholds.suspicious_deletions")
	config.Thresholds.MaxAdditionsPerMin = v.GetFloat64("thresholds.max_additions_per_min")
	config.Thresholds.MaxDeletionsPerMin = v.GetFloat64("thresholds.max_deletions_per_min")
	config.Thresholds.MinTimeDeltaSeconds = v.GetInt64("thresholds.min_time_delta_seconds")

	config.ExcludeFiles = v.GetStringSlice("exclude_files")

	return config, nil
}

func GenerateSampleConfig(path string) error {
	sample := `# Vibector Configuration File
# Detect potential AI-generated code by analyzing commit patterns

thresholds:
  # Commit size thresholds - flag commits with large changes
  # AI often generates many lines of code at once
  suspicious_additions: 500    # Flag commits with more than 500 added lines (0 to disable)
  suspicious_deletions: 1000   # Flag commits with more than 1000 deleted lines (0 to disable)

  # Velocity thresholds - flag commits with high coding speed
  # Typical human rates: 20-50 LOC/min additions, 50-200 LOC/min deletions
  # AI-assisted rates: often 100+ LOC/min additions, 500+ LOC/min deletions
  max_additions_per_min: 100   # Flag if adding code faster than 100 lines/min (0 to disable)
  max_deletions_per_min: 500   # Flag if deleting code faster than 500 lines/min (0 to disable)

  # Time threshold - flag commits that are too close together
  min_time_delta_seconds: 60   # Flag commits less than 60 seconds apart (0 to disable)

# File patterns to exclude from diff statistics (e.g., ["*.log", "*.tmp", "package-lock.json"])
exclude_files: []
`

	return os.WriteFile(path, []byte(sample), 0o600)
}
