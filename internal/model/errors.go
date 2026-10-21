package model

import (
	"errors"
	"fmt"
	"strings"
)

// Error points at the exact place a model is wrong: which file, which path
// inside it, and what was expected there.
type Error struct {
	File string
	Path string
	Want string
	Err  error
}

func (e *Error) Error() string {
	var parts []string

	if e.File != "" {
		parts = append(parts, e.File)
	}
	if e.Path != "" {
		parts = append(parts, e.Path)
	}

	message := strings.Join(parts, ": ")
	if e.Want != "" {
		if message != "" {
			message += ": "
		}
		message += "expected " + e.Want
	}
	if e.Err != nil {
		if message != "" {
			message += ": "
		}
		message += e.Err.Error()
	}
	if message == "" {
		return "the model could not be read"
	}
	return message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func at(file, path, want string) *Error {
	return &Error{File: file, Path: path, Want: want}
}

func atf(file, path, format string, args ...any) *Error {
	return &Error{File: file, Path: path, Err: fmt.Errorf(format, args...)}
}

// ErrNoModel says the site carries no Quiqr model at all, which is a different
// condition from a model that failed to parse.
var ErrNoModel = errors.New("this site has no Quiqr model")

// Warning is something worth telling the site implementer that does not stop
// the model from being used.
type Warning struct {
	File    string `json:"file,omitempty"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

func (w Warning) String() string {
	var parts []string
	if w.File != "" {
		parts = append(parts, w.File)
	}
	if w.Path != "" {
		parts = append(parts, w.Path)
	}
	parts = append(parts, w.Message)
	return strings.Join(parts, ": ")
}
