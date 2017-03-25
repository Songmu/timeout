package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/Songmu/timeout"
	"github.com/pborman/getopt"
)

func main() {
	optKillAfter := getopt.StringLong("kill-after", 'k', "", "also send a KILL signal if COMMAND is still running. this long after the initial signal was sent")
	optSig := getopt.StringLong("signal", 's', "", "specify the signal to be sent on timeout. IGNAL may be a name like 'HUP' or a number. see 'kill -l' for a list of signals")
	optForeground := getopt.BoolLong("foreground", 'f', "when not running timeout directly from a shell prompt, allow COMMAND to read from the TTY and get TTY signals. in this mode, children of COMMAND will not be timed out")
	p := getopt.BoolLong("preserve-status", 0, "help message for bool")

	opts := getopt.CommandLine
	opts.Parse(os.Args)

	rest := opts.Args()
	if len(rest) < 2 {
		opts.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	var err error
	killAfter := float64(0)
	if *optKillAfter != "" {
		killAfter, err = parseDuration(*optKillAfter)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(125)
		}
	}

	var sig os.Signal
	if *optSig != "" {
		sig, err = parseSignal(*optSig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(125)
		}
	}

	dur, err := parseDuration(rest[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(125)
	}

	cmd := exec.Command(rest[1], rest[2:]...)

	tio := &timeout.Timeout{
		Duration:   time.Duration(dur * float64(time.Second)),
		Cmd:        cmd,
		Foreground: *optForeground,
		KillAfter:  time.Duration(killAfter * float64(time.Second)),
		Signal:     sig,
	}
	exit := tio.RunSimple(*p)
	os.Exit(exit)
}

var durRe = regexp.MustCompile(`^([-0-9e.]+)([smhd])?$`)

func parseDuration(durStr string) (float64, error) {
	matches := durRe.FindStringSubmatch(durStr)
	if len(matches) == 0 {
		return 0, fmt.Errorf("duration format invalid: %s", durStr)
	}

	base, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid time interval `%s`", durStr)
	}
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
		return 0, fmt.Errorf("invalid time interval `%s`", durStr)
	}
}
