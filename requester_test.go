package conch

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func f(_ context.Context, v int) (int, error) {
	time.Sleep(10 * time.Millisecond)

	return v, nil
}

type sstat struct {
	i   int
	cnt int64
	md  time.Duration
}

func generateDummy(ctx context.Context) <-chan struct{} {
	outStream := make(chan struct{})

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case outStream <- struct{}{}:
			}
		}
	}()

	return outStream
}

type cc struct {
	mtx sync.Mutex
	d   time.Duration
	cnt int64
}

type ccs []*cc

func newCCS(n int) ccs {
	r := make(ccs, n)

	for i := 0; i < n; i++ {
		r[i] = &cc{}
	}

	return r
}

func (c ccs) refresh(idx int, d time.Duration) {
	p := c[idx]
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.d += d
	p.cnt++
}

func (c ccs) String() string {
	var sb strings.Builder

	for i, p := range c {
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(": ")
		sb.WriteString(strconv.FormatInt(p.cnt, 10))
		sb.WriteString(" ")
		sb.WriteString((p.d / time.Duration(p.cnt)).String())
		sb.WriteRune('\n')
	}

	return sb.String()
}

func TestRequester(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctxEnd, ctxEndCancel := context.WithTimeout(ctx, 60*time.Second)
	defer ctxEndCancel()

	const prioSpread = 10

	var wg sync.WaitGroup

	prioRequesters := UnfairRequestersC(
		ctx,
		&wg,
		prioSpread,
		RequestConsumerPoolC(
			func(_ context.Context, _ struct{}) (struct{}, error) {
				return struct{}{}, nil
			}, prioSpread,
		),
	)

	c := newCCS(prioSpread)

	var wg2 sync.WaitGroup

	ConsumerPoolC(
		1000,
		func(_ context.Context, v struct{}) {
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := rnd.Intn(prioSpread)
			start := time.Now()
			_, _ = prioRequesters[idx](ctxEnd, struct{}{})
			c.refresh(idx, time.Since(start))
		},
	)(ctxEnd, &wg2, generateDummy(ctxEnd))

	wg2.Wait()

	cancel()
	wg.Wait()

	fmt.Println(c.String())
}

func TestRequestProcessor(t *testing.T) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requester := RequesterC(
		ctx,
		&wg,
		RequestConsumerPoolC(
			func(ctx context.Context, v int) (int, error) {
				return v, nil
			}, 32,
		),
	)

	for i := 0; i < 100_000; i++ {
		r, err := requester(ctx, i)
		require.NoError(t, err)
		require.Equal(t, i, r)
	}

	cancel()
	wg.Wait()
}
