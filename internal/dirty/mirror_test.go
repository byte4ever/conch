package dirty

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

func TestMirrorHighThroughputC(t *testing.T) {
	goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	inStream := make(chan int)

	res := make([][]int, 5)

	MirrorHighThroughputC(
		5,
		ConsumersC(
			func(ctx context.Context, id int, t int) {
				res[id] = append(res[id], t)
			}),
	)(
		ctx,
		&wg,
		inStream,
	)

	inStream <- 10
	inStream <- 11
	inStream <- 12

	close(inStream)
	dirty.wgWait(t, &wg, time.Second, time.Millisecond)

	require.Exactly(
		t,
		[][]int{
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
		},
		res,
	)
}

func TestMirrorLowLatencyC(t *testing.T) {
	goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	inStream := make(chan int)

	res := make([][]int, 5)

	MirrorLowLatencyC(
		5,
		ConsumersC(
			func(ctx context.Context, id int, t int) {
				res[id] = append(res[id], t)
			}),
	)(
		ctx,
		&wg,
		inStream,
	)

	inStream <- 10
	inStream <- 11
	inStream <- 12

	close(inStream)
	dirty.wgWait(t, &wg, time.Second, time.Millisecond)

	require.Exactly(
		t,
		[][]int{
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
			{10, 11, 12},
		},
		res,
	)
}
