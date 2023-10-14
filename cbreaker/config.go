package cbreaker

import (
	"time"
)

type (
	config struct {
		nbFailureToOpen  int
		nbSuccessToClose int
		halfOpenTimeout  time.Duration
	}

	Option interface {
		apply(conf *config) error
	}
)

var (
	defaultConfig = config{
		nbFailureToOpen:  4,
		nbSuccessToClose: 2,
		halfOpenTimeout:  5 * time.Second,
	}
)
