package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

// Combinations produces all combinations of elements from the given slice.
func Combinations[T any](
	ctx context.Context,
	elements []T,
	r int,
) <-chan []T {
	if r > len(elements) {
		panic("Combinations: r > len(elements)")
	}

	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		ori := copyList(elements)

		for indexes := range CombinationsIndexes(ctx, len(ori), r) {
			t := make([]T, r)

			for idx, index := range indexes {
				t[idx] = ori[index]
			}

			select {
			case outStream <- t:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outStream
}

func CombinationsC[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	elements []T,
	r int,
	chain dirty.ChainFunc[[]T],
) {
	chain(ctx, wg, Combinations(ctx, elements, r))
}

// CombinationsIndexes produces all combinations of indexes.
func CombinationsIndexes[T dirty.Integer](
	ctx context.Context,
	n, r T,
) <-chan []T {
	if r > n {
		panic("CombinationsIndexes: r > n")
	}

	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		indexes := make([]T, r)

		for i := range indexes {
			indexes[i] = T(i)
		}

		temp := make([]T, r)

		copy(temp, indexes)

		select {
		case <-ctx.Done():
			return
		case outStream <- temp:
		}

		for {
			for i := r - 1; i >= 0; i-- {
				if indexes[i] < i+n-r {
					indexes[i]++

					for j := T(1); j < r-i; j++ {
						indexes[i+j] = indexes[i] + j
					}

					temp := make([]T, r)

					copy(temp, indexes) // avoid overwriting of indexes

					select {
					case <-ctx.Done():
						return
					case outStream <- temp:
					}

					break
				}
			}

			if indexes[0] >= n-r {
				return
			}
		}
	}()

	return outStream
}

func CombinationsIndexesC[T dirty.Integer](
	ctx context.Context,
	wg *sync.WaitGroup,
	n T,
	r T,
	chain dirty.ChainFunc[[]T],
) {
	chain(ctx, wg, CombinationsIndexes(ctx, n, r))
}

// CombinationsIndexesWithReplacement produces all combinations of indexes with
// replacement.
func CombinationsIndexesWithReplacement[T dirty.Unsigned](
	ctx context.Context,
	n, r T,
) <-chan []T {
	if r > n {
		panic("r > n")
	}

	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		indexes := make([]T, r+1)

		for {
			for i := r; i > 0; i-- {
				if indexes[i] > n-1 {
					indexes[i-1]++
					for j := i; j <= r; j++ {
						indexes[j] = indexes[j-1]
					}
				}
			}

			if indexes[0] > 0 {
				break
			}

			temp := make([]T, r)
			copy(temp, indexes[1:])
			select {
			case <-ctx.Done():
				return
			case outStream <- temp:
			}

			indexes[r]++
		}
	}()

	return outStream
}

func CombinationsIndexesWithReplacementC[T dirty.Unsigned](
	ctx context.Context,
	wg *sync.WaitGroup,
	n T,
	r T,
	chain dirty.ChainFunc[[]T],
) {
	chain(ctx, wg, CombinationsIndexesWithReplacement(ctx, n, r))
}

// CombinationsWithReplacement produces all combinations of elements with
// replacement.
func CombinationsWithReplacement[T any](
	ctx context.Context,
	elements []T, r int,
) <-chan []T {
	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		ori := copyList(elements)

		for indexes := range CombinationsIndexesWithReplacement(
			ctx, uint(len(ori)), uint(r),
		) {
			t := make([]T, r)

			for idx, index := range indexes {
				t[idx] = ori[index]
			}

			select {
			case outStream <- t:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outStream
}
