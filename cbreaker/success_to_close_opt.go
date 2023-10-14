package cbreaker

import (
	"fmt"
)

type SuccessToCloseOpt int

func WithSuccessToClose(n int) SuccessToCloseOpt {
	return SuccessToCloseOpt(n)
}

func (n SuccessToCloseOpt) apply(conf *config) error {
	if n <= 0 {
		return fmt.Errorf(
			"success_count=%d: %w",
			int(n),
			ErrInvalidSuccessToClose,
		)
	}

	conf.nbSuccessToClose = int(n)

	return nil
}
