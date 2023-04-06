package conch

import (
	"context"
	"fmt"
)

type IndexGetter[O any] func(O) int

func Sieve[T any](
	ctx context.Context,
	choice IndexGetter[T],
	count int,
	inStream <-chan T,
) []<-chan T {
	outStreams := make([]chan T, count)
	// create output streams
	for i := 0; i < count; i++ {
		outStreams[i] = make(chan T)
	}

	go func() {
		defer func() {
			for i := 0; i < count; i++ {
				go func(c chan<- T) {
					close(c)
				}(outStreams[i])
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					return
				}

				select {
				case outStreams[choice(v)] <- v:

				case <-ctx.Done():
					return
				}
			}
		}

	}()

	outStreamsDir := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		outStreamsDir[i] = outStreams[i]
	}

	return outStreamsDir
}

type sieveItem[T any] struct {
	idx     int
	payload T
}

func choiceIdx[T any](
	ctx context.Context,
	idx int,
	inStream <-chan sieveItem[T],
	bufLeft int,
	bufRight int,
) (
	<-chan sieveItem[T],
	<-chan sieveItem[T],
) {
	fmt.Println("choiceIdx", idx)
	outStream1 := make(chan sieveItem[T])
	outStream2 := make(chan sieveItem[T])

	go func() {
		defer close(outStream1)
		defer close(outStream2)

		for {
			select {
			case <-ctx.Done():
				return

			case s, more := <-inStream:
				if !more {
					return
				}

				if s.idx >= idx {
					// fmt.Println(s.idx, idx)
					select {
					case outStream2 <- s:
					case <-ctx.Done():
						return
					}
					continue
				}

				// fmt.Println(s.idx, idx-1)
				select {
				case outStream1 <- s:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream1, outStream2
}

func recBuildTeeTreeIdx[T any](
	ctx context.Context,
	depth int,
	input <-chan sieveItem[T],
	in, out int,
	outputs []<-chan sieveItem[T],
) {
	ln := len(outputs)

	if in < 0 || out < 0 {
		panic("shit")
	}
	lo := out - in

	if lo == 2 {
		outputs[in], outputs[in+1] = choiceIdx(
			ctx,
			in+1,
			input,
			ln,
			ln,
		)
		return
	}

	if lo == 3 {
		var ot <-chan sieveItem[T]
		outputs[in], ot = choiceIdx(
			ctx,
			in+1,
			input,
			ln,
			ln-1,
		)
		outputs[in+1], outputs[in+2] = choiceIdx(
			ctx,
			in+2,
			ot,
			ln,
			ln,
		)
		return
	}

	mid := in + lo/2
	o1, o2 := choiceIdx(ctx, mid, input, ln-(mid-in), ln-(out-mid))
	depth++

	recBuildTeeTreeIdx(ctx, depth, o1, in, mid, outputs)
	recBuildTeeTreeIdx(ctx, depth, o2, mid, out, outputs)
}

func unpayloadOperator[T any](
	ctx context.Context,
	input <-chan sieveItem[T],
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-input:
				if !more {
					return
				}

				select {
				case outStream <- v.payload:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func convertStream[T any](
	ctx context.Context,
	choice IndexGetter[T],
	input <-chan T,
) <-chan sieveItem[T] {
	outStream := make(chan sieveItem[T])

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-input:
				if !more {
					return
				}

				select {
				case outStream <- sieveItem[T]{
					idx:     choice(v),
					payload: v,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func SieveTree[T any](
	ctx context.Context,
	choice IndexGetter[T],
	count int,
	inStream <-chan T,
) []<-chan T {
	outStreams := make([]<-chan sieveItem[T], count)

	injectSieve := convertStream(
		ctx,
		choice,
		inStream,
	)

	recBuildTeeTreeIdx(
		ctx,
		0,
		injectSieve,
		0,
		count,
		outStreams,
	)

	outStreamsCvt := make([]<-chan T, count)

	for i, outStream := range outStreams {
		outStreamsCvt[i] = unpayloadOperator(
			ctx,
			outStream,
		)
	}

	return outStreamsCvt
}
