package video

import (
	"fmt"
	"image"
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

	currentItem *gocv.VideoWriter
	item        *Item
}

// Item in timeline
type Item struct {
	Start           time.Time
	VideoFile       string
	PreviewFile     string
	Duration        time.Duration
	VideoFileSize   int
	PreviewFileSize int
	BoundingRects   []image.Rectangle
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
	l := time.Now().Format("2006-01-02--15-04-05")
	base := path.Join(t.dataPath, l)

	t.item = &Item{
		Start:           time.Now(),
		VideoFile:       base + t.videoExt,
		PreviewFile:     base + ".jpg",
		VideoFileSize:   0,
		PreviewFileSize: 0,
		BoundingRects:   []image.Rectangle{},
	}
}

// CloseItem closes the current item
func (t *Timeline) CloseItem() (Item, error) {
	if t.currentItem != nil {
		t.currentItem.Close()
		t.currentItem = nil
		t.item.Duration = time.Now().Sub(t.item.Start)
		return *t.item, nil
	}
	return Item{}, fmt.Errorf("No item to close")
}

// WriteImage image to the current item
func (t *Timeline) WriteImage(img *gocv.Mat) error {
	var err error
	if t.currentItem == nil {
		t.currentItem, err = gocv.VideoWriterFile(
			t.item.VideoFile, t.codec, 10, img.Cols(), img.Rows(), true)

		if err == nil {
			// generate preview
			// preview := gocv.Mat{}
			// gocv.Resize(*img, &preview, image.Point{}, 0.3, 0.3, gocv.InterpolationDefault)
			preview := *img
			buf, e := gocv.IMEncode(".jpg", preview)
			if e == nil {
				ioutil.WriteFile(t.item.PreviewFile, buf, 0644)
				t.item.PreviewFileSize = len(buf)
			}
		}
	}

	if err == nil {
		t.item.VideoFileSize += img.Rows() * img.Cols()
		return t.currentItem.Write(*img)
	}

	return err
}

// WriteRect write bounding rectangle
func (t *Timeline) WriteRect(rect image.Rectangle) {
	t.item.BoundingRects = append(t.item.BoundingRects, rect)
}
