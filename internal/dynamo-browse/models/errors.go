package models

import (
	"github.com/pkg/errors"
)

var ErrReadOnly = errors.New("in read-only mode")

type PartialResultsError struct {
	err error
}

func NewPartialResultsError(err error) PartialResultsError {
	return PartialResultsError{err: err}
}

func (pr PartialResultsError) Error() string {
	return "partial results received"
}

func (pr PartialResultsError) Unwrap() error {
	return pr.err
}
