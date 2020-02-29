package video

import (
	"fmt"
	"time"

	"gocv.io/x/gocv"
)

// Timeline of videos
type Timeline struct {
	currentItem  *gocv.VideoWriter
	itemStart    time.Time
	codec        string
	fileTemplate string
}

// NewTimeline create a new Timeline
func NewTimeline(codec string, fileTemplate string) *Timeline {
	return &Timeline{
		codec:        codec,
		fileTemplate: fileTemplate,
	}
}

// Close whole timeline
func (t *Timeline) Close() {
	t.CloseItem()
}

// NewItem creates a new item in timeline
func (t *Timeline) NewItem() {
	t.CloseItem()
	t.itemStart = time.Now()
}

// CloseItem closes the current item
func (t *Timeline) CloseItem() {
	if t.currentItem != nil {
		t.currentItem.Close()
		t.currentItem = nil
	}
}

// Write image to the current item
func (t *Timeline) Write(img *gocv.Mat) error {
	var err error
	if t.currentItem == nil {
		l := time.Now().Format("2006-01-02--15-04-05")
		f := fmt.Sprintf(t.fileTemplate, l)
		t.currentItem, err = gocv.VideoWriterFile(f, t.codec, 10, img.Cols(), img.Rows(), true)
	}

	if err == nil {
		return t.currentItem.Write(*img)
	}

	return err
}
