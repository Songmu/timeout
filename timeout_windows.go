package timeout

import (
	"os/exec"
	"strconv"
	"syscall"
)

func (tio *Timeout) getCmd() *exec.Cmd {
	if tio.Cmd.SysProcAttr == nil {
		tio.Cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
		}
	}
	return tio.Cmd
}

func (tio *Timeout) killall() error {
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(tio.Cmd.Process.Pid)).Run()
}
