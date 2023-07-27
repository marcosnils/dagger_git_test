package main

import (
	"fmt"
	"log"
	"runtime/debug"
)

func main() {
	// get VCS informaiton from BuildInfo
	info, ok := debug.ReadBuildInfo()

	if !ok {
		log.Fatalf("Could not get build info")
	}
	for _, setting := range info.Settings {
		fmt.Printf("%s: %s\n", setting.Key, setting.Value)
	}
}
