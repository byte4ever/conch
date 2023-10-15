package fscan

import (
	"context"
	"crypto"
	"fmt"
	"sync"
	"testing"

	"github.com/dustin/go-humanize"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

func TestPathGenerator(t *testing.T) {
	// t.SkipNow()
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	conch.BalanceC(18,
		conch.ProcessorsC(
			ProcessRequest,
			conch.FanInC(
				conch.ConsumerC(
					0,
					func(
						ctx context.Context,
						_ int,
						param *FileHashResponse,
					) {
						fmt.Println(
							param.Path,
							humanize.Bytes(uint64(param.Size)),
							// resp.Duration,
							// hex.EncodeToString(resp.Hash),
							// float64(resp.Size)/(resp.Duration.Seconds()*1024*1024),
						)
					},
				),
			),
		),
	)(ctx, &wg, PathGenerator(
		ctx,
		"/home/lmartin/Downloads",
		crypto.SHA256,
	),
	)

	wg.Wait()
}
