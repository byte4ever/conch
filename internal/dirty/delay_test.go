package dirty

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dummyGenerator time.Duration

func (d dummyGenerator) Next() time.Duration {
	return time.Duration(d)
}

func TestDelay(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := genStream(
		func(i int) int {
			return i
		}, 15,
	)

	delay := 50 * time.Millisecond
	out := Delay(ctx, dummyGenerator(delay), s)

	prev := time.Now()

	for range out {
		now := time.Now()
		require.InDelta(
			t, delay.Seconds(), now.Sub(prev).Seconds(),
			0.001,
		)

		prev = now
	}
}
