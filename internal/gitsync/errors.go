package gitsync

import (
	"errors"
	"fmt"
)

// OperationError says which git operation failed and why, with anything
// credential-shaped already removed from the message.
type OperationError struct {
	Operation string
	Detail    string
	Err       error
}

func (e *OperationError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("git %s failed: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("git %s failed: %s", e.Operation, e.Detail)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

var (
	// ErrNothingChanged says a save wrote content identical to what was
	// already there, so there is no new version. It is not a failure.
	ErrNothingChanged = errors.New("nothing changed, so no version was made")

	// ErrNotAWorkingCopy says the directory holds no site.
	ErrNotAWorkingCopy = errors.New("this directory is not a site working copy")

	// ErrAlreadyAWorkingCopy says a clone would discard what is already there.
	ErrAlreadyAWorkingCopy = errors.New("this directory already holds a site working copy")

	// ErrEscapingPath says a save named a path outside the site.
	ErrEscapingPath = errors.New("this path is outside the site")

	// ErrPathInConflict says a path cannot be read because two people changed
	// it and the question has not been answered yet.
	ErrPathInConflict = errors.New("this page was edited by two people and the difference is unresolved")

	// ErrNoExchange says there is nothing to resolve or abandon.
	ErrNoExchange = errors.New("no exchange with the remote is in progress")

	// ErrExchangeInProgress says an exchange must be finished or abandoned
	// before another can start.
	ErrExchangeInProgress = errors.New("an exchange with the remote is already in progress")

	// ErrNotConflicted says a resolution named a path that is not in conflict.
	ErrNotConflicted = errors.New("this path is not in conflict")

	// ErrRemoteMoved says the remote gained versions since the last exchange,
	// so publishing would discard them.
	ErrRemoteMoved = errors.New("the remote has changes this site does not have yet")

	// ErrRemoteUnreachable says the remote did not answer.
	ErrRemoteUnreachable = errors.New("the remote could not be reached")
)
