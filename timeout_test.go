package timeout

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestRun(t *testing.T) {
	tio := &Timeout{
		Cmd: exec.Command("/bin/sh", "-c", "echo 1"),
	}
	exit := tio.Run()

	if exit != 0 {
		t.Errorf("something wrong")
	}
}

func TestRunTimeout(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("/bin/sh", "-c", "sleep 3"),
		Duration: 1,
		Signal:   os.Interrupt,
	}
	exit := tio.Run()

	if exit != 124 {
		t.Errorf("something wrong")
	}
}

func TestPreserveStatus(t *testing.T) {
	tio := &Timeout{
		Cmd:            exec.Command("perl", "test/exit_with_23.pl"),
		Duration:       1,
		PreserveStatus: true,
	}

	exit := tio.Run()
	if exit != 23 {
		t.Errorf("something wrong")
	}
}

func TestKillAfter(t *testing.T) {
	tio := &Timeout{
		Cmd:            exec.Command("perl", "test/ignore_sigterm.pl"),
		Signal:         syscall.SIGTERM,
		Duration:       1,
		KillAfter:      1,
		PreserveStatus: true,
	}
	exit := tio.Run()

	if exit != 137 {
		t.Errorf("something wrong")
	}
}

func TestCommandCannotBeInvoked(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("test/dummy"),
		Duration: 1,
	}
	exit := tio.Run()

	if exit != 126 {
		t.Errorf("something wrong")
	}
}

func TestCommandNotFound(t *testing.T) {
	tio := &Timeout{
		Cmd:      exec.Command("test/ignore_sigterm.pl-xxxxxxxxxxxxxxxxxxxxx"),
		Duration: 1,
	}
	exit := tio.Run()

	if exit != 127 {
		t.Errorf("something wrong")
	}
}

func TestParseDuration(t *testing.T) {
	v, err := ParseDuration("55s")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 55 {
		t.Errorf("parse failed!")
	}

	v, err = ParseDuration("55")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 55 {
		t.Errorf("parse failed!")
	}

	v, err = ParseDuration("10m")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 600 {
		t.Errorf("parse failed!")
	}

	v, err = ParseDuration("1h")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 3600 {
		t.Errorf("parse failed!")
	}

	v, err = ParseDuration("1d")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 86400 {
		t.Errorf("parse failed!")
	}

	_, err = ParseDuration("1w")
	if err == nil {
		t.Errorf("something wrong")
	}
}
