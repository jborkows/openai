package pages

import (
	"os"
	"time"
)

var startupVersion string

func init() {
	startupVersion = time.Now().Format("20060102150405")
}

func version() string {
	if os.Getenv("VERSION") == "PRODUCTION" {
		return startupVersion
	} else {
		return time.Now().Format("20060102150405")
	}
}
