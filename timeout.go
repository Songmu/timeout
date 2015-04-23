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

// ExitStatus stores exit information of the command
type ExitStatus struct {
	Code int
	typ  exitType
}

// IsTimedOut returns the command timed out or not
func (ex ExitStatus) IsTimedOut() bool {
	return ex.typ == exitTypeTimedOut || ex.typ == exitTypeKilled
}

// IsKilled returns the command is killed or not
func (ex ExitStatus) IsKilled() bool {
	return ex.typ == exitTypeKilled
}

// GetExitCode gets the exit code for command line tools
func (ex ExitStatus) GetExitCode() int {
	switch {
	case ex.IsKilled():
		return exitKilled
	case ex.IsTimedOut():
		return exitTimedOut
	default:
		return ex.Code
	}
}


// GetChildExitCode gets the exit code of the Cmd itself
func (ex ExitStatus) GetChildExitCode() int {
	return ex.Code
}

type exitType int

// exit types
const (
	exitTypeNormal exitType = iota + 1
	exitTypeTimedOut
	exitTypeKilled
)

func (tio *Timeout) signal() os.Signal {
	if tio.Signal == nil {
		return defaultSignal
	}
	return tio.Signal
}

// Run is synchronous interface of executing command and returning information
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

// RunSimple executes command and only returns integer as exit code. It is mainly for go-timeout command
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
	if preserveStatus {
		return exitSt.GetChildExitCode()
	}
	return exitSt.GetExitCode()
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
		ex.typ = exitTypeNormal
		return ex
	case <-time.After(tio.Duration):
		cmd.Process.Signal(tio.signal()) // XXX error handling
		ex.typ = exitTypeTimedOut
	}

	if tio.KillAfter > 0 {
		select {
		case ex.Code = <-exitChan:
		case <-time.After(tio.KillAfter):
			cmd.Process.Kill()
			ex.Code = exitKilled
			ex.typ = exitTypeKilled
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
		// The exit codes in some platforms aren't integer. e.g. plan9.
		return -1
	}
	return exitNormal
}
