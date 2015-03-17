// +build !windows

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func parseSignal(sigStr string) (os.Signal, error) {
	switch strings.ToUpper(sigStr) {
	case "":
		return nil, nil
	case "HUP", "1":
		return syscall.SIGHUP, nil
	case "INT", "2":
		return os.Interrupt, nil
	case "QUIT", "3":
		return syscall.SIGQUIT, nil
	case "KILL", "9":
		return os.Kill, nil
	case "ALRM", "14":
		return syscall.SIGALRM, nil
	case "TERM", "15":
		return syscall.SIGTERM, nil
	case "USR1":
		return syscall.SIGUSR1, nil
	case "USR2":
		return syscall.SIGUSR2, nil
	default:
		return nil, fmt.Errorf("%s: invalid signal", sigStr)
	}
}
