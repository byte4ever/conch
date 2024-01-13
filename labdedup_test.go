package conch

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKK_Subscribe(t *testing.T) {
	k := KK[PP, int]{
		m: make(map[Key]*Sub[int]),
	}

	ctx := context.Background()

	pool := newSubPool[int](100)

	const X = 10
	channels := make([]chan ValErrorPair[int], X)
	for i := 0; i < X; i++ {
		channels[i] = make(chan ValErrorPair[int])
	}

	var s chan ValErrorPair[int]

	for idx, chanel := range channels {
		if c := k.Subscribe(ctx, pool, 1234, chanel); c != nil {
			require.Zero(t, idx)
			s = c
			continue
		}

		require.NotZero(t, idx)
	}

	var wg sync.WaitGroup
	wg.Add(X + 1)
	go func() {
		defer wg.Done()
		s <- ValErrorPair[int]{
			V: 1000,
		}
	}()

	for idx, chanel := range channels {
		go func(c chan ValErrorPair[int], i int) {
			defer wg.Done()
			fmt.Println(i, <-c)
		}(chanel, idx)
	}

	wg.Wait()
}
