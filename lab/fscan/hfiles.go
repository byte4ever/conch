package fscan

import (
	"context"
	"crypto"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/byte4ever/conch"
)

type FileHashRequest struct {
	Path   string
	Hasher crypto.Hash
}

type FileHashResponse struct {
	FileHashRequest
	Size     int64
	Duration time.Duration
	Hash     []byte
	Err      error
}

type FileHashProcessorBuilder struct {
	InStream chan *FileHashRequest
}

func (b *FileHashProcessorBuilder) BuildProcessor() conch.Generator[*FileHashResponse] {
	return func(ctx context.Context) (
		<-chan *FileHashResponse, error,
	) {
		outStream := make(chan *FileHashResponse)

		go func() {
			defer close(outStream)

			for {
				select {
				case <-ctx.Done():
					return
				case request, more := <-b.InStream:
					if !more {
						return
					}

					start := time.Now()
					hasher := request.Hasher.New()

					f, err := os.Open(request.Path)

					if err != nil {
						select {
						case outStream <- &FileHashResponse{
							FileHashRequest: *request,
							Err:             err,
						}:
						case <-ctx.Done():
							return
						}

						continue
					}

					size, err := io.Copy(hasher, f)
					if err != nil {
						outStream <- &FileHashResponse{
							FileHashRequest: *request,
							Err:             err,
						}

						_ = f.Close()

						continue
					}

					_ = f.Close()

					select {
					case outStream <- &FileHashResponse{
						FileHashRequest: *request,
						Duration:        time.Since(start),
						Size:            size,
						Hash:            hasher.Sum(nil),
					}:

					case <-ctx.Done():
						return
					}
				}
			}
		}()

		return outStream, nil
	}
}

func PathGenerator(
	ctx context.Context,
	path string,
	hasher crypto.Hash,
) chan *FileHashRequest {
	outStream := make(chan *FileHashRequest)

	f := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		select {
		case outStream <- &FileHashRequest{
			Path:   path,
			Hasher: hasher,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	}

	go func() {
		defer close(outStream)

		_ = filepath.Walk(path, f)
	}()

	return outStream
}
