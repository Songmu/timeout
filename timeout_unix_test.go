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
		Duration:  100 * time.Microsecond,
		KillAfter: 1 * time.Second,
		Cmd:       exec.Command(shellcmd, shellflag, "sleep 10"),
	}
	ch, err := tio.RunCommand()
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	tio.Cmd.Process.Signal(syscall.SIGSTOP)
	st := <-ch

	expect := 128 + int(syscall.SIGTERM)
	if st.Code != expect {
		t.Errorf("exit code invalid. out: %d, expect: %d", st.Code, expect)
	}
}

func TestRunCommand_signaled(t *testing.T) {
	testCases := []struct {
		name     string
		cmd      *exec.Cmd
		exit     int
		signaled bool
	}{
		{
			name:     "signal handled",
			cmd:      exec.Command(stubCmd, "-trap", "SIGTERM", "-trap-exit", "23", "-sleep", "3"),
			exit:     23,
			signaled: false,
		},
		{
			name:     "termed by sigterm",
			cmd:      exec.Command("sleep", "1"),
			exit:     128 + int(syscall.SIGTERM),
			signaled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tio := &Timeout{
				Duration:  100 * time.Millisecond,
				KillAfter: 3 * time.Second,
				Cmd:       tc.cmd,
			}
			st, _, _, err := tio.Run()

			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}

			if st.GetChildExitCode() != tc.exit {
				t.Errorf("expected exitcode: %d, but: %d", tc.exit, st.GetChildExitCode())
			}
			if st.Signaled != tc.signaled {
				t.Errorf("something went wrong")
			}
		})
	}
}
