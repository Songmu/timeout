package main

import "os/exec"

func main() {
	cmd := exec.Command("test/countup.pl")
	cmd.Run()
}
