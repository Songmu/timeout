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

type exitCode int

// exit statuses are same with GNU timeout
const (
	exitNormal            exitCode = 0
	exitTimedOut                   = 124
	exitUnknownErr                 = 125
	exitCommandNotInvoked          = 126
	exitCommandNotFound            = 127
	exitKilled                     = 137
)

type tmError struct {
	ExitCode exitCode
	message  string
}

func (err *tmError) Error() string {
	return err.message
}

type exitState struct {
	ExitCode exitCode
	ExitType exitType
}

func (exSt exitState) String() string {
	return fmt.Sprintf("exitCode: %d, type: %s", exSt.ExitCode, exSt.ExitType)
}

type exitType int

const (
	normal exitType = iota + 1
	timedOut
	killed
)

func (eTyp exitType) String() string {
	switch eTyp {
	case normal:
		return "normal"
	case timedOut:
		return "timeout"
	case killed:
		return "killed"
	default:
		return "unknown"
	}
}

func (tio *Timeout) signal() os.Signal {
	if tio.Signal == nil {
		return defaultSignal
	}
	return tio.Signal
}

func (tio *Timeout) Run() exitCode {
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

func (tio *Timeout) RunCommand() (exitChan chan exitCode, stdoutPipe, stderrPipe io.ReadCloser, tmerr *tmError) {
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

	exitChan = make(chan exitCode)
	go func() {
		exitChan <- tio.handleTimeout()
	}()

	return
}

func (tio *Timeout) handleTimeout() exitCode {
	exit := exitNormal
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
		tmpExit := exitNormal
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

func getExitChan(cmd *exec.Cmd) chan exitCode {
	ch := make(chan exitCode)
	go func() {
		err := cmd.Wait()
		ch <- resolveExitCode(err)
	}()
	return ch
}

func resolveExitCode(err error) exitCode {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return exitCode(status.ExitStatus())
			}
		}
		// XXX The exit codes in some platforms aren't integer. e.g. plan9.
		return exitCode(-1)
	}
	return exitNormal
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
