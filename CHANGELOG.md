# Changelog

## [v0.3.1](https://github.com/Songmu/timeout/compare/v0.3.0...v0.3.1) (2018-03-04)

* add rough test to detect goroutine leak [#20](https://github.com/Songmu/timeout/pull/20) ([Songmu](https://github.com/Songmu))
* fix goroutine leak [#19](https://github.com/Songmu/timeout/pull/19) ([Songmu](https://github.com/Songmu))

## [v0.3.0](https://github.com/Songmu/timeout/compare/v0.2.1...v0.3.0) (2018-02-14)

* Context support with RunContext method [#16](https://github.com/Songmu/timeout/pull/16) ([Songmu](https://github.com/Songmu))
* [incompatible] use pointer when returning ExitStatus [#17](https://github.com/Songmu/timeout/pull/17) ([Songmu](https://github.com/Songmu))
* test for signaled [#15](https://github.com/Songmu/timeout/pull/15) ([Songmu](https://github.com/Songmu))

## [v0.2.1](https://github.com/Songmu/timeout/compare/v0.2.0...v0.2.1) (2018-01-07)

* send SIGCONT after sending termination signal just to make sure [#14](https://github.com/Songmu/timeout/pull/14) ([Songmu](https://github.com/Songmu))
* remove reflect and refactor [#13](https://github.com/Songmu/timeout/pull/13) ([Songmu](https://github.com/Songmu))

## [v0.2.0](https://github.com/Songmu/timeout/compare/v0.1.0...v0.2.0) (2018-01-07)

* Adjust files for releasing [#12](https://github.com/Songmu/timeout/pull/12) ([Songmu](https://github.com/Songmu))
* adjust testing(introduce table driven test) [#11](https://github.com/Songmu/timeout/pull/11) ([Songmu](https://github.com/Songmu))
* Wait for the command to finish properly and add Signaled field to ExitStatus [#10](https://github.com/Songmu/timeout/pull/10) ([Songmu](https://github.com/Songmu))
* introduce github.com/Songmu/wrapcommander [#9](https://github.com/Songmu/timeout/pull/9) ([Songmu](https://github.com/Songmu))
* update doc [#8](https://github.com/Songmu/timeout/pull/8) ([Songmu](https://github.com/Songmu))

## [v0.1.0](https://github.com/Songmu/timeout/compare/v0.0.1...v0.1.0) (2017-03-26)

* [incompatible] Support Foreground option [#6](https://github.com/Songmu/timeout/pull/6) ([Songmu](https://github.com/Songmu))
* [incompatible] killall child processes when sending SIGKILL on Unix systems [#5](https://github.com/Songmu/timeout/pull/5) ([Songmu](https://github.com/Songmu))
* Call taskkill [#3](https://github.com/Songmu/timeout/pull/3) ([mattn](https://github.com/mattn))
* update ci related files [#4](https://github.com/Songmu/timeout/pull/4) ([Songmu](https://github.com/Songmu))

## [v0.0.1](https://github.com/Songmu/timeout/compare/fca682e36f92...v0.0.1) (2015-04-23)

* Fix document [#2](https://github.com/Songmu/timeout/pull/2) ([syohex](https://github.com/syohex))
* Support windows [#1](https://github.com/Songmu/timeout/pull/1) ([mattn](https://github.com/mattn))
