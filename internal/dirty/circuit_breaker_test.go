package dirty

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

var (
	// ErrBadass represent an error where ....
	ErrBadass = errors.New("badass")

	// ErrUnavailable represent an error where....
	ErrUnavailable = errors.New("unavailable")
)

func Test_BreakerPassingNoError(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	be := dirty.NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(false).
		Once()
	be.
		On("ReportSuccess").
		Return().
		Once()

	dirty.scaffoldTest(
		t,
		func(
			t *testing.T,
			inStream chan dirty.Request[int, int],
			wg *sync.WaitGroup,
		) {

			BreakerC(
				be,
				ErrUnavailable,
				func(
					ctx context.Context,
					group *sync.WaitGroup,
					inStream <-chan dirty.Request[int, int],
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
								case req.Chan <- dirty.ValErrorPair[int]{
									V:   2*req.P + 1,
									Err: nil,
								}:
								}
							}
						}
					}()
				},
			)(ctx, wg, inStream)

			outC := make(chan dirty.ValErrorPair[int])

			require.Eventually(t, func() bool {
				inStream <- dirty.Request[int, int]{
					P:    101,
					Ctx:  ctx,
					Chan: outC,
				}

				return true
			}, time.Second, time.Millisecond)

			require.Eventually(t, func() bool {
				require.Equal(
					t,
					dirty.ValErrorPair[int]{
						V: 101*2 + 1,
					},
					<-outC,
				)

				return true
			}, time.Second, time.Millisecond)

			close(inStream)
		},
	)
}

func Test_BreakerBlocking(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	be := dirty.NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(true).
		Once()

	dirty.scaffoldTest(
		t,
		func(
			t *testing.T,
			inStream chan dirty.Request[int, int],
			wg *sync.WaitGroup,
		) {
			BreakerC(
				be,
				ErrUnavailable,
				func(
					ctx context.Context,
					group *sync.WaitGroup,
					inStream <-chan dirty.Request[int, int],
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
								case req.Chan <- dirty.ValErrorPair[int]{
									V:   2*req.P + 1,
									Err: nil,
								}:
								}
							}
						}
					}()
				},
			)(ctx, wg, inStream)

			outC := make(chan dirty.ValErrorPair[int])

			require.Eventually(t, func() bool {
				inStream <- dirty.Request[int, int]{
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
					dirty.ValErrorPair[int]{
						Err: ErrUnavailable,
					},
					response,
				)

				return true
			}, time.Second, time.Millisecond)

			close(inStream)
		},
	)
}

func Test_BreakerPassingError(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	be := dirty.NewMockBreakerEngine(t)
	be.
		On("IsOpen").
		Return(false).
		Once()
	be.
		On("ReportFailure").
		Return().
		Once()

	dirty.scaffoldTest(
		t,
		func(
			t *testing.T,
			inStream chan dirty.Request[int, int],
			wg *sync.WaitGroup,
		) {
			BreakerC(
				be,
				ErrUnavailable,
				func(
					ctx context.Context,
					group *sync.WaitGroup,
					inStream <-chan dirty.Request[int, int],
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
								case req.Chan <- dirty.ValErrorPair[int]{
									Err: ErrBadass,
								}:
								}
							}
						}
					}()
				},
			)(ctx, wg, inStream)

			outC := make(chan dirty.ValErrorPair[int])

			require.Eventually(t, func() bool {
				inStream <- dirty.Request[int, int]{
					P:    101,
					Ctx:  ctx,
					Chan: outC,
				}

				return true
			}, time.Second, time.Millisecond)

			require.Eventually(t, func() bool {
				require.Equal(
					t,
					dirty.ValErrorPair[int]{
						Err: ErrBadass,
					},
					<-outC,
				)

				return true
			}, time.Second, time.Millisecond)

			close(inStream)
		},
	)
}

func Test_BreakerClosingInput(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	be := dirty.NewMockBreakerEngine(t)

	dirty.scaffoldTest(
		t,
		func(
			t *testing.T,
			inStream chan dirty.Request[int, int],
			wg *sync.WaitGroup,
		) {
			BreakerC(
				be,
				ErrUnavailable,
				func(
					ctx context.Context,
					group *sync.WaitGroup,
					inStream <-chan dirty.Request[int, int],
				) {
					group.Add(1)

					go func() {
						defer group.Done()

						for {
							select {
							case <-ctx.Done():
								return
							case req, more := <-inStream:
								require.False(t, more)
								require.Equal(t, dirty.Request[int, int]{}, req)

								return
							}
						}
					}()
				},
			)(ctx, wg, inStream)

			close(inStream)
		},
	)
}

func Test_BreakerCancelCtx(t *testing.T) {
	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	be := dirty.NewMockBreakerEngine(t)

	dirty.scaffoldTest(
		t,
		func(
			t *testing.T,
			inStream chan dirty.Request[int, int],
			wg *sync.WaitGroup,
		) {
			BreakerC(
				be,
				ErrUnavailable,
				func(
					ctx context.Context,
					group *sync.WaitGroup,
					inStream <-chan dirty.Request[int, int],
				) {
					group.Add(1)

					go func() {
						defer group.Done()

						for {
							select {
							case <-ctx.Done():
								return
							case req, more := <-inStream:
								require.False(t, more)
								require.Equal(t, dirty.Request[int, int]{}, req)

								return
							}
						}
					}()
				},
			)(ctx, wg, inStream)

			cancel()
		},
	)
}
