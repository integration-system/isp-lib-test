// +build linux

package docker

import (
	"os"
	"syscall"
)

var signals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT}
