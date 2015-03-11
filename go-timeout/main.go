package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/Songmu/timeout"
)

func main() {
	tio := &timeout.Timeout{
		Cmd:    exec.Command("test/countup.pl"),
		Signal: syscall.SIGTERM,
	}
	exit := tio.Run()

	fmt.Printf("command exited with: %d\n", exit)

	os.Exit(exit)
}
