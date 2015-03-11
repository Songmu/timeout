package timeout

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type Timeout struct {
	PreserveStatus bool
	Duration       uint64
	KillAfter      uint64
	Signal         os.Signal
	Cmd            *exec.Cmd
}

var defaultSignal = func() os.Signal {
	if runtime.GOOS == "windows" {
		return os.Interrupt
	}
	return syscall.SIGTERM
}()

// exit statuses are same with GNU timeout
const (
	exitTimedOut          = 124
	exitUnknownErr        = 125
	exitCommandNotInvoked = 126
	exitCommandNotFound   = 127
	exitKilled            = 137
)

func (tio *Timeout) signal() os.Signal {
	if tio.Signal == nil {
		return defaultSignal
	}
	return tio.Signal
}

func (tio *Timeout) Run() int {
	ch, stdoutPipe, stderrPipe, err := tio.RunCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err.ExitCode
	}
	defer func() {
		stdoutPipe.Close()
		stderrPipe.Close()
	}()

	go readAndOut(stdoutPipe, os.Stdout)
	go readAndOut(stderrPipe, os.Stderr)

	return <-ch
}

type tmError struct {
	ExitCode int
	message  string
}

func (err *tmError) Error() string {
	return err.message
}

func (tio *Timeout) RunCommand() (exitChan chan int, stdoutPipe, stderrPipe io.ReadCloser, tmerr *tmError) {
	cmd := tio.Cmd

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		tmerr = &tmError{
			ExitCode: exitUnknownErr,
			message:  fmt.Sprintf("unknown error: %s", err),
		}
		return
	}
	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		tmerr = &tmError{
			ExitCode: exitUnknownErr,
			message:  fmt.Sprintf("unknown error: %s", err),
		}
		return
	}
	if err = cmd.Start(); err != nil {
		switch {
		case os.IsNotExist(err):
			tmerr = &tmError{
				ExitCode: exitCommandNotFound,
				message:  err.Error(),
			}
		case os.IsPermission(err):
			tmerr = &tmError{
				ExitCode: exitCommandNotInvoked,
				message:  err.Error(),
			}
		default:
			tmerr = &tmError{
				ExitCode: exitUnknownErr,
				message:  fmt.Sprintf("unknown error: %s", err),
			}
		}
		return
	}

	exitChan = make(chan int)
	go func() {
		exitChan <- tio.handleTimeout()
	}()

	return
}

func (tio *Timeout) handleTimeout() int {
	exit := 0
	cmd := tio.Cmd
	timedOut := false
	exitChan := getExitChan(cmd)

	if tio.Duration > 0 {
		select {
		case exit = <-exitChan:
		case <-time.After(time.Duration(tio.Duration) * time.Second):
			cmd.Process.Signal(tio.signal())
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
		err := cmd.Wait()
		ch <- resolveExitCode(err)
	}()
	return ch
}

func resolveExitCode(err error) int {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return -1 // XXX
	}
	return 0
}

func readAndOut(r io.Reader, f *os.File) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fmt.Fprintln(f, s.Text())
	}
}

var durRe = regexp.MustCompile(`^([0-9]+)([smhd])?$`)

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
