package retrier

import (
	"errors"
	"fmt"
	"time"
)

const (
	// goldenRatioValue represents the mathematical constant Phi, the golden ratio.
	goldenRatioValue    = 1.61803398874989484820458683436563811772030917980576286213544862270526046281890244970720720418939113748475408807538689175212663386222353693179318
	DefaultRetryDelay   = 100 * time.Millisecond
	DefaultJitterFactor = goldenRatioValue - 1
)

// config holds the configuration parameters for retry behavior.
// It defines limits, delays, and functions that control retry logic.
type config struct {
	maxRetry    *int           // Maximum number of retry attempts (nil for unlimited)
	maxDuration *time.Duration // Maximum total duration for all retry attempts

	retryDelay    *time.Duration // Initial delay between retry attempts
	jitterFactor  *float64      // Factor for adding randomness to retry delays
	backoffFactor *float64      // Factor for exponential backoff growth

	mustRetryFunc func(error) bool // Function to determine if an error should trigger retry
	maxRetryDelay *time.Duration    // Maximum delay between individual retry attempts
}

// ErrInvalidValue represent an error where ....
var ErrInvalidValue = errors.New("invalid value")

// ErrDuplicateOption represent an error where ....
var ErrDuplicateOption = errors.New("duplicate option")

// ErrInvalidConfiguration represent an error where ....
var ErrInvalidConfiguration = errors.New("invalid configuration")

type MustRetryFuncOpt func(error) bool

func (m MustRetryFuncOpt) apply(conf *config) error {
	if m == nil {
		return fmt.Errorf(
			"mil must retry function: %w",
			ErrInvalidValue,
		)
	}

	if conf.mustRetryFunc != nil {
		return fmt.Errorf(
			"must retry func option: %w",
			ErrDuplicateOption,
		)
	}
	conf.mustRetryFunc = m

	return nil
}

func WithMustRetryFunc(mustRetryFunc func(error) bool) Option {
	return MustRetryFuncOpt(mustRetryFunc)
}

type BackoffFactorOpt float64

func (b BackoffFactorOpt) apply(conf *config) error {
	if conf.backoffFactor != nil {
		return fmt.Errorf(
			"backoff factor option: %w",
			ErrDuplicateOption,
		)
	}

	v := float64(b)

	if v < 1 {
		return fmt.Errorf(
			"backoff factor option is out of range: must be >= 1 %w",
			ErrInvalidValue,
		)
	}

	conf.backoffFactor = &v

	return nil
}

func WithBackoffFactor(backoffFactor float64) Option {
	return BackoffFactorOpt(backoffFactor)
}

type JitterFactorOpt float64

func (j JitterFactorOpt) apply(conf *config) error {
	if conf.jitterFactor != nil {
		return fmt.Errorf(
			"jitter factor option: %w",
			ErrDuplicateOption,
		)
	}

	v := float64(j)
	if v < 0 || v > 1 {
		return fmt.Errorf(
			"jitter factor option is out of range: must be [0.0, 1.0] %w",
			ErrInvalidValue,
		)
	}

	conf.jitterFactor = &v

	return nil
}

func WithJitterFactor(jitterFactor float64) Option {
	return JitterFactorOpt(jitterFactor)
}

func WithNoJitter() Option {
	return JitterFactorOpt(0)
}

type RetryDelayOpt time.Duration

func (r RetryDelayOpt) apply(conf *config) error {
	if conf.retryDelay != nil {
		return fmt.Errorf(
			"retry delay option: %w",
			ErrDuplicateOption,
		)
	}

	v := time.Duration(r)
	if v == 0 {
		return fmt.Errorf(
			"retry delay option is zero: %w",
			ErrInvalidValue,
		)
	}

	conf.retryDelay = &v

	return nil
}

func WithRetryDelay(retryDelay time.Duration) Option {
	return RetryDelayOpt(retryDelay)
}

type MaxTotalDurationOpt time.Duration

func (m MaxTotalDurationOpt) apply(conf *config) error {
	if conf.maxDuration != nil {
		return fmt.Errorf(
			"max duration option: %w",
			ErrDuplicateOption,
		)
	}

	if conf.maxRetry != nil {
		return fmt.Errorf(
			"max duration option set with max retry: %w",
			ErrInvalidConfiguration,
		)
	}

	v := time.Duration(m)
	conf.maxDuration = &v

	return nil
}

func WithMaxTotalDuration(maxDuration time.Duration) Option {
	return MaxTotalDurationOpt(maxDuration)
}

type MaxRetryDelay time.Duration

func (m MaxRetryDelay) apply(conf *config) error {
	if conf.maxRetryDelay != nil {
		return fmt.Errorf(
			"max retry delay option: %w",
			ErrDuplicateOption,
		)
	}

	v := time.Duration(m)
	conf.maxRetryDelay = &v

	return nil
}

func WithMaxRetryDelay(maxDuration time.Duration) Option {
	return MaxRetryDelay(maxDuration)
}

type MaxRetryOpt int

func (m MaxRetryOpt) apply(conf *config) error {
	if conf.maxRetry != nil {
		return fmt.Errorf(
			"max retry option: %w",
			ErrDuplicateOption,
		)
	}

	if conf.maxDuration != nil {
		return fmt.Errorf(
			"max retry option set with max duration: %w",
			ErrInvalidConfiguration,
		)
	}

	v := int(m)
	conf.maxRetry = &v
	return nil
}

func WithMaxRetry(maxRetry int) Option {
	return MaxRetryOpt(maxRetry)
}

func mustRetry(err error) bool {
	return errors.Is(err, ErrRetryable)
}

type Option interface {
	apply(conf *config) error
}

func ref[T any](v T) *T {
	return &v
}

func (c *config) establishRetryDefaults() {
	if c.retryDelay == nil {
		c.retryDelay = ref(DefaultRetryDelay)
	}

	if c.jitterFactor == nil {
		c.jitterFactor = ref(DefaultJitterFactor)
	}

	if c.backoffFactor == nil {
		c.backoffFactor = ref(goldenRatioValue)
	}

	if c.mustRetryFunc == nil {
		c.mustRetryFunc = mustRetry
	}
}
