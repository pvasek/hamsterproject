package motion

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

	delayedState := NewDelayedState(1*time.Second, 5*time.Second)
	det := NewDetector(3000)
	defer det.Close()

	green := color.RGBA{0, 255, 0, 0}
	status := "Ready"
	counter := 0

	var writer *gocv.VideoWriter

	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&det.img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if det.IsEmpty() {
			continue
		}

		status = "Ready"
		statusColor := color.RGBA{0, 255, 0, 0}

		// now find contours
		inmotion, contours := det.Detect()

		if delayedState.detect(inmotion, time.Now()) {
			l := time.Now().Format("2006-01-02--15-04-05")
			fileName := fmt.Sprintf("/Users/pvasek/Documents/test_videos/video_%v.avi", l)
			fmt.Printf("Writing new video to %v\n", fileName)
			writer, err = gocv.VideoWriterFile(fileName, "MJPG", 10, det.img.Cols(), det.img.Rows(), true)
			if err != nil {
				panic(err)
			}
			defer writer.Close()
			counter++
		} else {
			writer.Close()
		}

		// embed date time to the video
		gocv.PutText(&det.img, time.Now().Format("2006-01-02 15:04:05"), image.Pt(840, 25), gocv.FontHersheyPlain, 2.2, green, 1)

		if inmotion {
			writer.Write(det.img)
		}

		// draw bounding rectangels
		for _, c := range contours {
			status = "Motion detected"
			statusColor = color.RGBA{255, 0, 0, 0}
			rect := gocv.BoundingRect(c)
			gocv.Rectangle(&det.img, rect, color.RGBA{0, 0, 255, 0}, 2)
		}

		gocv.PutText(&det.img, fmt.Sprintf("%v %v %v", status, inmotion, counter), image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		// update web stream
		buf, _ := gocv.IMEncode(".jpg", det.img)
		stream.UpdateJPEG(buf)
	}
}
