package timeout

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	shellcmd  = "/bin/sh"
	shellflag = "-c"
)

func init() {
	if runtime.GOOS == "windows" {
		shellcmd = "cmd"
		shellflag = "/c"
	}
}

func TestRun(t *testing.T) {
	tio := &Timeout{
		Duration: 10 * time.Second,
		Cmd:      exec.Command(shellcmd, shellflag, "echo 1"),
	}
	_, stdout, stderr, err := tio.Run()

	if strings.TrimSpace(stdout) != "1" {
		t.Errorf("something wrong")
	}

	if stderr != "" {
		t.Errorf("something wrong")
	}

	if err != nil {
		t.Errorf("something wrong: %v", err)
	}
}

var isWin = runtime.GOOS == "windows"

func TestRunSimple(t *testing.T) {
	testCases := []struct {
		name           string
		duration       time.Duration
		killAfter      time.Duration
		cmd            *exec.Cmd
		signal         os.Signal
		preserveStatus bool
		expectedExit   int
		skipOnWin      bool
	}{
		{
			name:         "simple echo",
			duration:     time.Duration(0.1 * float64(time.Second)),
			cmd:          exec.Command(shellcmd, shellflag, "echo 1"),
			expectedExit: 0,
		},
		{
			name:         "timed out",
			cmd:          exec.Command(shellcmd, shellflag, "sleep 3"),
			duration:     1 * time.Second,
			signal:       os.Interrupt,
			expectedExit: 124,
		},
		{
			name:           "timed out with preserve status",
			cmd:            exec.Command(shellcmd, shellflag, "sleep 3"),
			duration:       time.Duration(0.1 * float64(time.Second)),
			preserveStatus: true,
			expectedExit:   128 + 15,
			skipOnWin:      true,
		},
		{
			name:           "preserve status (signal handled)",
			cmd:            exec.Command("perl", "testdata/exit_with_23.pl"),
			duration:       1 * time.Second,
			preserveStatus: true,
			expectedExit:   23,
			skipOnWin:      true,
		},
		{
			name:         "kill after",
			cmd:          exec.Command("perl", "testdata/ignore_sigterm.pl"),
			duration:     1 * time.Second,
			killAfter:    1 * time.Second,
			signal:       syscall.SIGTERM,
			expectedExit: exitKilled,
		},
		{
			name:           "ignore sigterm but exited before kill after",
			cmd:            exec.Command("perl", "testdata/ignore_sigterm.pl"),
			duration:       1 * time.Second,
			killAfter:      5 * time.Second,
			signal:         syscall.SIGTERM,
			preserveStatus: true,
			expectedExit:   0,
		},
		{
			name:         "command cannnot be invoked",
			cmd:          exec.Command("testdata/dummy"),
			duration:     1 * time.Second,
			expectedExit: 126, // TODO cmd should return 125 on win
			skipOnWin:    true,
		},
		{
			name:         "command cannnot be invoked",
			cmd:          exec.Command("testdata/ignore_sigterm.pl-xxxxxxxxxxxxxxxxxxxxx"),
			duration:     1 * time.Second,
			expectedExit: 127, // TODO cmd should return 125 on win
			skipOnWin:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipOnWin && isWin {
				t.Skipf("%s: skip on windows", tc.name)
			}
			tio := &Timeout{
				Duration:  tc.duration,
				KillAfter: tc.killAfter,
				Cmd:       tc.cmd,
				Signal:    tc.signal,
			}
			exit := tio.RunSimple(tc.preserveStatus)
			if exit != tc.expectedExit {
				t.Errorf("%s: expected exitcode: %d, but: %d", tc.name, tc.expectedExit, exit)
			}
		})
	}
}
