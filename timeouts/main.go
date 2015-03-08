package main

import (
	"fmt"

	"github.com/Songmu/timeouts"
)

func main() {
	tio := &timeouts.Timeouts{
		Command: "test/countup.pl",
	}
	exit := tio.Run()

	fmt.Printf("command exited with: %d\n", exit)
}
