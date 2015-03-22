timeout
=======

[![Build Status](https://travis-ci.org/Songmu/timeout.png?branch=master)][travis]
[![Coverage Status](https://coveralls.io/repos/Songmu/timeout/badge.png?branch=master)][coveralls]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[travis]: https://travis-ci.org/Songmu/timeout
[coveralls]: https://coveralls.io/r/Songmu/timeout?branch=master
[license]: https://github.com/Songmu/timeout/blob/master/LICENSE

Timeout invocation.

## Description

Run a given command with a time limit.


## Disclaimer

This software is still alpha quality. We may change APIs without notice.

## Synopsis

	tio := &timeout.Timeout{
		Cmd:            exec.Command("perl", "-E", "say 'Hello'"),
		Duration:       10 * time.Second,
		KillAfter:      5 * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()

## Author

[Songmu](https://github.com/Songmu)
