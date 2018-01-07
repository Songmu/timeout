// +build !windows

package timeout

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestRunSimple_withStop(t *testing.T) {
	tio := &Timeout{
		Duration:  2 * time.Second,
		KillAfter: 1 * time.Second,
		Cmd:       exec.Command(shellcmd, shellflag, "sleep 10"),
	}
	ch, err := tio.RunCommand()
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	tio.Cmd.Process.Signal(syscall.SIGSTOP)
	st := <-ch

	expect := 128 + 15
	if st.Code != expect {
		t.Errorf("exit code invalid. out: %d, expect: %d", st.Code, expect)
	}
}
