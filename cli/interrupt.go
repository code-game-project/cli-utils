//go:build !windows

package cli

import (
	"os"
	"syscall"
	"time"
)

func sendInterrupt() {
	p, _ := os.FindProcess(syscall.Getpid())
	err := p.Signal(os.Interrupt)
	if err != nil {
		os.Exit(2)
	}
	time.Sleep(5 * time.Second)
	os.Exit(2)
}
