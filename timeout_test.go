package timeout

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestRunSimple(t *testing.T) {
	tio := &Timeout{
		Duration: time.Duration(0.1 * float64(time.Second)),
		Cmd:      exec.Command("/bin/sh", "-c", "echo 1"),
	}
	exit := tio.RunSimple(false)

	if exit != 0 {
		t.Errorf("something wrong")
	}
}

func TestRun(t *testing.T) {
	tio := &Timeout{
		Duration: 10 * time.Second,
		Cmd:      exec.Command("/bin/sh", "-c", "echo 1"),
	}
	_, stdout, stderr, err := tio.Run()

	if stdout != "1\n" {
		t.Errorf("something wrong")
	}

	if stderr != "" {
		t.Errorf("something wrong")
	}

	if err != nil {
		t.Errorf("something wrong")
	}
}

func TestRunTimeout(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("/bin/sh", "-c", "sleep 3"),
		Duration: 1 * time.Second,
		Signal:   os.Interrupt,
	}
	exit := tio.RunSimple(false)

	if exit != 124 {
		t.Errorf("something wrong")
	}
}

func TestPreserveStatus(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("perl", "test/exit_with_23.pl"),
		Duration: 1 * time.Second,
	}

	exit := tio.RunSimple(true)
	if exit != 23 {
		t.Errorf("something wrong")
	}
}

func TestKillAfter(t *testing.T) {
	tio := &Timeout{
		Cmd:       exec.Command("perl", "test/ignore_sigterm.pl"),
		Signal:    syscall.SIGTERM,
		Duration:  1 * time.Second,
		KillAfter: 1 * time.Second,
	}
	exit := tio.RunSimple(true)

	if exit != 137 {
		t.Errorf("something wrong")
	}
}

func TestKillAfterNotKilled(t *testing.T) {
	tio := &Timeout{
		Cmd:       exec.Command("perl", "test/ignore_sigterm.pl"),
		Signal:    syscall.SIGTERM,
		Duration:  1 * time.Second,
		KillAfter: 5 * time.Second,
	}
	exit := tio.RunSimple(true)

	if exit != 0 {
		t.Errorf("something wrong")
	}
}

func TestCommandCannotBeInvoked(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("test/dummy"),
		Duration: 1 * time.Second,
	}
	exit := tio.RunSimple(false)

	if exit != 126 {
		t.Errorf("something wrong")
	}
}

func TestCommandNotFound(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("test/ignore_sigterm.pl-xxxxxxxxxxxxxxxxxxxxx"),
		Duration: 1 * time.Second,
	}
	exit := tio.RunSimple(false)

	if exit != 127 {
		t.Errorf("something wrong")
	}
}
