package data

import "time"

// Motion represents one motion that was detected
type Motion struct {
	ID              uint64
	Start           time.Time
	Duration        time.Duration
	VideoFile       string
	VideoFileSize   int
	PreviewFile     string
	PreviewFileSize int
}
