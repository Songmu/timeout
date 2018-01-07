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
	sig := tio.signal()
	syssig, ok := sig.(syscall.Signal)
	if !ok || tio.Foreground {
		return tio.Cmd.Process.Signal(sig)
	}
	err := syscall.Kill(-tio.Cmd.Process.Pid, syssig)
	if err != nil {
		return err
	}
	if syssig != syscall.SIGKILL && syssig != syscall.SIGCONT {
		return syscall.Kill(-tio.Cmd.Process.Pid, syscall.SIGCONT)
	}
	return nil
}

func (tio *Timeout) killall() error {
	return syscall.Kill(-tio.Cmd.Process.Pid, syscall.SIGKILL)
}
