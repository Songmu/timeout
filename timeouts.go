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
)

type Timeouts struct {
	PreserveStatus bool
	Duration       uint64
	KillAfter      uint64
	Signal         os.Signal
	Command        string
	CommandArgs    []string
}

func (tio *Timeouts) Run() int {
	ch, stdoutPipe, stderrPipe, err := tio.RunCommand()
	if err != nil {
		panic("something went wrong")
	}

	defer func() {
		stdoutPipe.Close()
		stderrPipe.Close()
	}()

	go func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	}(stdoutPipe)

	go func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			fmt.Fprintln(os.Stderr, s.Text())
		}
	}(stderrPipe)

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
	go func(cmd *exec.Cmd) {
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
		exitChan <- exit
	}(cmd)

	return
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
