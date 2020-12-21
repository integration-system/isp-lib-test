// +build darwin windows

package docker

import (
	"os"
)

var signals = []os.Signal{os.Interrupt, os.Kill}
