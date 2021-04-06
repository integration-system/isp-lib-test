// +build !linux

package docker

import (
	"os"
)

var signals = []os.Signal{os.Interrupt, os.Kill}
