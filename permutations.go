package conch

import (
	"context"
)

func copyList[T any](elements []T) []T {
	result := make([]T, len(elements))
	copy(result, elements)
	return result
}

// PermutationsOf provides all permutations of a list of elements.
func PermutationsOf[T any](
	ctx context.Context,
	elements []T,
) <-chan []T {
	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		ori := copyList(elements)

		lo := len(ori)
		for indexes := range PermutationsIndexes(ctx, lo) {
			t := make([]T, lo)

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

// PermutationsIndexes provides all permutations indexes.
func PermutationsIndexes(
	ctx context.Context,
	n int,
) <-chan []int {
	outStream := make(chan []int)

	go func() {
		defer close(outStream)

		ori := make([]int, n)
		for i := 0; i < n; i++ {
			ori[i] = i
		}

		select {
		case <-ctx.Done():
			return
		case outStream <- copyList(ori):
		}

		c := make([]int, n)

		var i int
		for i < n {
			if c[i] < i {
				var k int

				if i%2 != 0 {
					k = c[i]
				}

				ori[k], ori[i] = ori[i], ori[k]

				select {
				case <-ctx.Done():
					return
				case outStream <- copyList(ori):
				}

				c[i]++
				i = 0

				continue
			}

			c[i] = 0
			i++
		}
	}()

	return outStream
}
