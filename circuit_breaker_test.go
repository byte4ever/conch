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

/*func Test_circuitStateEngine(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	success, failure, state := circuitStateEngine(
		ctx,
		5,
		3,
		5*time.Second,
	)

	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	failure(ctx)
	fmt.Println("              - ", <-state)
	time.Sleep(5 * time.Second)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)
	success(ctx)
	fmt.Println("              - ", <-state)

}
*/
// ErrBadass represent an error where ....
var ErrBadass = errors.New("badass")

func TestBreaker(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	requester, reqStream := Requester[int, int](ctx)
	breaker := Breaker(
		ctx,
		reqStream,
		5,
		3,
		5*time.Second,
	)
	sync := SpawnRequestProcessorsPool(
		ctx, breaker,
		func(ctx context.Context, p2 int) (int, error) {
			if rand.Float64() < 0.2 {
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
				}
				time.Sleep(1 * time.Millisecond)
				i++
			}
		}
	}()

	sync()
}
