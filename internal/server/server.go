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

// Options for server
type Options struct {
	DeviceID string
	Host     string
	DataPath string
	Codec    string
	VideoExt string
	Treshold float32
	MinSize  float64
}

// Start creates echo server
func Start(opt Options) {
	out := mjpeg.NewStream()

	store, err := data.OpenStore(path.Join(opt.DataPath, "data.bolt"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	cam, err := gocv.OpenVideoCapture(opt.DeviceID)
	if err != nil {
		panic(fmt.Errorf("Error opening video capture device: %v", opt.DeviceID))
	}
	defer cam.Close()
	fmt.Printf("Camera mode: %v, format: %v", cam.Get(gocv.VideoCaptureMode), cam.Get(gocv.VideoCaptureFormat))

	t := &Template{
		templates: template.Must(template.ParseGlob("../../public/views/*.html")),
	}

	go captureLoop(cam, out, store, opt)

	e := echo.New()
	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		m, err := store.GetAllMotions()
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read motions")
		}
		reverse(m)
		return c.Render(http.StatusOK, "index", m)
	})

	e.GET("/item/:id", func(c echo.Context) error {
		i := c.Param("id")
		id, err := strconv.ParseUint(i, 10, 0)
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read id")
		}

		m, err := store.GetMotion(id)
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read motion")
		}

		return c.Render(http.StatusOK, "item", m)
	})

	e.GET("/live", func(c echo.Context) error {
		out.ServeHTTP(c.Response().Writer, c.Request())
		return c.String(http.StatusOK, "")
	})

	e.GET("/api/motions/:id/video", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 0)
		if err != nil {
			return c.String(http.StatusNotFound, "wrong id")
		}
		i, err := store.GetMotion(id)
		return c.File(i.VideoFile)
	})

	e.GET("/api/motions/:id/preview", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 0)
		if err != nil {
			return c.String(http.StatusNotFound, "wrong id")
		}
		i, err := store.GetMotion(id)

		return c.File(i.PreviewFile)
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
	e.Logger.Fatal(e.Start(opt.Host))
}

func captureLoop(cam *gocv.VideoCapture, out *mjpeg.Stream, store *data.Store, options Options) {
	delayedState := motion.NewDelayedState(400*time.Millisecond, 1*time.Second)
	det := motion.NewDetector(options.MinSize, options.Treshold)
	defer det.Close()
	tl := video.NewTimeline(options.DataPath, options.Codec, options.VideoExt)
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
				item, err := tl.CloseItem()
				if err == nil {
					item := data.Motion{
						VideoFile:   item.VideoFile,
						PreviewFile: item.PreviewFile,
						Start:       item.Start,
						Duration:    item.Duration,
					}
					store.UpdateMotion(&item)
					// TODO: write bounding rects to parquet file
					// and stores it in different bucket with the same ID
				}
				status = ""
			}
		}

		// embed date time to the video
		l := time.Now().Format("2006-01-02 15:04:05")
		gocv.PutText(det.Img(), l, image.Pt(840, 25), gocv.FontHersheyPlain, 2.2, green, 1)

		if on {
			tl.WriteImage(det.Img())
		}

		// draw bounding rectangels
		for _, c := range contours {
			rect := gocv.BoundingRect(c)
			if on {
				tl.WriteRect(rect)
			}
			gocv.Rectangle(det.Img(), rect, blue, 2)
		}

		gocv.PutText(det.Img(), status, image.Pt(10, 25), gocv.FontHersheyPlain, 2.2, red, 2)

		// update web stream
		buf, _ := gocv.IMEncode(".jpg", *det.Img())
		out.UpdateJPEG(buf)
	}
}

func reverse(a []data.Motion) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}
