# Vibector

[![CI](https://github.com/anisimov-anthony/vibector/actions/workflows/ci.yml/badge.svg)](https://github.com/anisimov-anthony/vibector/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/anisimov-anthony/vibector/graph/badge.svg?token=75TMCZBVP4)](https://codecov.io/gh/anisimov-anthony/vibector)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/anisimov-anthony/vibector)](go.mod)

**Vibector** is a Git repository analysis tool designed to detect potential AI-generated code by examining commit patterns, code velocity, and statistical anomalies.

## Features

- Analyze local git repositories for suspicious commit patterns
- Configurable thresholds for detecting anomalous coding velocity
- Separate thresholds for code additions vs deletions
- Statistical analysis with percentile calculations
- File exclusion support (ignore logs, generated files, etc.)
- Multiple output formats (text and JSON)
- Read-only operations - never modifies your repository
- No external dependencies or data transmission

## Installation

### Build from Source

```bash
git clone https://github.com/anisimov-anthony/vibector.git
cd vibector
go build -o vibector ./cmd/vibector
```

## Quick Start

### 1. Generate Configuration File (Optional)

```bash
vibector config init
```

This creates a `.vibector.yaml` file that you can edit to set default thresholds.

### 2. Analyze a Repository

```bash
# Analyze with command-line flags
vibector analyze /path/to/repo \
  --output report.txt \
  --suspicious-additions 500 \
  --max-additions-pm 100

# Analyze with separate thresholds for additions and deletions
vibector analyze /path/to/repo \
  --output report.json \
  --suspicious-additions 500 \
  --suspicious-deletions 1000 \
  --max-additions-pm 100 \
  --max-deletions-pm 500

# Exclude files from analysis
vibector analyze /path/to/repo \
  --output report.txt \
  --suspicious-additions 500 \
  --exclude-files "*.log,*.tmp,package-lock.json"
```

## Usage

### `vibector analyze <repository>`

Analyze a repository for suspicious commits that may indicate AI-generated code.

**Required Arguments:**
- `<repository>` - Path to local git repository

**Required Flags:**
- `--output, -o <file>` - Output file path. Format detected from extension (`.txt` or `.json`)

**Optional Flags:**
- `--config <file>` - Path to configuration file
- `--suspicious-additions <n>` - Flag commits with more than N additions (0 to disable)
- `--suspicious-deletions <n>` - Flag commits with more than N deletions (0 to disable)
- `--max-additions-pm <n>` - Maximum additions per minute (0 to disable)
- `--max-deletions-pm <n>` - Maximum deletions per minute (0 to disable)
- `--min-time-delta <n>` - Minimum seconds between commits (0 to disable)
- `--branch <name>` - Specific branch to analyze
- `--exclude-files <patterns>` - Comma-separated file patterns to exclude

**Note:** At least one threshold must be configured via flags or config file.

**Examples:**

```bash
# Basic analysis with additions threshold
vibector analyze . \
  --output report.txt \
  --suspicious-additions 500 \
  --max-additions-pm 100

# Full analysis with all thresholds
vibector analyze . \
  --output analysis.json \
  --suspicious-additions 500 \
  --suspicious-deletions 1000 \
  --max-additions-pm 100 \
  --max-deletions-pm 500 \
  --min-time-delta 60

# Analyze specific branch with file exclusions
vibector analyze . \
  --output report.txt \
  --branch main \
  --suspicious-additions 500 \
  --exclude-files "*.log,*.tmp"
```

### `vibector config init`

Generate a sample `.vibector.yaml` configuration file in the current directory.

## Configuration

### Configuration File

Create a `.vibector.yaml` file to persist your settings:

```yaml
thresholds:
  # Flag commits with large changes
  suspicious_additions: 500    # Flag commits with more than 500 added lines (0 to disable)
  suspicious_deletions: 1000   # Flag commits with more than 1000 deleted lines (0 to disable)

  # Flag commits with high coding speed
  # Typical human rates: 20-50 LOC/min additions, 50-200 LOC/min deletions
  # AI-assisted rates: often 100+ LOC/min additions, 500+ LOC/min deletions
  max_additions_per_min: 100   # Flag if adding code faster than 100 lines/min (0 to disable)
  max_deletions_per_min: 500   # Flag if deleting code faster than 500 lines/min (0 to disable)

  # Flag commits that are too close together
  min_time_delta_seconds: 60   # Flag commits less than 60 seconds apart (0 to disable)

# File patterns to exclude from diff statistics
exclude_files: []
```

### Environment Variables

Configure via environment variables:

```bash
export VIBECTOR_THRESHOLDS_SUSPICIOUS_ADDITIONS=500
export VIBECTOR_THRESHOLDS_MAX_ADDITIONS_PER_MIN=100
export VIBECTOR_THRESHOLDS_MAX_DELETIONS_PER_MIN=500

vibector analyze /path/to --output report.txt
```

## Output Formats

Format is automatically detected from the file extension

## How It Works

1. **Commit Retrieval** - Extracts commits from the repository
2. **Pair Creation** - Creates consecutive commit pairs (skips merge commits)
3. **Diff Analysis** - Calculates additions/deletions for each pair
4. **Velocity Calculation** - Computes LOC per minute based on time delta
5. **Statistical Analysis** - Calculates percentiles and repository-wide metrics
6. **Threshold Detection** - Flags commits exceeding configured thresholds
7. **Report Generation** - Outputs results in requested format

### Detection Methods

Vibector flags commits based on:

- **Commit Size** - Absolute number of additions or deletions
- **Velocity Thresholds** - Code written/deleted per minute
- **Time Delta** - Commits made too quickly in succession
- **Statistical Context** - Includes percentile analysis for repository context

## Use Cases

- Code review prioritization
- Repository auditing and compliance
- Identifying commits requiring extra scrutiny
- Research on coding patterns
- Team coding velocity analysis

## Security & Privacy

- **Read-only** - Never modifies your repository
- **Local analysis** - All processing happens on your machine
- **No telemetry** - No data sent to external services
- **No network access** - Works completely offline

## License

MIT License - see [LICENSE](LICENSE) file for details.
