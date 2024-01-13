package conch

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func Test_interceptPanic(t *testing.T) {
	t.Run(
		"when ok", func(t *testing.T) {
			v, err := interceptPanic82763487263(
				func(
					ctx context.Context,
					_ int,
					p int,
				) (int, error) {
					return p*13 + 1, nil
				},
			)(context.Background(), 0, 100)

			require.NoError(t, err)
			require.Equal(t, 1301, v)
		},
	)

	t.Run(
		"when error", func(t *testing.T) {
			v, err := interceptPanic82763487263(
				func(
					ctx context.Context,
					_ int,
					p int,
				) (int, error) {
					return 0, ErrMocked
				},
			)(context.Background(), 0, 100)

			require.ErrorIs(t, err, ErrMocked)
			require.Zero(t, v)
		},
	)

	t.Run(
		"when panicking", func(t *testing.T) {
			v, err := interceptPanic82763487263(
				func(
					ctx context.Context,
					_ int,
					p int,
				) (
					int,
					error,
				) {
					panic("test")
					return 0, nil
				},
			)(
				context.Background(),
				0,
				100,
			)

			require.Zero(t, v)

			var errPanic *PanicError
			require.ErrorAs(t, err, &errPanic)
		},
	)
}

func lcgRequest(
	_ context.Context,
	p uint64,
) (
	uint64,
	error,
) {
	return lcg(p)
}

func lcgRequestSlow(d time.Duration) func(
	ctx context.Context,
	p uint64,
) (
	uint64,
	error,
) {
	return func(
		ctx context.Context,
		p uint64,
	) (
		uint64,
		error,
	) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(d):
			return lcgRequest(ctx, p)
		}
	}
}

func TestRequestConsumersC(t *testing.T) {
	t.Run(
		"answer requests", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			const (
				concurrency  = 10
				requestCount = 1_000
			)

			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var group sync.WaitGroup

			var inStreams []chan Request[uint64, uint64]
			var inStreamsAdapt []<-chan Request[uint64, uint64]
			for i := 0; i < concurrency; i++ {
				c := make(chan Request[uint64, uint64])
				inStreams = append(inStreams, c)
				inStreamsAdapt = append(inStreamsAdapt, c)
			}

			RequestConsumersC(ExposeID(lcgRequest))(ctx, &group, inStreamsAdapt...)

			requests := NewTestRequests(
				requestCount,
				rnd,
				context.Background,
			)

			var collectedCount atomic.Uint64
			var runGroup sync.WaitGroup
			requests.SendToConcurrently(&runGroup, rnd, inStreams...)
			requests.CollectConcurrently(
				&runGroup,
				func(
					ctx context.Context,
					p uint64,
					r uint64,
					err error,
				) {
					expectedValue, expectedErr := lcg(p)
					require.Equal(t, expectedValue, r)
					require.ErrorIs(t, err, expectedErr)
					collectedCount.Add(1)
				},
			)
			runGroup.Wait()

			for _, inStream := range inStreams {
				close(inStream)
			}

			cancel()
			wgWait(t, &group, time.Second, 10*time.Millisecond)
			require.Equal(t, collectedCount.Load(), uint64(requestCount))
		},
	)

	t.Run(
		"closing main ctx", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			for i := 0; i < 100; i++ {
				const (
					concurrency = 2
				)

				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

				ctx, cancel := context.WithTimeout(
					context.Background(),
					50*time.Millisecond,
				)

				var group sync.WaitGroup

				var inStreams []chan Request[uint64, uint64]
				var inStreamsAdapt []<-chan Request[uint64, uint64]
				for i := 0; i < concurrency; i++ {
					c := make(chan Request[uint64, uint64])
					inStreams = append(inStreams, c)
					inStreamsAdapt = append(inStreamsAdapt, c)
				}

				RequestConsumersC(ExposeID(lcgRequest))(ctx, &group, inStreamsAdapt...)

				var runGroup sync.WaitGroup

				runGroup.Add(1)
				go func() {
					defer runGroup.Done()

					respChan := make(chan ValErrorPair[uint64])
					request := Request[uint64, uint64]{
						Ctx:  context.Background(),
						Chan: respChan,
					}

					for {
						p := rnd.Uint64()
						request.P = p

						inStreams[rnd.Intn(concurrency)] <- request
						response := <-respChan

						if errors.Is(
							response.Err,
							context.DeadlineExceeded,
						) {
							require.ErrorIs(
								t,
								ctx.Err(),
								context.DeadlineExceeded,
							)
							return
						}
					}
				}()

				runGroup.Wait()

				for _, inStream := range inStreams {
					close(inStream)
				}

				cancel()
				wgWait(t, &group, time.Second, 10*time.Millisecond)
			}
		},
	)

	t.Run(
		"closing inner ctx", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			const (
				concurrency = 1
			)

			ctx, cancel := context.WithCancel(
				context.Background(),
			)

			var group sync.WaitGroup

			var inStreams []chan Request[uint64, uint64]
			var inStreamsAdapt []<-chan Request[uint64, uint64]
			for i := 0; i < concurrency; i++ {
				c := make(chan Request[uint64, uint64])
				inStreams = append(inStreams, c)
				inStreamsAdapt = append(inStreamsAdapt, c)
			}

			RequestConsumersC(ExposeID(lcgRequestSlow(time.Second)))(
				ctx,
				&group,
				inStreamsAdapt...,
			)

			innerCtx, cancelInner := context.WithTimeout(
				context.Background(),
				50*time.Millisecond,
			)
			defer cancelInner()

			respChan := make(chan ValErrorPair[uint64])

			request := Request[uint64, uint64]{
				Ctx:  innerCtx,
				Chan: respChan,
				P:    1234,
			}

			inStreams[0] <- request
			valErr := <-respChan

			require.Equal(t, uint64(0), valErr.V)
			require.ErrorIs(t, valErr.Err, context.DeadlineExceeded)

			for _, inStream := range inStreams {
				close(inStream)
			}

			cancel()

			wgWait(t, &group, time.Second, 10*time.Millisecond)
		},
	)

	t.Run(
		"panic when processing is nil", func(t *testing.T) {
			require.PanicsWithValue(
				t, ErrNilRequestProcessingFunc, func() {
					RequestConsumersC[int, int](
						nil,
					)(
						nil,
						nil,
						(chan Request[int, int])(nil),
					)
				},
			)

		},
	)
}

