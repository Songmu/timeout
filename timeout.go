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

// exit statuses are same with GNU timeout
const (
	exitNormal     = 0
	exitTimedOut   = 124
	exitUnknownErr = 125
	exitKilled     = 137
)

// overwritten with syscall.SIGTERM on unix environment (see timeout_unix.go)
var defaultSignal = os.Interrupt

// Error is error of timeout
type Error struct {
	ExitCode int
	Err      error
}

func (err *Error) Error() string {
	return fmt.Sprintf("exit code: %d, %s", err.ExitCode, err.Err.Error())
}

// Timeout is main struct of timeout package
type Timeout struct {
	Duration   time.Duration
	KillAfter  time.Duration
	Signal     os.Signal
	Foreground bool
	Cmd        *exec.Cmd

	KillAfterCancel time.Duration
}

func (tio *Timeout) signal() os.Signal {
	if tio.Signal == nil {
		return defaultSignal
	}
	return tio.Signal
}

// Run is synchronous interface of executing command and returning information
func (tio *Timeout) Run() (*ExitStatus, string, string, error) {
	cmd := tio.getCmd()
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	ch, err := tio.RunCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, string(outBuffer.Bytes()), string(errBuffer.Bytes()), err
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
func (tio *Timeout) RunContext(ctx context.Context) (*ExitStatus, error) {
	if err := tio.start(); err != nil {
		return nil, err
	}
	return tio.wait(ctx), nil
}

// RunCommand is executing the command and handling timeout. This is primitive interface of Timeout
func (tio *Timeout) RunCommand() (<-chan *ExitStatus, error) {
	if err := tio.start(); err != nil {
		return nil, err
	}

	exitChan := make(chan *ExitStatus)
	go func() {
		exitChan <- tio.wait(context.Background())
	}()
	return exitChan, nil
}

func (tio *Timeout) start() error {
	if err := tio.getCmd().Start(); err != nil {
		return &Error{
			ExitCode: wrapcommander.ResolveExitCode(err),
			Err:      err,
		}
	}
	return nil
}

func (tio *Timeout) wait(ctx context.Context) *ExitStatus {
	ex := &ExitStatus{}
	cmd := tio.getCmd()
	exitChan := getExitChan(cmd)
	killCh := make(chan struct{}, 2)
	done := make(chan struct{})
	defer close(done)

	delayedKill := func(dur time.Duration) {
		select {
		case <-done:
			return
		case <-time.After(dur):
			killCh <- struct{}{}
		}
	}

	if tio.KillAfter > 0 {
		go delayedKill(tio.Duration + tio.KillAfter)
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
			ex.killed = true
			if ex.typ != exitTypeCanceled {
				ex.typ = exitTypeKilled
			}
		case <-ctx.Done():
			// XXX handling etx.Err()?
			tio.terminate()
			ex.typ = exitTypeCanceled
			go delayedKill(tio.getKillAfterCancel())
		}
	}
}

func (tio *Timeout) getKillAfterCancel() time.Duration {
	if tio.KillAfterCancel == 0 {
		return 3 * time.Second
	}
	return tio.KillAfterCancel
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
