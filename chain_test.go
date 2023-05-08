package conch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	t.Parallel()

	t.Run(
		"success", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("s1", 2)
			stream2 := fakeStream("s2", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			chained := Chain(ctx, stream1, stream2)
			checkStreamContent(
				t, chained, []string{
					"s1-0",
					"s1-1",
					"s2-0",
					"s2-1",
				},
			)
		},
	)

	t.Run(
		"detect duplicate", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("s1", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			require.Panics(
				t, func() {
					_ = Chain(ctx, stream1, stream1)
				},
			)
		},
	)

}
