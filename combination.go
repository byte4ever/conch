package conch

import (
	"context"
	"fmt"
)

func Combinations[T any](
	ctx context.Context,
	elements []T, r int,
) <-chan []T {
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

func CombinationsIndexes(
	ctx context.Context,
	n, r int,
) <-chan []int {
	if r > n {
		panic("r > n")
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

func CombinationsIndexesWithReplacement(
	ctx context.Context,
	n, r int,
) <-chan []int {
	if r > n {
		panic("r > n")
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

					fmt.Println(">", indexes)
					for j := 1; j < r-i; j++ {
						indexes[i+j] = indexes[i] + j
					}
					fmt.Println("<", indexes)

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
