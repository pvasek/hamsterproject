package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

const MinimumArea = 3000

func main() {

	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tmotion-detect [camera ID]")
		return
	}

	// parse args
	deviceID := os.Args[1]
	host := os.Args[2]

	stream := mjpeg.NewStream()
	go mainLoop(stream, deviceID)

	fmt.Printf("Host: %v\n", host)

	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe(host, nil))
}

func mainLoop(stream *mjpeg.Stream, deviceID string) {

	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// window := gocv.NewWindow("Motion Window")
	// defer window.Close()
	// dwindow := gocv.NewWindow("Delta")
	// defer dwindow.Close()

	green := color.RGBA{0, 255, 0, 0}
	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	status := "Ready"

	counter := 0
	motionInProgres := false
	var (
		lastMotion time.Time
		lastStill  time.Time
		writer     *gocv.VideoWriter
	)

	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		status = "Ready"
		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		defer kernel.Close()
		gocv.Dilate(imgThresh, &imgThresh, kernel)

		// now find contours
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		hasMotion := false
		for _, c := range contours {
			area := gocv.ContourArea(c)
			if area < MinimumArea {
				continue
			}
			hasMotion = true
			break
		}

		if hasMotion {
			lastMotion = time.Now()
		} else if motionInProgres {
			lastStill = time.Now()
		}

		diff := lastMotion.Sub(lastStill)
		if diff.Seconds() > 1 && motionInProgres == false {
			motionInProgres = true
			fileName := fmt.Sprintf("/Users/pvasek/Documents/test_videos/video_%v.avi", counter)
			fmt.Printf("Writing new video to %v\n", fileName)
			writer, err = gocv.VideoWriterFile(fileName, "MJPG", 10, img.Cols(), img.Rows(), true)
			if err != nil {
				panic(err)
			}
			defer writer.Close()
			counter++
		} else if diff.Seconds() < -5 && motionInProgres == true {
			motionInProgres = false
			writer.Close()
		}

		// Mon Jan 2 15:04:05 MST 2006
		gocv.PutText(&img, time.Now().Format("2006-01-02 15:04:05"), image.Pt(840, 25), gocv.FontHersheyPlain, 2.2, green, 1)
		if motionInProgres {
			writer.Write(img)
		}

		for _, c := range contours {
			area := gocv.ContourArea(c)
			if area < MinimumArea {
				continue
			}
			status = "Motion detected"
			statusColor = color.RGBA{255, 0, 0, 0}
			//gocv.DrawContours(&img, contours, i, statusColor, 2)

			rect := gocv.BoundingRect(c)
			gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
		}

		gocv.PutText(&img, fmt.Sprintf("%v %v %v", status, motionInProgres, counter), image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		// window.IMShow(img)
		// dwindow.IMShow(imgDelta)

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf)
	}
}
