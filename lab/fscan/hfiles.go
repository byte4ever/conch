package fscan

import (
	"context"
	"crypto"
	"io"
	"os"
	"path/filepath"
	"time"
)

type (
	FileHashRequest struct {
		Path   string
		Hasher crypto.Hash
	}
	FileHashResponse struct {
		FileHashRequest
		Size     int64
		Duration time.Duration
		Hash     []byte
		Err      error
	}
)

func HashFilesProcessor(
	ctx context.Context,
	inStream <-chan *FileHashRequest,
) <-chan *FileHashResponse {
	outStream := make(chan *FileHashResponse)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case request, more := <-inStream:
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
					select {
					case outStream <- &FileHashResponse{
						FileHashRequest: *request,
						Err:             err,
					}:
						_ = f.Close()
					case <-ctx.Done():
						_ = f.Close()
						return
					}

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

	return outStream
}

func PathGenerator(
	ctx context.Context,
	path string,
	hasher crypto.Hash,
) <-chan *FileHashRequest {
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

func ProcessRequest(
	_ context.Context,
	request *FileHashRequest,
) *FileHashResponse {
	start := time.Now()
	hasher := request.Hasher.New()

	f, err := os.Open(request.Path)
	if err != nil {
		return &FileHashResponse{
			FileHashRequest: *request,
			Err:             err,
		}
	}

	size, err := io.Copy(hasher, f)
	if err != nil {
		return &FileHashResponse{
			FileHashRequest: *request,
			Err:             err,
		}
	}

	_ = f.Close()

	return &FileHashResponse{
		FileHashRequest: *request,
		Duration:        time.Since(start),
		Size:            size,
		Hash:            hasher.Sum(nil),
	}
}
