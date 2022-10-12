package main

import (
	"log"

	"github.com/cobalt-robotics/gortsplib"
	"github.com/cobalt-robotics/gortsplib/pkg/url"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. get and print informations about tracks published on a path.

func main() {
	c := gortsplib.Client{}

	u, err := url.Parse("rtsp://localhost:8554/mypath")
	if err != nil {
		panic(err)
	}

	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	tracks, _, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	log.Printf("available tracks: %v\n", tracks)
}
