package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestChain(t *testing.T) {
	t.Run(
		"success", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 2)
			stream2 := fakeStream("s2", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			chained := Chain(ctx, &wg, stream1, stream2)
			checkStreamContent(
				t, chained, []string{
					"s1-0",
					"s1-1",
					"s2-0",
					"s2-1",
				},
			)

			wgWait(t, &wg, time.Second, time.Millisecond)
		},
	)

	t.Run(
		"detect duplicate", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 2)

			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			require.Panics(
				t, func() {
					_ = Chain(ctx, &wg, stream1, stream1)
				},
			)

			wgWait(t, &wg, time.Second, time.Millisecond)
		},
	)
}
