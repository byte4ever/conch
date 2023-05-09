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

	conch.ProcessorPoolC(
		64,
		conch.GetProcessorFor(ProcessRequest),
		conch.Consumer(
			func(ctx context.Context, resp *FileHashResponse) {
				fmt.Println(
					resp.Path,
					humanize.Bytes(uint64(resp.Size)),
					// resp.Duration,
					// hex.EncodeToString(resp.Hash),
					// float64(resp.Size)/(resp.Duration.Seconds()*1024*1024),
				)
			},
		),
	)(
		ctx, &wg, PathGenerator(
			ctx,
			"/home/lmartin/Downloads",
			crypto.SHA256,
		),
	)

	wg.Wait()
}
