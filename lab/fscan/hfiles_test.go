package fscan

import (
	"context"
	"crypto"
	"fmt"
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

	poolOutStream := conch.ProcessorPool(
		ctx,
		16,
		conch.GetProcessorFor(ProcessRequest),
		PathGenerator(
			ctx,
			"/home/lmartin/Downloads",
			crypto.SHA256,
		),
	)

	for resp := range poolOutStream {
		fmt.Println(
			resp.Path,
			humanize.Bytes(uint64(resp.Size)),
			// resp.Duration,
			// hex.EncodeToString(resp.Hash),
			// float64(resp.Size)/(resp.Duration.Seconds()*1024*1024),
		)
	}
}