type TestRequest struct {
	r Request[uint64, uint64]
	c chan ValErrorPair[uint64]
}

type TestRequests []*TestRequest

func NewTestRequests(
	count int,
	rnd *rand.Rand,
	ctxProvider func() context.Context,
) (requests TestRequests) {
	requests = make(TestRequests, count)

	for i := 0; i < count; i++ {
		param := rnd.Uint64()
		respChan := make(chan ValErrorPair[uint64])
		requests[i] = &TestRequest{
			r: Request[uint64, uint64]{
				P:    param,
				Ctx:  ctxProvider(),
				Chan: respChan,
			},
			c: respChan,
		}
	}

	return
}

func (r TestRequests) SendToConcurrently(
	group *sync.WaitGroup,
	rnd *rand.Rand,
	inStreams ...chan Request[uint64, uint64],
) {
	group.Add(len(r))

	for idx, request := range r {
		go func(
			i int,
			g *sync.WaitGroup,
			r *TestRequest,
			s chan Request[uint64, uint64],
		) {
			defer g.Done()
			s <- r.r
		}(
			idx,
			group,
			request,
			inStreams[rnd.Intn(len(inStreams))],
		)
	}
}

func (r TestRequests) CollectConcurrently(
	group *sync.WaitGroup,
	f func(context.Context, uint64, uint64, error),
) {
	group.Add(len(r))

	for idx, request := range r {
		go func(
			i int,
			g *sync.WaitGroup,
			r *TestRequest,
		) {
			defer g.Done()

			valErr := <-(r.c)
			f(r.r.Ctx, r.r.P, valErr.V, valErr.Err)
		}(
			idx,
			group,
			request,
		)
	}
}

func (r TestRequests) Visit(f func(request *TestRequest)) {
	for _, request := range r {
		f(request)
	}
}
