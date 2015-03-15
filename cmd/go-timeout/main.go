package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/Songmu/timeout"

	"code.google.com/p/getopt"
)

func main() {
	optKillAfter := getopt.StringLong("kill-after", 'k', "", "help message for f")
	optSig := getopt.StringLong("signal", 's', "", "help message for long")
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
		Duration:       dur,
		Cmd:            cmd,
		KillAfter:      killAfter,
		PreserveStatus: *p,
		Signal:         sig,
	}
	exit := tio.RunSimple()
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

func parseSignal(sigStr string) (os.Signal, error) {
	switch strings.ToUpper(sigStr) {
	case "":
		return nil, nil
	case "HUP", "1":
		return syscall.SIGHUP, nil
	case "INT", "2":
		return os.Interrupt, nil
	case "QUIT", "3":
		return syscall.SIGQUIT, nil
	case "KILL", "9":
		return os.Kill, nil
	case "ALRM", "14":
		return syscall.SIGALRM, nil
	case "TERM", "15":
		return syscall.SIGTERM, nil
	case "USR1":
		return syscall.SIGUSR1, nil
	case "USR2":
		return syscall.SIGUSR2, nil
	default:
		return nil, fmt.Errorf("%s: invalid signal", sigStr)
	}
}
