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

func TestRunSimple(t *testing.T) {
	tio := &Timeout{
		Duration: time.Duration(0.1 * float64(time.Second)),
		Cmd:      exec.Command(shellcmd, shellflag, "echo 1"),
	}
	exit := tio.RunSimple(false)

	if exit != 0 {
		t.Errorf("something wrong")
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

func TestRunTimeout(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command(shellcmd, shellflag, "sleep 3"),
		Duration: 1 * time.Second,
		Signal:   os.Interrupt,
	}
	exit := tio.RunSimple(false)

	if exit != 124 {
		t.Errorf("something wrong")
	}
}

func TestPreserveStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("skip test on Windows")
	}
	tio := &Timeout{
		Cmd:      exec.Command("perl", "testdata/exit_with_23.pl"),
		Duration: 1 * time.Second,
	}

	exit := tio.RunSimple(true)
	if exit != 23 {
		t.Errorf("something wrong: %v", exit)
	}
}

func TestKillAfter(t *testing.T) {
	tio := &Timeout{
		Cmd:       exec.Command("perl", "testdata/ignore_sigterm.pl"),
		Signal:    syscall.SIGTERM,
		Duration:  1 * time.Second,
		KillAfter: 1 * time.Second,
	}
	exit := tio.RunSimple(true)

	if exit != 137 {
		t.Errorf("something wrong: %v", exit)
	}
}

func TestKillAfterNotKilled(t *testing.T) {
	tio := &Timeout{
		Cmd:       exec.Command("perl", "testdata/ignore_sigterm.pl"),
		Signal:    syscall.SIGTERM,
		Duration:  1 * time.Second,
		KillAfter: 5 * time.Second,
	}
	exit := tio.RunSimple(true)

	if exit != 0 {
		t.Errorf("something wrong: %v", exit)
	}
}

func TestCommandCannotBeInvoked(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO cmd return 125 for this case
		t.Skipf("skip test on Windows")
	}
	tio := &Timeout{
		Cmd:      exec.Command("testdata/dummy"),
		Duration: 1 * time.Second,
	}
	exit := tio.RunSimple(false)

	if exit != 126 {
		t.Errorf("something wrong: %v", exit)
	}
}

func TestCommandNotFound(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO cmd return 125 for this case
		t.Skipf("skip test on Windows")
	}
	tio := &Timeout{
		Cmd:      exec.Command("testdata/ignore_sigterm.pl-xxxxxxxxxxxxxxxxxxxxx"),
		Duration: 1 * time.Second,
	}
	exit := tio.RunSimple(false)

	if exit != 127 {
		t.Errorf("something wrong: %v", exit)
	}
}
