package conch

import (
	"context"
	"testing"

	"go.uber.org/goleak"
)

func TestSkip(t *testing.T) {
	t.Run(
		"some", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 6)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			skipped := Skip(ctx, stream1, 2)

			checkStreamContent(
				t, skipped, []string{
					"s1-2",
					"s1-3",
					"s1-4",
					"s1-5",
				},
			)
		},
	)

	t.Run(
		"none", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 6)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			skipped := Skip(ctx, stream1, 100)

			checkStreamContent(
				t, skipped, []string{},
			)
		},
	)
}
