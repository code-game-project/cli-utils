//go:build windows

package cli

import (
	"os"
	"syscall"
	"time"
)

func sendInterrupt() {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		os.Exit(1)
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		os.Exit(1)
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(syscall.Getpid()))
	if r == 0 {
		os.Exit(1)
	}
	time.Sleep(5 * time.Second)
}
