package conch

import (
	"context"
	"fmt"
	"math/big"
	"testing"
)

func TestPermutationsOf(t *testing.T) {
	p := PermutationsOf(
		context.Background(), []string{
			"3", "2", "1", "L",
		},
	)

	var cnt uint64

	for k := range p {
		fmt.Println(cnt, k)
		cnt++
	}

	fmt.Println(cnt)
}

func perm(n, k *big.Int) *big.Int {
	r := fact(n)
	r.Div(r, fact(n.Sub(n, k)))
	return r
}

func TestPermutationsIndexes(t *testing.T) {

	p := PermutationsIndexes(
		context.Background(),
		7,
	)

	var cnt uint64

	for k := range p {
		fmt.Println(cnt, k)
		cnt++
	}

	fmt.Println(cnt)
}
