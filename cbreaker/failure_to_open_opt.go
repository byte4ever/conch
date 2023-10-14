package cbreaker

import (
	"fmt"
)

type FailureToOpenOpt int

func WithFailuresToOpen(n int) FailureToOpenOpt {

	return FailureToOpenOpt(n)
}

func (n FailureToOpenOpt) apply(conf *config) error {
	if n <= 0 {
		return fmt.Errorf(
			"failure_count=%d: %w",
			int(n),
			ErrInvalidFailureToOpen,
		)
	}

	conf.nbFailureToOpen = int(n)

	return nil
}
