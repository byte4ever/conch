package retrier

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"iter"
	"math"
	mrand "math/rand/v2"
	"time"

	"github.com/byte4ever/conch/domain"
)

var (

	// ErrRetryable represents an error that indicates a retryable condition.
	ErrRetryable = errors.New("retryable error")

	// ErrReachMaxRetry indicates the maximum number of retry attempts was reached.
	ErrReachMaxRetry = errors.New("reach max retry")

	// ErrReachMaxDuration indicates the maximum allowable duration was reached.
	ErrReachMaxDuration = errors.New("reach max duration")
)

// RetryableError wraps an error to indicate it is a retryable condition.
// Err is the underlying error wrapped by RetryableError.
type RetryableError struct {
	Err error
}

// Error returns the string representation of the RetryableError.
func (re *RetryableError) Error() string {
	return fmt.Sprintf("%s: %s", ErrRetryable.Error(), re.Err.Error())
}

// Unwrap returns the underlying error contained in the RetryableError.
func (re *RetryableError) Unwrap() error {
	return re.Err
}

// Is checks if the given error matches the ErrRetryable error type.
func (re *RetryableError) Is(e error) bool {
	return errors.Is(e, ErrRetryable)
}

// WrapToRetryableError wraps the given error into a RetryableError instance.
func WrapToRetryableError(e error) *RetryableError {
	return &RetryableError{
		Err: e,
	}
}

// floatChan is a read-only channel providing float64 values asynchronously.
var floatChan <-chan float64 //nolint:gochecknoglobals // ok

func init() {
	var key [32]byte
	_, _ = rand.Read(key[:32]) //nolint:errcheck // really don't care

	rnd := mrand.New(mrand.NewChaCha8(key))

	c := make(chan float64, 100)

	go func() {
		for {
			c <- rnd.Float64()
		}
	}()

	floatChan = c
}

// Retrier is a generic type for handling retryable operations.
// It encapsulates retry logic, configuration, and coefficient calculations.
type Retrier struct {

	// config represents the retry configuration parameters.
	config

	jitterFunc func() float64
	growthFunc func(int) float64
}

// New creates a new Retrier instance with the given configuration options.
// It returns a pointer to the Retrier and an error if configuration fails.
func New(options ...Option) (*Retrier, error) {
	var config config

	for _, o := range options {
		err := o.apply(&config)
		if err != nil {
			return nil, err
		}
	}

	config.establishRetryDefaults()

	result := &Retrier{
		config:     config,
		jitterFunc: jitterFunc(*config.jitterFactor),
		growthFunc: growthFunc(*config.backoffFactor),
	}

	return result, nil
}

// WrapProcessorFunc wraps a ProcessorFunc with retry logic using a Retrier.
// retrier defines the retry configuration and policies.
// f is the original ProcessorFunc that will be wrapped with retry handling.
// Returns a wrapped ProcessorFunc with integrated retry behavior.
func WrapProcessorFunc[
	T domain.Processable[Param, Result],
	Param any,
	Result any](
	retrier *Retrier,
	f domain.ProcessorFunc[T, Param, Result],
) domain.ProcessorFunc[T, Param, Result] {
	return func(ctx context.Context, elem T) {
		if retrier.maxDuration == nil {
			if retrier.maxRetry == nil {
				executeForEver(
					ctx,
					&retrier.config,
					retrier.delayGenerator(),
					f,
					elem,
				)
				return
			}

			executeMaxAttempts(
				ctx,
				&retrier.config,
				retrier.delayGenerator(),
				f,
				elem,
			)
			return
		}

		executeMaxDuration(
			ctx,
			&retrier.config,
			retrier.delayGenerator(),
			f,
			elem,
		)
	}
}

// executeMaxDuration attempts processing within a maximum allowed duration.
// ctx is the context to handle cancellation or timeouts.
// c provides configuration for maximum duration and retry policies.
// delayGenerator generates retry intervals with backoff and jitter.
// f is the processing function to execute with the provided element.
// elem is the processable object managed by the function.
func executeMaxDuration[T domain.Processable[Param, Result], Param any, Result any](
	ctx context.Context,
	c *config,
	delayGenerator iter.Seq2[int, time.Duration],
	f domain.ProcessorFunc[T, Param, Result],
	elem T,
) {
	f(ctx, elem)
	if !c.mustRetryFunc(elem.GetError()) {
		return
	}

	globalTimer := time.NewTimer(*c.maxDuration)
	timer := time.NewTimer(0)

	for _, w := range delayGenerator {
		fmt.Println(w)
		timer.Reset(w)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			if !globalTimer.Stop() {
				<-globalTimer.C
			}

			elem.SetError(ctx.Err())
			return

		case <-timer.C:

		case <-globalTimer.C:
			if !timer.Stop() {
				<-timer.C
			}

			elem.SetError(ErrReachMaxDuration)
			return
		}

		f(ctx, elem)
		if !c.mustRetryFunc(elem.GetError()) {
			return
		}
	}

	panic("unreachable")
}

