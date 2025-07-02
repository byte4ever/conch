package conch

import (
	"context"
	"fmt"
	"iter"
	"math/rand"
	"runtime"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestIterInjectorC(t *testing.T) {
	l := []int{1, 2, 3, 4, 5}

	ctx := context.Background()

	var wg sync.WaitGroup

	IterInjectorC(
		ctx,
		&wg,
		slices.Values(l),
		func(
			ctx context.Context,
			wg *sync.WaitGroup,
			inStream ...<-chan int,
		) {
			for i := range inStream[0] {
				fmt.Println(i)
			}
		},
	)

	wg.Wait()
}

func randIter(count int, provider PoolProvider[Payload]) iter.Seq[*Payload] {
	return func(yield func(*Payload) bool) {
		for i := 0; i < count; i++ {
			v := rand.Intn(10000000)
			n := provider.Get()
			n.V = v
			if !yield(n) {
				return
			}
		}
	}
}

func nowIter(count int, provider PoolProvider[NowItem]) iter.Seq[*NowItem] {
	return func(yield func(param *NowItem) bool) {
		for i := 0; i < count; i++ {
			n := provider.Get()
			n.Now = time.Now()
			if !yield(n) {
				return
			}
		}
	}
}

func CPUBurn(ctx context.Context, d time.Duration) {
	lctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	for {
		select {
		case <-lctx.Done():
			return
		default:
			for i := 0; i < 4000000; i++ {
			}
			runtime.Gosched()
		}
	}
}

type Payload struct {
	V int
}

func (p *Payload) Clean() {

}

func NewPayload() *Payload {
	return new(Payload)
}

/*
	func Test_bozo(t *testing.T) {
		// l := []int{1, 2, 3, 4, 5, 10, 32, 3144}

		ctx := context.Background()

		var wg sync.WaitGroup

		valueProvider := NewPool[Payload](100)

		IterInjectorC(
			ctx,
			&wg,
			randIter(
				20000,
				valueProvider,
			),
			MultiplexC(
				64,
				ProcessorsC(
					valueProvider,
					func(
						ctx context.Context,
						p *Payload,
						r *Payload,
					) {
						CPUBurn(
							ctx,
							time.Duration(500+rand.Intn(500))*time.Millisecond,
						)
						r.V = p.V * 2
					},
					MultiplexC(
						100,
						SinksC(
							valueProvider,
							func(t *Payload) {
								fmt.Println(t.V)
							},
						),
					),
				),
			),
		)

		wg.Wait()
	}
*/
type NowItem struct {
	Now      time.Time
	Duration time.Duration
}

func (p *NowItem) Clean() {
	p.Now = time.Time{}
	p.Duration = time.Duration(0)
}

func NewNowParam() *NowItem {
	return new(NowItem)
}

/*
func Test_bozo2(t *testing.T) {
	// l := []int{1, 2, 3, 4, 5, 10, 32, 3144}

	ctx := context.Background()

	var wg sync.WaitGroup

	valueProvider := NewPool[NowItem](1000)

	n := time.Now()
	IterInjectorC(
		ctx,
		&wg,
		nowIter(
			20_000_000,
			valueProvider,
		),
		MultiplexC(
			20,
			DecoratorsC(
				func(
					ctx context.Context,
					p *NowItem,
				) {
					p.Duration = time.Since(p.Now)
				},
				MultiplexC(20,
					SinksC(
						valueProvider,
						func(t *NowItem) {
							// fmt.Println(t.Duration.Nanoseconds())
						},
					),
				),
			),
		),
	)

	wg.Wait()
	fmt.Println(time.Since(n))
}
*/
