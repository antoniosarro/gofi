package entry

import "errors"

var (
	// ErrMissingName indicates the entry has no name
	ErrMissingName = errors.New("entry: missing name")

	// ErrMissingExec indicates the entry has no executable command
	ErrMissingExec = errors.New("entry: missing exec command")

	// ErrLaunchFailed indicates the launch operation failed
	ErrLaunchFailed = errors.New("entry: launch failed")

	// ErrNoTerminal indicates no terminal emulator was found
	ErrNoTerminal = errors.New("entry: no terminal emulator found")

	// ErrEmptyExec indicates the exec command is empty after parsing
	ErrEmptyExec = errors.New("entry: empty exec command after parsing")
)
