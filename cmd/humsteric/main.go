package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pvasek/hamsterproject/internal/motion"
	"github.com/pvasek/hamsterproject/internal/video"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

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

	delayedState := motion.NewDelayedState(400*time.Millisecond, 1*time.Second)
	det := motion.NewDetector(3000)
	defer det.Close()
	tl := video.NewTimeline("MJPG", "/Users/pvasek/Documents/test_videos/video_%v.avi")
	defer tl.Close()

	var statusColor color.RGBA
	status := "-"
	fmt.Printf("Start reading device: %v\n", deviceID)

	for {
		if ok := webcam.Read(det.Img()); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if det.IsEmpty() {
			continue
		}

		// now find contours
		inmotion, contours := det.Detect()

		changed, on := delayedState.Detect(inmotion, time.Now())
		if changed {
			if on {
				tl.NewItem()
				status = "REC"
				statusColor = color.RGBA{255, 0, 0, 0}
			} else {
				tl.CloseItem()
				status = "-"
				statusColor = color.RGBA{0, 255, 0, 0}
			}
		}

		// embed date time to the video
		gocv.PutText(
			det.Img(), time.Now().Format("2006-01-02 15:04:05"), image.Pt(840, 25), gocv.FontHersheyPlain, 2.2, color.RGBA{0, 255, 0, 0}, 1)

		if on {
			tl.Write(det.Img())
		}

		// draw bounding rectangels
		for _, c := range contours {
			rect := gocv.BoundingRect(c)
			gocv.Rectangle(det.Img(), rect, color.RGBA{0, 0, 255, 0}, 2)
		}

		gocv.PutText(det.Img(), status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		// update web stream
		buf, _ := gocv.IMEncode(".jpg", *det.Img())
		stream.UpdateJPEG(buf)
	}
}
