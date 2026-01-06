package reporter

import (
	"fmt"

	"github.com/anisimov-anthony/vibector/internal/detector"
	"github.com/anisimov-anthony/vibector/internal/metrics"
)

type ReportData struct {
	Suspicious []*detector.SuspiciousCommit
	Stats      *metrics.RepositoryStats
	Thresholds *detector.Thresholds
}

type Reporter interface {
	Generate(data *ReportData) (string, error)
}

func NewReporter(format string) (Reporter, error) {
	switch format {
	case "text":
		return &TextReporter{}, nil
	case "json":
		return &JSONReporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
}
