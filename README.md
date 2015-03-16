timeout
=======

Timeout invocation.

## Description

Run a given command with a time limit.


## Disclaimer

This software is still alpha quality. We may change APIs without notice.

## Synopsis

	tio := &Timeout{
		Cmd:            exec.Command("perl", "-E", "say 'Hello'"),
		Duration:       10,
		KillAfter:      5,
	}
	exit, stdout, stderr, err := tio.Run()

## Author

[Songmu](https://github.com/Songmu)
