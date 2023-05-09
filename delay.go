package conch

import (
	"context"
	"time"
)

type DurationGenerator interface {
	Next() time.Duration
}

// Delay an output stream that replicate inputs stream elements with a delay
// constant. Provided duration generator defines the minimal time interval
// between each element emission.
//
// Please note it can be used as a time bases event generator  when pressure is
// low. Otherwise, it can be used to create pressure simulate processing for
// example.
func Delay[T any](
	ctx context.Context,
	durationGenerator DurationGenerator,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		timer := time.NewTimer(durationGenerator.Next())

		nextTS := time.Now().Add(durationGenerator.Next())

		for {
			select {
			case <-ctx.Done():
				return
			case e, more := <-inStream:
				if !more {
					return
				}

				now := time.Now()

				if nextTS.After(now) {
					timer.Reset(nextTS.Sub(now))
					select {
					case <-ctx.Done():
						if !timer.Stop() {
							<-timer.C
						}

						return
					case <-timer.C:
						nextTS = time.Now().Add(durationGenerator.Next())
						select {
						case <-ctx.Done():
							return
						case outStream <- e:
						}
					}
				} else {
					nextTS = time.Now().Add(durationGenerator.Next())
					select {
					case <-ctx.Done():
						return
					case outStream <- e:
					}
				}
			}
		}
	}()

	return outStream
}
