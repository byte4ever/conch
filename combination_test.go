package conch

import (
	"context"
	"fmt"
	"testing"
)

func TestCombinationsIndexes(t *testing.T) {
	for p := range CombinationsIndexes(context.Background(), 6, 3) {
		fmt.Println(p)
	}
}

func TestCombinationsIndexesWithReplacement(t *testing.T) {
	for p := range CombinationsIndexesWithReplacement(
		context.Background(), 6, 3,
	) {
		fmt.Println(p)
	}
}

func TestCombinations(t *testing.T) {
	for p := range Combinations(
		context.Background(),
		[]string{"a", "b", "c", "d", "e", "f"},
		3,
	) {
		fmt.Println(p)
	}
}

func p(x []int, k, n, r int, f func([]int), depth int) {
	fmt.Println(" - ", depth, k, x, n, r)
	// if depth == 3 {
	// 	return
	// }

	if r == 0 {
		f(x)
		return
	}

	for i := k; i < n-r+1; i++ {
		x[depth] = i
		p(x, i, n, r-1, f, depth+1)
	}
}

func Test2(t *testing.T) {

	// n := 7
	//
	// for i := 0; i < n-2; i++ {
	// 	for j := i; j < n-1; j++ {
	// 		for k := j; k < n; k++ {
	// 			fmt.Println(i, j, k)
	// 		}
	// 	}
	// }

	x := make([]int, 3)
	p(
		x, 0, 7, 3, func(x []int) {
			fmt.Println(x)
		}, 0,
	)

}
