package cbreaker

import (
	"fmt"
	"time"
)

type OpenTimeoutOpt time.Duration

func (n OpenTimeoutOpt) apply(conf *config) error {
	if n <= 0 {
		return fmt.Errorf(
			"success_count=%d: %w",
			time.Duration(n),
			ErrInvalidOpenTimeout,
		)
	}

	conf.halfOpenTimeout = time.Duration(n)

	return nil
}

func WithOpenTimeout(timeout time.Duration) OpenTimeoutOpt {
	return OpenTimeoutOpt(timeout)
}
