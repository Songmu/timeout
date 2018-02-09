package timeout

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	shellcmd  = "/bin/sh"
	shellflag = "-c"
	stubCmd   = "./testdata/stubcmd"
)

func init() {
	if runtime.GOOS == "windows" {
		shellcmd = "cmd"
		shellflag = "/c"
		stubCmd = `.\testdata\stubcmd.exe`
	}
	err := exec.Command("go", "build", "-o", stubCmd, "testdata/stubcmd.go").Run()
	if err != nil {
		panic(err)
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
			cmd:          exec.Command(shellcmd, shellflag, fmt.Sprintf("%s -sleep 3", stubCmd)),
			duration:     1 * time.Second,
			signal:       os.Interrupt,
			expectedExit: 124,
		},
		{
			name:           "timed out with preserve status",
			cmd:            exec.Command(shellcmd, shellflag, fmt.Sprintf("%s -sleep 3", stubCmd)),
			duration:       time.Duration(0.1 * float64(time.Second)),
			preserveStatus: true,
			expectedExit:   128 + 15,
			skipOnWin:      true,
		},
		{
			name:           "preserve status (signal trapd)",
			cmd:            exec.Command(stubCmd, "-trap", "SIGTERM", "-trap-exit", "23", "-sleep", "3"),
			duration:       1 * time.Second,
			preserveStatus: true,
			expectedExit:   23,
			skipOnWin:      true,
		},
		{
			name:         "kill after",
			cmd:          exec.Command(stubCmd, "-trap", "SIGTERM", "-sleep", "3"),
			duration:     1 * time.Second,
			killAfter:    1 * time.Second,
			signal:       syscall.SIGTERM,
			expectedExit: exitKilled,
		},
		{
			name:           "trap sigterm but exited before kill after",
			cmd:            exec.Command(stubCmd, "-trap", "SIGTERM", "-sleep", "3"),
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
			name:         "command cannnot be invoked (command not found)",
			cmd:          exec.Command("testdata/command-not-found"),
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

func TestRunContext(t *testing.T) {
	expect := ExitStatus{
		Code:     128 + int(syscall.SIGTERM),
		Signaled: true,
		typ:      exitTypeCanceled,
		killed:   false,
	}
	if isWin {
		expect = ExitStatus{
			Code:     1,
			Signaled: false,
			typ:      exitTypeCanceled,
			killed:   true,
		}
	}

	t.Run("cancel", func(t *testing.T) {
		tio := &Timeout{
			Duration: 3 * time.Second,
			Cmd:      exec.Command(stubCmd, "-sleep", "10"),
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()
		st, err := tio.RunContext(ctx)
		if err != nil {
			t.Errorf("error should be nil but: %s", err)
		}
		if !reflect.DeepEqual(expect, *st) {
			t.Errorf("invalid exit status\n   out: %v\nexpect: %v", *st, expect)
		}
	})

	t.Run("with timeout", func(t *testing.T) {
		tio := &Timeout{
			Duration: 3 * time.Second,
			Cmd:      exec.Command(stubCmd, "-sleep", "10"),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		st, err := tio.RunContext(ctx)
		if err != nil {
			t.Errorf("error should be nil but: %s", err)
		}
		if !reflect.DeepEqual(expect, *st) {
			t.Errorf("invalid exit status\n   out: %v\nexpect: %v", *st, expect)
		}
	})

	t.Run("with timeout and signal trapped", func(t *testing.T) {
		if isWin {
			t.Skip("skip on windows")
		}
		tio := &Timeout{
			Duration:        3 * time.Second,
			Cmd:             exec.Command(stubCmd, "-sleep", "10", "-trap", "SIGTERM"),
			KillAfterCancel: time.Duration(10 * time.Millisecond),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		st, err := tio.RunContext(ctx)
		if err != nil {
			t.Errorf("error should be nil but: %s", err)
		}
		expect := ExitStatus{
			Code:     exitKilled,
			Signaled: true,
			typ:      exitTypeCanceled,
			killed:   true,
		}
		if !reflect.DeepEqual(expect, *st) {
			t.Errorf("invalid exit status\n   out: %v\nexpect: %v", *st, expect)
		}
	})
}
