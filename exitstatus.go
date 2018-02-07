package timeout

// ExitStatus stores exit information of the command
type ExitStatus struct {
	Code     int
	Signaled bool
	typ      exitType
}

// IsTimedOut returns the command timed out or not
func (ex *ExitStatus) IsTimedOut() bool {
	return ex.typ == exitTypeTimedOut || ex.typ == exitTypeKilled
}

// IsKilled returns the command is killed or not
func (ex *ExitStatus) IsKilled() bool {
	return ex.typ == exitTypeKilled
}

// GetExitCode gets the exit code for command line tools
func (ex *ExitStatus) GetExitCode() int {
	switch {
	case ex.IsKilled():
		return exitKilled
	case ex.IsTimedOut():
		return exitTimedOut
	default:
		return ex.Code
	}
}

// GetChildExitCode gets the exit code of the Cmd itself
func (ex *ExitStatus) GetChildExitCode() int {
	return ex.Code
}

type exitType int

// exit types
const (
	exitTypeNormal exitType = iota
	exitTypeTimedOut
	exitTypeKilled
)
