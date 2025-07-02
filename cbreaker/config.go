package cbreaker

import (
	"time"
)

type (
	config struct {
		nbFailureToOpen  *int
		nbSuccessToClose *int
		halfOpenTimeout  *time.Duration
	}

	Option interface {
		apply(conf *config) error
	}
)

func ref[T any](v T) *T {
	return &v
}

var (
	defaultConfig = config{ //nolint:gochecknoglobals // thats default
		nbFailureToOpen:  ref(4),
		nbSuccessToClose: ref(2),
		halfOpenTimeout:  ref(5 * time.Second),
	}
)
