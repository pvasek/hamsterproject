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

	flag.Parse()

	if *dataPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Host: %v, Data path: %v\n", *host, *dataPath)

	server.Start(*deviceID, *dataPath, *host)

}
