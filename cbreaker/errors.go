package cbreaker

import (
	"errors"
)

var (
	// ErrInvalidFailureToOpen represent an error where ....
	ErrInvalidFailureToOpen = errors.New("must be non zero positive integer")

	// ErrInvalidSuccessToClose represent an error where ....
	ErrInvalidSuccessToClose = errors.New("must be non zero positive integer")

	// ErrInvalidOpenTimeout represent an error where ....
	ErrInvalidOpenTimeout = errors.New("must be non zero positive duration")
)
