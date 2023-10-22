package dirty

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/byte4ever/conch"
)

func Test_tracker(t *testing.T) {
	var m sync.Map

	pool := conch.newValErrorChanPool[int](10)
	trackerPool := newTrackerPool[int](10)

	const testCnt = 1

	rng1 := rand.New(rand.NewSource(time.Now().UnixNano() + 1))
	rng2 := rand.New(rand.NewSource(time.Now().UnixNano() + 2))
	rng3 := rand.New(rand.NewSource(time.Now().UnixNano() + 3))

	for i := 0; i < testCnt; i++ {
		trk := trackerPool.get()

		c1 := make(chan dirty.ValErrorPair[int], 1)
		c2 := make(chan dirty.ValErrorPair[int], 1)

		d1 := time.Duration(1+rng1.Intn(98)) * time.Millisecond
		d2 := time.Duration(1+rng2.Intn(98)) * time.Millisecond
		d3 := time.Duration(1+rng3.Intn(98)) * time.Millisecond

		t.Logf("#%d starting tracker %v %v %v", i, d1, d2, d3)

		cOut := pool.get()

		expectedValue := dirty.ValErrorPair[int]{
			V: 11,
		}

		go func() {
			time.Sleep(d1)
			trk.copyFromNewReturnStream(c2, cOut, Key2{}, &m, pool)
		}()

		go func() {
			time.Sleep(d2)
			trk.replicateValueIn(c1)
		}()

		go func() {
			time.Sleep(d3)
			cOut <- expectedValue
		}()

		require.Eventually(
			t, func() bool {
				return assert.Equal(t, expectedValue, <-c2)
			}, time.Second, time.Millisecond,
		)

		require.Eventually(
			t, func() bool {
				return assert.Equal(t, expectedValue, <-c1)
			}, time.Second, time.Millisecond,
		)
	}
}
