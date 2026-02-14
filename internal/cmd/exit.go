package cmd

import (
	"errors"

	"github.com/failup-ventures/attio-cli/internal/config"
)

const (
	ExitCodeSuccess  = 0
	ExitCodeGeneric  = 1
	ExitCodeUsage    = 2
	ExitCodeNoResult = 3
	ExitCodeAuth     = 4
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type UsageError struct {
	Err error
}

func (e *UsageError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *UsageError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newUsageError(err error) error {
	if err == nil {
		return nil
	}
	return &ExitError{Code: ExitCodeUsage, Err: &UsageError{Err: err}}
}

func stableExitCode(err error) error {
	if err == nil {
		return nil
	}
	if ExitCode(err) != ExitCodeGeneric {
		return err
	}

	var authErr *config.AuthRequiredError
	if errors.As(err, &authErr) {
		return &ExitError{Code: ExitCodeAuth, Err: err}
	}

	return &ExitError{Code: ExitCodeGeneric, Err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitCodeSuccess
	}
	var ee *ExitError
	if errors.As(err, &ee) && ee != nil {
		if ee.Code < 0 {
			return ExitCodeGeneric
		}
		return ee.Code
	}
	return ExitCodeGeneric
}
