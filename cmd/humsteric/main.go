package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pvasek/hamsterproject/internal/server"
)

func main() {

	host := flag.String("host", "127.0.0.1:9001", "host with ports, default 127.0.0.1:9001")
	deviceID := flag.String("deviceId", "0", "id of cam device")
	dataPath := flag.String("dataPath", "", "the absolute path where data are stored")
	treshold := flag.Float64("trashold", 30, "Trashold that will be skiped 0-255")
	minSize := flag.Float64("minarea", 2000, "Min area that should be taken into account")
	codec := flag.String("codec", "H264", "Codec that should be used")
	videoExt := flag.String("videoExt", ".mp4", "Video extension that should be used")

	flag.Parse()

	if *dataPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Host: %v, Data path: %v\n", *host, *dataPath)

	options := server.Options{
		DataPath: *dataPath,
		DeviceID: *deviceID,
		Host:     *host,
		Treshold: float32(*treshold),
		MinSize:  *minSize,
		Codec:    *codec,
		VideoExt: *videoExt,
	}
	server.Start(options)

}
