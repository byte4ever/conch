package conch

import (
	"context"
)

func Spread[V Integer, Payload any](
	ctx context.Context,
	inputStream <-chan IndexedInteger[V, Payload],
	count int,
) []<-chan Payload {
	outputStreams := make([]chan Payload, count)
	outputStreamsDir := make([]<-chan Payload, count)

	for i := 0; i < count; i++ {
		outputStreams[i] = make(chan Payload, count)
		outputStreamsDir[i] = outputStreams[i]
	}

	go func() {
		defer func() {
			for i := 0; i < count; i++ {
				close(outputStreams[i])
			}
		}()

		for v := range inputStream {
			select {
			case outputStreams[v.Index] <- v.Payload:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputStreamsDir
}
