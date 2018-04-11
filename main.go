package main

import (
	"os"

	"github.com/cv/sd/cli"
)

var (
	// version is set by goreleaser
	version string
)

func main() {
	sd := cli.New(version)
	err := sd.Run()
	if err != nil {
		os.Exit(-1)
	}
}
