package conch

import (
	"context"
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

// CombinationsIndexes produces all combinations of indexes.
func CombinationsIndexes(
	ctx context.Context,
	n, r int,
) <-chan []int {
	if r > n {
		panic("CombinationsIndexes: r > n")
	}

	outStream := make(chan []int)

	go func() {
		defer close(outStream)

		indexes := make([]int, r)

		for i := range indexes {
			indexes[i] = i
		}

		temp := make([]int, r)

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

					for j := 1; j < r-i; j++ {
						indexes[i+j] = indexes[i] + j
					}

					temp := make([]int, r)

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

// CombinationsIndexesWithReplacement produces all combinations of indexes with
// replacement.
func CombinationsIndexesWithReplacement(
	ctx context.Context,
	n, r uint8,
) <-chan []uint8 {
	if r > n {
		panic("r > n")
	}

	outStream := make(chan []uint8)

	go func() {
		defer close(outStream)

		indexes := make([]uint8, r+1)

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

			temp := make([]uint8, r)
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

// CombinationsWithReplacement produces all combinations of elements with
// replacement.
func CombinationsWithReplacement[T any](
	ctx context.Context,
	elements []T, r uint8,
) <-chan []T {
	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		ori := copyList(elements)

		for indexes := range CombinationsIndexesWithReplacement(
			ctx, uint8(len(ori)), r,
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
