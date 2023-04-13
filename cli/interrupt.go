//go:build !windows

package cli

import (
	"os"
	"syscall"
)

func sendInterrupt() {
	p, _ := os.FindProcess(syscall.Getpid())
	err := p.Signal(os.Interrupt)
	if err != nil {
		os.Exit(1)
	}
}
