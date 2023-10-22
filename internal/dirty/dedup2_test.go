package dirty

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

const (
	nbCallers = 1000
)

func Test_replicateReturnStream(t *testing.T) {
	t.Parallel()

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var (
		m     sync.Map
		start sync.WaitGroup
		all   sync.WaitGroup
	)

	tPool := newTrackerPool[int](1000)
	vePool := conch.newValErrorChanPool[int](1000)

	channels := make([]chan dirty.ValErrorPair[int], nbCallers)

	for i := 0; i < nbCallers; i++ {
		channels[i] = make(chan dirty.ValErrorPair[int], 1)
	}

	start.Add(1)

	expectedValue := dirty.ValErrorPair[int]{
		V: 1001,
	}

	var chOut chan<- dirty.ValErrorPair[int]

	all.Add(len(channels))

	for _, ch := range channels {
		go func(ch chan dirty.ValErrorPair[int]) {
			all.Done()
			start.Wait()

			if c := replicateReturnStream(
				tPool, vePool, Key2{}, &m, ch,
			); c != nil {
				chOut = c
			}
		}(ch)
	}

	require.Eventually(
		t,
		func() bool {
			all.Wait()
			return true
		},
		time.Second,
		100*time.Millisecond,
	)

	start.Done()

	require.Eventually(
		t,
		func() bool {
			chOut <- expectedValue
			return true
		},
		time.Second,
		100*time.Millisecond,
	)

	for idx, ch := range channels {
		go func(idx int, ch chan dirty.ValErrorPair[int]) {
			require.Eventually(
				t,
				func() bool {
					return assert.Equal(t, expectedValue, <-ch)
				},
				time.Second,
				10*time.Millisecond,
			)
		}(idx, ch)
	}

	require.Eventually(
		t,
		func() bool {
			var cnt int
			m.Range(
				func(_, _ any) bool {
					cnt++
					return true
				},
			)
			return 0 == cnt
		},
		time.Second,
		100*time.Millisecond,
	)

	t.Log(len(tPool))
	t.Log(len(vePool))
}

//go:noinline
func tTest(wg *sync.WaitGroup) {
	wg.Done()
}
