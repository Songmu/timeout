// Package timeout is for handling timeout invocation of external command
package timeout

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Songmu/wrapcommander"
)

// overwritten with syscall.SIGTERM on unix environment (see timeout_unix.go)
var defaultSignal = os.Interrupt

// Timeout is main struct of timeout package
type Timeout struct {
	Duration   time.Duration
	KillAfter  time.Duration
	Signal     os.Signal
	Foreground bool
	Cmd        *exec.Cmd
}

// exit statuses are same with GNU timeout
const (
	exitNormal     = 0
	exitTimedOut   = 124
	exitUnknownErr = 125
	exitKilled     = 137
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
	Code     int
	Signaled bool
	typ      exitType
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
	exitTypeNormal exitType = iota
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
	cmd := tio.getCmd()
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
	cmd := tio.getCmd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ch, err := tio.RunCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return getExitCodeFromErr(err)
	}

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

// RunContext runs command with context
func (tio *Timeout) RunContext(ctx context.Context) (ExitStatus, error) {
	exChan, err := tio.runContext(ctx)
	if err != nil {
		return ExitStatus{}, err
	}
	return <-exChan, nil
}

// RunCommand is executing the command and handling timeout. This is primitive interface of Timeout
func (tio *Timeout) RunCommand() (<-chan ExitStatus, error) {
	return tio.runContext(context.Background())
}

func (tio *Timeout) runContext(ctx context.Context) (<-chan ExitStatus, error) {
	cmd := tio.getCmd()

	if err := cmd.Start(); err != nil {
		return nil, &Error{
			ExitCode: wrapcommander.ResolveExitCode(err),
			Err:      err,
		}
	}

	exitChan := make(chan ExitStatus)
	go func() {
		exitChan <- tio.handleTimeout(ctx)
	}()

	return exitChan, nil
}

func (tio *Timeout) handleTimeout(ctx context.Context) (ex ExitStatus) {
	cmd := tio.getCmd()
	exitChan := getExitChan(cmd)
	var killCh <-chan time.Time
	if tio.KillAfter > 0 {
		killCh = time.After(tio.Duration + tio.KillAfter)
	}
	for {
		select {
		case st := <-exitChan:
			ex.Code = wrapcommander.WaitStatusToExitCode(st)
			ex.Signaled = st.Signaled()
			return ex
		case <-time.After(tio.Duration):
			tio.terminate()
			ex.typ = exitTypeTimedOut
		case <-killCh:
			tio.killall()
			// just to make sure
			cmd.Process.Kill()
			ex.typ = exitTypeKilled
		}
	}
}

func getExitChan(cmd *exec.Cmd) chan syscall.WaitStatus {
	ch := make(chan syscall.WaitStatus)
	go func() {
		err := cmd.Wait()
		st, _ := wrapcommander.ErrorToWaitStatus(err)
		ch <- st
	}()
	return ch
}
