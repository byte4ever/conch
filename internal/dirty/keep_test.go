package dirty

import (
	"context"
	"testing"

	"go.uber.org/goleak"
)

func TestKeep(t *testing.T) {
	t.Run(
		"some", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 200)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			kept := Keep(ctx, stream1, 4)

			checkStreamContent(
				t, kept, []string{
					"s1-0",
					"s1-1",
					"s1-2",
					"s1-3",
				},
			)
		},
	)

	t.Run(
		"none", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			stream1 := fakeStream("s1", 200)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			kept := Keep(ctx, stream1, 0)

			checkStreamContent(
				t, kept, []string{},
			)
		},
	)
}
