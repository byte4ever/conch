package retrier

import (
	"context"
	"errors"
	"fmt"
	"math"
	mrand "math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLab(t *testing.T) {
	timeStep := goldenRatioValue
	rndPerc := 0.3333
	startDuration := 100 * time.Millisecond

	mrand.Float64()

	rnd := mrand.New(mrand.NewPCG(
		uint64(time.Now().UnixNano()),
		uint64(time.Now().UnixNano()),
	))

	var accum time.Duration

	for i := 0; i < 16; i++ {
		if i == 0 {
			fmt.Println(i, 0, startDuration)
			accum = startDuration
			continue
		}

		v := 1 + rndPerc*((rnd.Float64()*2)-1.0)
		r := v * math.Pow(timeStep, float64(i))
		w := time.Duration(r * float64(startDuration.Nanoseconds()))
		fmt.Println(i, accum, w)
		accum += w
	}
}

var ErrFakeError = errors.New("fake error")

type FakeElem[Param any, Result any] struct {
	p Param
	r Result
	e error
}

func (f *FakeElem[Param, Result]) GetParam() Param {
	return f.p
}

func (f *FakeElem[Param, Result]) SetValue(r Result) {
	f.r = r
}

func (f *FakeElem[Param, Result]) GetValue() Result {
	return f.r
}

func (f *FakeElem[Param, Result]) SetError(e error) {
	f.e = e
}

func (f *FakeElem[Param, Result]) GetError() error {
	return f.e
}

func TestMaxDuration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := New(
		WithRetryDelay(200*time.Millisecond),
		// WithMaxTotalDuration(5*time.Second),
		WithJitterFactor(0.123),
		WithMaxRetry(10),
		// WithNoJitter(),
		WithMaxRetryDelay(10*time.Second),
		// WithBackoffFactor(1.0),
		// WithMustRetryFunc(
		// 	func(err error) bool {
		// 		return errors.Is(err, ErrFakeError)
		// 	},
		// ),
	)

	require.NoError(t, err)
	require.NotNil(t, r)

	k := 0
	f := WrapProcessorFunc(
		r,
		func(
			ctx context.Context,
			elem *FakeElem[int, int],
		) {
			elem.SetError(WrapToRetryableError(ErrFakeError))
			k++
			fmt.Println("call", k)
		},
	)

	n := time.Now()

	e := &FakeElem[int, int]{
		p: 10,
	}

	f(ctx, e)

	fmt.Println(time.Since(n))

	require.ErrorIs(t, e.GetError(), ErrReachMaxDuration)
}

func TestRetrier_delayGenerator(t1 *testing.T) {
	r := Retrier{
		config: config{
			retryDelay:    ref(DefaultRetryDelay),
			jitterFactor:  ref(1 / 5.0),
			backoffFactor: ref(goldenRatioValue),
			maxRetryDelay: ref(15 * time.Second),
		},
	}

	k := 30
	for i, d := range r.delayGenerator() {
		fmt.Println(i, d)
		k--
		if k == 0 {
			break
		}
	}
}
