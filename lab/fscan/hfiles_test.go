package fscan

import (
	"context"
	"crypto"
	"fmt"
	"testing"

	"github.com/byte4ever/conch"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestPathGenerator(t *testing.T) {
	// t.SkipNow()
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outStream := PathGenerator(
		ctx,
		"/home/lmartin/Downloads",
		crypto.SHA256,
	)

	fhpb := &FileHashProcessorBuilder{
		InStream: outStream,
	}

	processor := fhpb.BuildProcessor()

	poolOutStream, err := conch.WorkerPool(
		ctx,

		processor,
		processor,
		processor,
		processor,
		processor,
		processor,
		processor,
		processor,
	)
	require.NoError(t, err)

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