func executeForEver[T domain.Processable[Param, Result], Param any, Result any](
	ctx context.Context,
	c *config,
	delayGenerator iter.Seq2[int, time.Duration],
	f domain.ProcessorFunc[T, Param, Result],
	elem T,
) {
	f(ctx, elem)
	if !c.mustRetryFunc(elem.GetError()) {
		return
	}

	timer := time.NewTimer(0)

	for _, w := range delayGenerator {
		fmt.Println(w)
		timer.Reset(w)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			elem.SetError(ctx.Err())
			return

		case <-timer.C:
		}

		f(ctx, elem)
		if !c.mustRetryFunc(elem.GetError()) {
			return
		}
	}

	panic("unreachable")
}

func executeMaxAttempts[T domain.Processable[Param, Result], Param any, Result any](
	ctx context.Context,
	c *config,
	delayGenerator iter.Seq2[int, time.Duration],
	f domain.ProcessorFunc[T, Param, Result],
	elem T,
) {
	f(ctx, elem)
	if !c.mustRetryFunc(elem.GetError()) {
		return
	}

	attempt := 1

	timer := time.NewTimer(0)

	for idx, w := range delayGenerator {
		fmt.Println(attempt, idx+1)

		if idx+1 >= *c.maxRetry {
			elem.SetError(ErrReachMaxRetry)
			return
		}

		timer.Reset(w)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			elem.SetError(ctx.Err())
			return

		case <-timer.C:
		}

		f(ctx, elem)
		if !c.mustRetryFunc(elem.GetError()) {
			return
		}

		attempt++
	}

	panic("unreachable")
}

// jitterFunc returns a function to calculate jitter values dynamically.
// If jitterFactor is 0, it returns a constant jitter value of 1.
// Otherwise, it generates jitter using the jitterFactor and floatChan values.
func jitterFunc(jitterFactor float64) func() float64 {
	if jitterFactor == 0 {
		return constantJitter
	}
	return func() float64 {
		return 1 + jitterFactor*((<-floatChan*2)-1.0)
	}
}

// constantJitter returns a fixed jitter value of 1.0.
func constantJitter() float64 {
	return 1.0
}

// backoffCache defines the fixed limit for precomputing backoff values.
const backoffCache = 25

// growthFunc generates a function to calculate exponential backoff values.
// The returned function takes an int (retry count) and returns a float64.
// It precomputes values up to a fixed cache limit for efficiency.
// If `backoffFactor` is 1, all values will be 1.0.
func growthFunc(backoffFactor float64) func(int) float64 {
	if backoffFactor == 1 {
		return func(int) float64 {
			return 1.0
		}
	}

	values := make([]float64, 0, backoffCache)

	for i := 0; i < backoffCache; i++ {
		values = append(values, math.Pow(backoffFactor, float64(i)))
	}

	return func(i int) float64 {
		if i >= len(values) {
			return math.Pow(backoffFactor, float64(i))
		}

		return values[i]
	}
}

// delayGenerator generates a sequence of retry intervals with backoff and jitter.
func (t *Retrier) delayGenerator() iter.Seq2[int, time.Duration] {
	if t.maxRetryDelay == nil {
		return func(yield func(int, time.Duration) bool) {
			if !yield(0, *t.retryDelay) {
				return
			}

			i := 1
			for {
				if !yield(i, time.Duration(
					t.growthFunc(i)*
						t.jitterFunc()*
						float64(t.retryDelay.Nanoseconds()),
				)) {
					return
				}
				i++
			}
		}
	}

	return func(yield func(int, time.Duration) bool) {
		if !yield(0, *t.retryDelay) {
			return
		}

		i := 1
		for {
			w := time.Duration(
				t.growthFunc(i) *
					t.jitterFunc() *
					float64(t.retryDelay.Nanoseconds()),
			)

			if w > *t.maxRetryDelay {
				break
			}

			if !yield(i, w) {
				return
			}

			i++
		}

		for {
			if !yield(i, time.Duration(
				t.jitterFunc()*
					float64(t.maxRetryDelay.Nanoseconds()),
			)) {
				return
			}

			i++
		}
	}
}
