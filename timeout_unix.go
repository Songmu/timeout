// +build !windows

package timeout

import (
	"os/exec"
	"syscall"
)

func (tio *Timeout) getCmd() *exec.Cmd {
	cmd := tio.Cmd
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	return cmd
}

func (tio *Timeout) killall() error {
	return syscall.Kill(-tio.Cmd.Process.Pid, syscall.SIGKILL)
}
