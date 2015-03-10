package main

import (
	"fmt"
	"syscall"

	"github.com/Songmu/timeout"
)

func main() {
	tio := &timeout.Timeout{
		Command: "test/countup.pl",
		Signal:  syscall.SIGTERM,
	}
	exit := tio.Run()

	fmt.Printf("command exited with: %d\n", exit)
}
