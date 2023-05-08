package conch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeStream(prefix string, count int) <-chan string {
	outStream := make(chan string, count)

	for i := 0; i < count; i++ {
		outStream <- fmt.Sprintf("%s-%d", prefix, i)
	}

	close(outStream)

	return outStream
}

func checkStreamContent[T any](
	t *testing.T,
	stream <-chan T,
	expected []T,
) {
	t.Helper()

	for i := 0; i < len(expected); i++ {
		idx := i

		require.Eventually(
			t,
			func() bool {

				require.Equalf(
					t,
					expected[idx],
					<-stream,
					"idx: %d",
					idx,
				)

				return true
			},
			10*time.Millisecond,
			1*time.Millisecond,
		)
	}

	require.Eventually(
		t,
		func() bool {
			_, more := <-stream
			return assert.False(t, more)
		},
		10*time.Millisecond,
		1*time.Millisecond,
	)
}

func TestDistribute(t *testing.T) {
	t.Parallel()

	t.Run(
		"single stream", func(t *testing.T) {
			t.Parallel()

			stream := fakeStream("test", 4)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test-0",
					"test-1",
					"test-2",
					"test-3",
				},
			)
		},
	)

	t.Run(
		"two streams same length", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 4)
			stream2 := fakeStream("test2", 4)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test1-1",
					"test2-1",
					"test1-2",
					"test2-2",
					"test1-3",
					"test2-3",
				},
			)
		},
	)

	t.Run(
		"two streams first longer", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 4)
			stream2 := fakeStream("test2", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test1-1",
					"test2-1",
					"test1-2",
					"test1-3",
				},
			)
		},
	)

	t.Run(
		"two streams second longer", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 2)
			stream2 := fakeStream("test2", 4)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test1-1",
					"test2-1",
					"test2-2",
					"test2-3",
				},
			)
		},
	)

	t.Run(
		"three streams ", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 1)
			stream2 := fakeStream("test2", 2)
			stream3 := fakeStream("test3", 3)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2, stream3)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test3-0",
					"test2-1",
					"test3-1",
					"test3-2",
				},
			)
		},
	)

	t.Run(
		"three streams case 1", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 1)
			stream2 := fakeStream("test2", 3)
			stream3 := fakeStream("test3", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2, stream3)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test3-0",
					"test2-1",
					"test3-1",
					"test2-2",
				},
			)
		},
	)

	t.Run(
		"three streams case 2", func(t *testing.T) {
			t.Parallel()

			stream1 := fakeStream("test1", 3)
			stream2 := fakeStream("test2", 1)
			stream3 := fakeStream("test3", 2)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Distribute(ctx, stream1, stream2, stream3)

			checkStreamContent(
				t,
				outStream,
				[]string{
					"test1-0",
					"test2-0",
					"test3-0",
					"test1-1",
					"test3-1",
					"test1-2",
				},
			)
		},
	)
}
