package timeout

import (
	"os/exec"
	"strconv"
)

func (tio *Timeout) getCmd() *exec.Cmd {
	return tio.Cmd
}

func (tio *Timeout) kill() error {
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(tio.Cmd.Process.Pid)).Run()
}
