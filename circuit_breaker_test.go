package conch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	// ErrBadass represent an error where ....
	ErrBadass = errors.New("badass")

	// ErrUnavailable represent an error where....
	ErrUnavailable = errors.New("unavailable")
)

func Test_PassingNoError(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	be := NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(false).
		Once()
	be.
		On("ReportSuccess").
		Return().
		Once()

	inStream := make(chan Request[int, int])

	BreakerC(
		be,
		ErrUnavailable,
		func(
			ctx context.Context,
			group *sync.WaitGroup,
			inStream <-chan Request[int, int],
		) {
			group.Add(1)

			go func() {
				defer group.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case req, more := <-inStream:
						if !more {
							return
						}

						select {
						case <-ctx.Done():
						case req.Chan <- ValErrorPair[int]{
							V:   2*req.P + 1,
							Err: nil,
						}:
						}
					}
				}
			}()
		},
	)(ctx, &wg, inStream)

	outC := make(chan ValErrorPair[int])

	require.Eventually(t, func() bool {
		inStream <- Request[int, int]{
			P:    101,
			Ctx:  ctx,
			Chan: outC,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		require.Equal(
			t,
			ValErrorPair[int]{
				V: 101*2 + 1,
			},
			<-outC,
		)

		return true
	}, time.Second, time.Millisecond)

	close(inStream)

	wg.Wait()
}

func Test_Blocking(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	be := NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(true).
		Once()

	inStream := make(chan Request[int, int])

	BreakerC(
		be,
		ErrUnavailable,
		func(
			ctx context.Context,
			group *sync.WaitGroup,
			inStream <-chan Request[int, int],
		) {
			group.Add(1)

			go func() {
				defer group.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case req, more := <-inStream:
						if !more {
							return
						}

						select {
						case <-ctx.Done():
						case req.Chan <- ValErrorPair[int]{
							V:   2*req.P + 1,
							Err: nil,
						}:
						}
					}
				}
			}()
		},
	)(ctx, &wg, inStream)

	outC := make(chan ValErrorPair[int])

	require.Eventually(t, func() bool {
		inStream <- Request[int, int]{
			P:    0,
			Ctx:  ctx,
			Chan: outC,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		response := <-outC

		require.Equal(
			t,
			ValErrorPair[int]{
				Err: ErrUnavailable,
			},
			response,
		)

		return true
	}, time.Second, time.Millisecond)

	close(inStream)

	wg.Wait()
}

func Test_PassingError(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	be := NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(false).
		Once()
	be.
		On("ReportFailure").
		Return().
		Once()

	inStream := make(chan Request[int, int])

	BreakerC(
		be,
		ErrUnavailable,
		func(
			ctx context.Context,
			group *sync.WaitGroup,
			inStream <-chan Request[int, int],
		) {
			group.Add(1)

			go func() {
				defer group.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case req, more := <-inStream:
						if !more {
							return
						}

						select {
						case <-ctx.Done():
						case req.Chan <- ValErrorPair[int]{
							Err: ErrBadass,
						}:
						}
					}
				}
			}()
		},
	)(ctx, &wg, inStream)

	outC := make(chan ValErrorPair[int])

	require.Eventually(t, func() bool {
		inStream <- Request[int, int]{
			P:    101,
			Ctx:  ctx,
			Chan: outC,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		require.Equal(
			t,
			ValErrorPair[int]{
				Err: ErrBadass,
			},
			<-outC,
		)

		return true
	}, time.Second, time.Millisecond)

	close(inStream)

	wg.Wait()
}
