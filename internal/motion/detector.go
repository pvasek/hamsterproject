package motion

import (
	"image"

	"gocv.io/x/gocv"
)

// Detector for motion
type Detector struct {
	img     gocv.Mat
	imgth   gocv.Mat
	imgdt   gocv.Mat
	mog2    gocv.BackgroundSubtractorMOG2
	minArea float64
}

// NewDetector creates a new motion detector
func NewDetector(minArea float64) Detector {
	return Detector{
		img:     gocv.NewMat(),
		imgdt:   gocv.NewMat(),
		imgth:   gocv.NewMat(),
		mog2:    gocv.NewBackgroundSubtractorMOG2(),
		minArea: minArea,
	}
}

// Close the detector and free up all resources
func (d *Detector) Close() {
	d.img.Close()
	d.imgth.Close()
	d.imgdt.Close()
	d.mog2.Close()
}

// IsEmpty if detection image is empty
func (d *Detector) IsEmpty() bool {
	return d.img.Empty()
}

// Detect motion in the image
func (d *Detector) Detect() (bool, [][]image.Point) {
	// first phase of cleaning up image, obtain foreground only
	d.mog2.Apply(d.img, &d.imgdt)

	// remaining cleanup of the image to use for finding contours.
	// first use threshold
	gocv.Threshold(d.imgdt, &d.imgth, 25, 255, gocv.ThresholdBinary)

	// then dilate
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(d.imgdt, &d.imgdt, kernel)

	// now find contours
	contours := gocv.FindContours(d.imgdt, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	m := false
	cl := [][]image.Point{}
	for _, c := range contours {
		area := gocv.ContourArea(c)
		if area < d.minArea {
			continue
		}
		m = true
		cl = append(cl, c)
		break
	}
	return m, cl
}
