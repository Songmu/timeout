// +build !windows

package timeout

import (
	"os/exec"
	"syscall"
)

func (tio *Timeout) getCmd() *exec.Cmd {
	if tio.Cmd.SysProcAttr == nil {
		tio.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	return tio.Cmd
}

func (tio *Timeout) terminate() error {
	syssig, ok := tio.signal().(syscall.Signal)
	if !ok || tio.Foreground {
		return tio.Cmd.Process.Signal(tio.signal())
	}
	return syscall.Kill(-tio.Cmd.Process.Pid, syssig)
}

func (tio *Timeout) killall() error {
	return syscall.Kill(-tio.Cmd.Process.Pid, syscall.SIGKILL)
}
