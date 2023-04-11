package conch

import (
	"context"
	"fmt"
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
