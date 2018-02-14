package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var sigmap = map[string]os.Signal{
	"INT":  os.Interrupt,
	"TERM": syscall.SIGTERM,
}

func main() {
	var (
		trap     = flag.String("trap", "", "signals")
		exit     = flag.Int("exit", 0, "exit status")
		trapExit = flag.Int("trap-exit", 0, "exit status when trapping signal")
		sleep    = flag.Float64("sleep", 0, "sleep seconds")
	)
	flag.Parse()

	if *trap != "" {
		var sigs []os.Signal
		for _, sigStr := range strings.Split(*trap, ",") {
			sig, ok := sigmap[strings.TrimPrefix(strings.ToUpper(sigStr), "SIG")]
			if !ok {
				log.Printf("unknown signal name: %s\n", sigStr)
				os.Exit(1)
			}
			sigs = append(sigs, sig)
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c, sigs...)
		go func() {
			for _ = range c {
				if *trapExit > 0 {
					os.Exit(*trapExit)
				}
			}
		}()
	}
	if *sleep > 0 {
		time.Sleep(time.Duration(float64(time.Second) * *sleep))
	}
	os.Exit(*exit)
}
