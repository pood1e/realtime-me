package domain

import "time"

// ContentObject is one immutable, deduplicated source file.
type ContentObject struct {
	UID         string
	SHA256      []byte
	SizeBytes   int64
	ContentType string
	StorageKey  string
	CreateTime  time.Time
}

// SealedContent is a newly published upload ready to be claimed transactionally.
type SealedContent struct {
	SHA256      []byte
	SizeBytes   int64
	ContentType string
	StorageKey  string
}

// ProcessingStatus is the shared readiness state for derived metadata.
type ProcessingStatus string

const (
	// ProcessingStatusPending identifies queued or running work.
	ProcessingStatusPending ProcessingStatus = "pending"
	// ProcessingStatusReady identifies usable derived metadata.
	ProcessingStatusReady ProcessingStatus = "ready"
	// ProcessingStatusFailed identifies work requiring an explicit retry.
	ProcessingStatusFailed ProcessingStatus = "failed"
)

// ProcessingJob is one leased background processing request.
type ProcessingJob struct {
	UID         string
	Kind        string
	ResourceUID string
	Attempts    int
}

// WorkerHealth summarizes background processing readiness.
type WorkerHealth struct {
	HeartbeatTime *time.Time
	PendingJobs   int64
}

// Artifact is a regenerable file derived from a content object.
type Artifact struct {
	UID         string
	ContentUID  string
	Kind        string
	Variant     string
	ContentType string
	StorageKey  string
	Width       int
	Height      int
	CreateTime  time.Time
}
