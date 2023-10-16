package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/byte4ever/conch/internal/conch"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	inStream := make(chan string)

	var k atomic.Int64

	const outFmt = "%s-%d"

	dirty.BalanceC(
		10, //nolint:gomnd // dgas
		dirty.ProcessorsC(
			func(ctx context.Context, param string) (result string) {
				v := k.Add(1)
				return fmt.Sprintf(outFmt, param, v)
			},
			dirty.FanInC(
				dirty.ConsumerC(
					0,
					func(ctx context.Context, _ int, t string) {
						fmt.Println(t)
					},
				),
			),
		),
	)(ctx, &wg, inStream)

	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	close(inStream)
	wg.Wait()

	fmt.Println("------------------------------------")

	inStream = make(chan string)

	k.Store(0)

	dirty.BufferC(
		50, //nolint:gomnd // dgas
		dirty.BalanceC(
			32, //nolint:gomnd // dgas
			dirty.ProcessorsC(
				func(ctx context.Context, param string) (result string) {
					v := k.Add(1)
					return fmt.Sprintf(outFmt, param, v)
				},
				dirty.FanInReducerC(3,
					dirty.ConsumersC(
						func(ctx context.Context, _ int, t string) {
							fmt.Println(t)
						},
					),
				),
			),
		),
	)(ctx, &wg, inStream)

	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	inStream <- "bolos"
	close(inStream)
	wg.Wait()
}
