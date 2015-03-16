// Package timeout is for handling timeout invocation of external command
package timeout

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"syscall"
	"time"
)

// Timeout is main struct of timeout package
type Timeout struct {
	Duration  time.Duration
	KillAfter time.Duration
	Signal    os.Signal
	Cmd       *exec.Cmd
}

var defaultSignal os.Signal

func init() {
	switch runtime.GOOS {
	case "windows":
		defaultSignal = os.Interrupt
	default:
		defaultSignal = syscall.SIGTERM
	}
}

// exit statuses are same with GNU timeout
const (
	exitNormal            = 0
	exitTimedOut          = 124
	exitUnknownErr        = 125
	exitCommandNotInvoked = 126
	exitCommandNotFound   = 127
	exitKilled            = 137
)

// Error is error of timeout
type Error struct {
	ExitCode int
	Err      error
}

func (err *Error) Error() string {
	return fmt.Sprintf("exit code: %d, %s", err.ExitCode, err.Err.Error())
}

// ExitStatus stores exit informations of the command
type ExitStatus struct {
	Code int
	Type exitType
}

func (ex ExitStatus) String() string {
	return fmt.Sprintf("exitCode: %d, type: %s", ex.Code, ex.Type)
}

func (ex ExitStatus) IsTimedOut() bool {
	return ex.Type == ExitTypeTimedOut
}

func (ex ExitStatus) IsKilled() bool {
	return ex.Type == ExitTypeKilled
}

func (ex ExitStatus) GetExitCode(preserveStatus bool) int {
	switch {
	case ex.IsKilled():
		return exitKilled
	case preserveStatus:
		return ex.Code
	case ex.IsTimedOut():
		return exitTimedOut
	default:
		return ex.Code
	}
}

type exitType int

// exit types
const (
	ExitTypeNormal exitType = iota + 1
	ExitTypeTimedOut
	ExitTypeKilled
)

func (eTyp exitType) String() string {
	switch eTyp {
	case ExitTypeNormal:
		return "normal"
	case ExitTypeTimedOut:
		return "timeout"
	case ExitTypeKilled:
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

// Run is synchronous interface of exucuting command and returning informations
func (tio *Timeout) Run() (ExitStatus, string, string, error) {
	cmd := tio.Cmd
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	ch, err := tio.RunCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitStatus{}, string(outBuffer.Bytes()), string(errBuffer.Bytes()), err
	}
	exitSt := <-ch
	return exitSt, string(outBuffer.Bytes()), string(errBuffer.Bytes()), nil
}

// RunSimple execute command and only returns integer as exit code. It is mainly for go-timeout command
func (tio *Timeout) RunSimple(preserveStatus bool) int {
	cmd := tio.Cmd

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitUnknownErr
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitUnknownErr
	}

	ch, err := tio.RunCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return getExitCodeFromErr(err)
	}

	go func() {
		defer stdoutPipe.Close()
		io.Copy(os.Stdout, stdoutPipe)
	}()

	go func() {
		defer stderrPipe.Close()
		io.Copy(os.Stderr, stderrPipe)
	}()

	exitSt := <-ch
	return exitSt.GetExitCode(preserveStatus)
}

func getExitCodeFromErr(err error) int {
	if err != nil {
		if tmerr, ok := err.(*Error); ok {
			return tmerr.ExitCode
		}
		return -1
	}
	return exitNormal
}

// RunCommand is executing the command and handling timeout. This is primitive interface of Timeout
func (tio *Timeout) RunCommand() (chan ExitStatus, error) {
	cmd := tio.Cmd

	if err := cmd.Start(); err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, &Error{
				ExitCode: exitCommandNotFound,
				Err:      err,
			}
		case os.IsPermission(err):
			return nil, &Error{
				ExitCode: exitCommandNotInvoked,
				Err:      err,
			}
		default:
			return nil, &Error{
				ExitCode: exitUnknownErr,
				Err:      err,
			}
		}
	}

	exitChan := make(chan ExitStatus)
	go func() {
		exitChan <- tio.handleTimeout()
	}()

	return exitChan, nil
}

func (tio *Timeout) handleTimeout() (ex ExitStatus) {
	cmd := tio.Cmd
	exitChan := getExitChan(cmd)
	select {
	case exitCode := <-exitChan:
		ex.Code = exitCode
		ex.Type = ExitTypeNormal
		return ex
	case <-time.After(tio.Duration):
		cmd.Process.Signal(tio.signal()) // XXX error handling
		ex.Type = ExitTypeTimedOut
	}

	if tio.KillAfter > 0 {
		select {
		case ex.Code = <-exitChan:
		case <-time.After(tio.KillAfter):
			cmd.Process.Kill()
			ex.Code = exitKilled
			ex.Type = ExitTypeKilled
		}
	} else {
		ex.Code = <-exitChan
	}

	return ex
}

func getExitChan(cmd *exec.Cmd) chan int {
	ch := make(chan int)
	go func() {
		err := cmd.Wait()
		ch <- resolveCode(err)
	}()
	return ch
}

func resolveCode(err error) int {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		// The exit codes in some platforms aren't integer. e.g. plan9.
		return -1
	}
	return exitNormal
}
