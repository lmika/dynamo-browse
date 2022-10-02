package models

import "github.com/pkg/errors"

var ErrReadOnly = errors.New("in read-only mode")
