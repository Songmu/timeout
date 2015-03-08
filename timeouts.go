package timeouts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

type Timeouts struct {
	PreserveStatus bool
	Duration       uint64
	KillAfter      uint64
	Signal         os.Signal
	Command        string
	CommandArgs    []string
}

const (
	exitTimedOut = 124
	exitKilled   = 137
)

func (tio *Timeouts) Run() int {
	ch, stdoutPipe, stderrPipe, err := tio.RunCommand()
	if err != nil {
		panic(fmt.Sprintf("something went wrong: %+v", err))
	}
	defer func() {
		stdoutPipe.Close()
		stderrPipe.Close()
	}()

	go readAndOut(stdoutPipe, os.Stdout)
	go readAndOut(stderrPipe, os.Stderr)

	return <-ch
}

func (tio *Timeouts) prepareCmd() *exec.Cmd {
	args := tio.CommandArgs
	return exec.Command(tio.Command, args...)
}

func (tio *Timeouts) RunCommand() (exitChan chan int, stdoutPipe, stderrPipe io.ReadCloser, err error) {
	cmd := tio.prepareCmd()
	if err != nil {
		return
	}

	stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}

	exitChan = make(chan int)
	go func() {
		exitChan <- tio.handleTimeout(cmd)
	}()

	return
}

func (tio *Timeouts) handleTimeout(cmd *exec.Cmd) int {
	exit := 0
	timedOut := false
	exitChan := getExitChan(cmd)

	if tio.Duration > 0 {
		select {
		case exit = <-exitChan:
		case <-time.After(time.Duration(tio.Duration) * time.Second):
			cmd.Process.Signal(tio.Signal)
			timedOut = true
			exit = exitTimedOut
		}
	} else {
		exit = <-exitChan
	}

	killed := false
	if timedOut {
		tmpExit := 0
		if tio.KillAfter > 0 {
			select {
			case tmpExit = <-exitChan:
			case <-time.After(time.Duration(tio.KillAfter) * time.Second):
				cmd.Process.Kill()
				killed = true
				exit = exitKilled
			}
		} else {
			tmpExit = <-exitChan
		}
		if tio.PreserveStatus && !killed {
			exit = tmpExit
		}
	}

	return exit
}

func getExitChan(cmd *exec.Cmd) chan int {
	ch := make(chan int)
	go func() {
		exit := 0
		err := cmd.Wait()
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exit = status.ExitStatus()
				}
			} else {
				exit = -1
			}
		}
		ch <- exit
	}()
	return ch
}

func readAndOut(r io.Reader, f *os.File) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fmt.Fprintln(f, s.Text())
	}
}

var durRe = regexp.MustCompile("^([0-9]+)([smhd])?$")

func parseDuration(durStr string) (uint64, error) {
	matches := durRe.FindStringSubmatch(durStr)
	if len(matches) == 0 {
		return 0, fmt.Errorf("duration format invalid: %s", durStr)
	}

	base, _ := strconv.ParseUint(matches[1], 10, 64)
	switch matches[2] {
	case "", "s":
		return base, nil
	case "m":
		return base * 60, nil
	case "h":
		return base * 60 * 60, nil
	case "d":
		return base * 60 * 60 * 24, nil
	default:
		return 0, fmt.Errorf("something went wrong")
	}
}
