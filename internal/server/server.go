package server

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/pvasek/hamsterproject/internal/data"
	"github.com/pvasek/hamsterproject/internal/motion"
	"github.com/pvasek/hamsterproject/internal/video"

	"github.com/hybridgroup/mjpeg"
	"github.com/labstack/echo/v4"
	"gocv.io/x/gocv"
)

// Start creates echo server
func Start(deviceID string, dataPath string, host string) {
	out := mjpeg.NewStream()

	store, err := data.OpenStore(path.Join(dataPath, "data.bolt"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	cam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		panic(fmt.Errorf("Error opening video capture device: %v", deviceID))
	}
	defer cam.Close()

	t := &Template{
		templates: template.Must(template.ParseGlob("../../public/views/*.html")),
	}

	go captureLoop(cam, out, store, dataPath)

	e := echo.New()
	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		// return c.String(http.StatusOK, "Hello, World!")
		return c.Render(http.StatusOK, "index", "World")
	})

	e.GET("/live", func(c echo.Context) error {
		out.ServeHTTP(c.Response().Writer, c.Request())
		return c.String(http.StatusOK, "")
	})

	e.GET("/api/video/:id", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 0)
		if err != nil {
			return c.String(http.StatusNotFound, "wrong id")
		}
		i, err := store.GetMotion(id)
		return c.File(i.FileName)
	})

	e.GET("/api/motions", func(c echo.Context) error {
		m, err := store.GetAllMotions()
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read motions")
		}
		return c.JSON(200, &m)
	})

	if err != nil {
		panic(err)
	}
	e.Logger.Fatal(e.Start(host))
}

func captureLoop(cam *gocv.VideoCapture, out *mjpeg.Stream, store *data.Store, dataPath string) {
	delayedState := motion.NewDelayedState(400*time.Millisecond, 1*time.Second)
	det := motion.NewDetector(3000)
	defer det.Close()
	tl := video.NewTimeline("MJPG", path.Join(dataPath, "video_%v.avi"))
	defer tl.Close()

	red := color.RGBA{255, 0, 0, 0}
	green := color.RGBA{0, 255, 0, 0}
	blue := color.RGBA{0, 0, 255, 0}
	status := ""

	for {
		if ok := cam.Read(det.Img()); !ok {
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
			} else {
				start, dur, name := tl.CloseItem()
				item := data.Motion{
					FileName:  name,
					StartTime: start,
					Duration:  dur,
				}
				store.UpdateMotion(&item)
				status = ""
			}
		}

		// embed date time to the video
		l := time.Now().Format("2006-01-02 15:04:05")
		gocv.PutText(det.Img(), l, image.Pt(840, 25), gocv.FontHersheyPlain, 2.2, green, 1)

		if on {
			tl.Write(det.Img())
		}

		// draw bounding rectangels
		for _, c := range contours {
			rect := gocv.BoundingRect(c)
			gocv.Rectangle(det.Img(), rect, blue, 2)
		}

		gocv.PutText(det.Img(), status, image.Pt(10, 25), gocv.FontHersheyPlain, 2.2, red, 2)

		// update web stream
		buf, _ := gocv.IMEncode(".jpg", *det.Img())
		out.UpdateJPEG(buf)
	}
}
