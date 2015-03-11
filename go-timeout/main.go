package main

import (
	"os"
	"os/exec"

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

	killAfter := uint64(0)
	if *optKillAfter != "" {
		killAfter, _ = timeout.ParseDuration(*optKillAfter)
	}

	var sig os.Signal
	if *optSig != "" {
		sig, _ = timeout.ParseSignal(*optSig)
	}

	dur, _ := timeout.ParseDuration(rest[0])
	cmd := exec.Command(rest[1], rest[2:]...)

	tio := &timeout.Timeout{
		Duration:       dur,
		Cmd:            cmd,
		KillAfter:      killAfter,
		PreserveStatus: *p,
		Signal:         sig,
	}
	exit := tio.Run()
	os.Exit(exit)
}
