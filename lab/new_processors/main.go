package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/byte4ever/conch"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	inStream := make(chan string)

	var k atomic.Int64

	const outFmt = "%s-%d"

	conch.BalanceC(
		10, //nolint:gomnd // dgas
		conch.ProcessorsC(
			func(ctx context.Context, param string) (result string) {
				v := k.Add(1)
				return fmt.Sprintf(outFmt, param, v)
			},
			conch.FanInC(
				conch.ConsumerC(
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

	conch.BufferC(
		50, //nolint:gomnd // dgas
		conch.BalanceC(
			32, //nolint:gomnd // dgas
			conch.ProcessorsC(
				func(ctx context.Context, param string) (result string) {
					v := k.Add(1)
					return fmt.Sprintf(outFmt, param, v)
				},
				conch.FanInReducerC(3,
					conch.ConsumersC(
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
