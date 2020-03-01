package data

import "time"

// Motion represents one motion that was detected
type Motion struct {
	ID              uint64
	StartTime       time.Time
	Duration        time.Duration
	FileName        string
	PreviewFileName string
}
