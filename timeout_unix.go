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
	return cmd.Process.Signal(tio.signal())
}

func (tio *Timeout) killall() error {
	return syscall.Kill(-tio.Cmd.Process.Pid, syscall.SIGKILL)
}
