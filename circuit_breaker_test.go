package conch

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// ErrBadass represent an error where ....
var ErrBadass = errors.New("badass")

// ErrUnavailable represent an error where....
var ErrUnavailable = errors.New("unavailable")

func TestBreaker(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	requester, reqStream := Requester[int, int](ctx)
	breaker := Breaker(
		ctx,
		reqStream,
		10,
		3,
		2*time.Second,
		ErrUnavailable,
	)
	sync := SpawnRequestProcessorsPool(
		ctx, breaker,
		func(ctx context.Context, p2 int) (int, error) {
			if rand.Float64() < 0.5 {
				return 0, ErrBadass
			}
			return p2, nil
		},
		10,
		"",
	)

	go func() {
		var i int

		for {
			select {
			case <-ctx.Done():
				return

			default:
				k, err := requester(ctx, i)
				if err == nil {
					fmt.Println(k)
				} else {
					fmt.Println(err)
				}

				time.Sleep(1 * time.Millisecond)

				i++
			}
		}
	}()

	sync()
}
