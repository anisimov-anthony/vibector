package git

import "time"

type Commit struct {
	Hash      string
	Author    string
	Email     string
	Timestamp time.Time
	Message   string
	Parents   []string
}

type CommitPair struct {
	Previous  *Commit
	Current   *Commit
	TimeDelta time.Duration
	Stats     *DiffStats
}

type DiffStats struct {
	FilesChanged      int
	Additions         int64
	Deletions         int64
	TotalAdditions    int64
	TotalDeletions    int64
	FilesChangedTotal int
}

type CommitOptions struct {
	Branch   string
	MaxDepth int
}
