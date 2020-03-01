package video

import (
	"io/ioutil"
	"path"
	"time"

	"gocv.io/x/gocv"
)

// Timeline of videos
type Timeline struct {
	dataPath string
	codec    string
	videoExt string

	currentItem     *gocv.VideoWriter
	itemStart       time.Time
	itemName        string
	itemPreviewName string
}

// NewTimeline create a new Timeline
func NewTimeline(dataPath string, codec string, videoExt string) *Timeline {
	return &Timeline{
		dataPath: dataPath,
		codec:    codec,
		videoExt: videoExt,
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
	l := time.Now().Format("2006-01-02--15-04-05")
	base := path.Join(t.dataPath, l)
	t.itemName = base + t.videoExt
	t.itemPreviewName = base + ".jpg"

}

// CloseItem closes the current item
func (t *Timeline) CloseItem() (time.Time, time.Duration, string, string) {
	if t.currentItem != nil {
		t.currentItem.Close()
		t.currentItem = nil
		return t.itemStart, time.Now().Sub(t.itemStart), t.itemName, t.itemPreviewName
	}
	return time.Time{}, 0, "", ""
}

// Write image to the current item
func (t *Timeline) Write(img *gocv.Mat) error {
	var err error
	if t.currentItem == nil {
		t.currentItem, err = gocv.VideoWriterFile(t.itemName, t.codec, 10, img.Cols(), img.Rows(), true)
		if err == nil {
			// generate preview
			buf, e := gocv.IMEncode(".jpg", *img)
			if e == nil {
				ioutil.WriteFile(t.itemPreviewName, buf, 0644)
			}
		}
	}

	if err == nil {
		return t.currentItem.Write(*img)
	}

	return err
}
