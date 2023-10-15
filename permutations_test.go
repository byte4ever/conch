package conch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermutationsOf(t *testing.T) {
	p := PermutationsOf(
		context.Background(),
		[]string{
			"3", "2", "1", "L",
		},
	)

	var r [][]string

	for k := range p {
		r = append(r, k)
	}

	require.Exactly(
		t,
		[][]string{
			{"3", "2", "1", "L"},
			{"2", "3", "1", "L"},
			{"1", "3", "2", "L"},
			{"3", "1", "2", "L"},
			{"2", "1", "3", "L"},
			{"1", "2", "3", "L"},
			{"L", "2", "3", "1"},
			{"2", "L", "3", "1"},
			{"3", "L", "2", "1"},
			{"L", "3", "2", "1"},
			{"2", "3", "L", "1"},
			{"3", "2", "L", "1"},
			{"3", "1", "L", "2"},
			{"1", "3", "L", "2"},
			{"L", "3", "1", "2"},
			{"3", "L", "1", "2"},
			{"1", "L", "3", "2"},
			{"L", "1", "3", "2"},
			{"L", "1", "2", "3"},
			{"1", "L", "2", "3"},
			{"2", "L", "1", "3"},
			{"L", "2", "1", "3"},
			{"1", "2", "L", "3"},
			{"2", "1", "L", "3"},
		},
		r)
}

//func perm(n, k *big.Int) *big.Int {
//	r := fact(n)
//	r.Div(r, fact(n.Sub(n, k)))
//	return r
//}

func TestPermutationsIndexes(t *testing.T) {
	p := PermutationsIndexes(
		context.Background(),
		4,
	)

	var r [][]int

	for k := range p {
		r = append(r, k)
	}

	require.Exactly(
		t,
		[][]int{
			{0, 1, 2, 3},
			{1, 0, 2, 3},
			{2, 0, 1, 3},
			{0, 2, 1, 3},
			{1, 2, 0, 3},
			{2, 1, 0, 3},
			{3, 1, 0, 2},
			{1, 3, 0, 2},
			{0, 3, 1, 2},
			{3, 0, 1, 2},
			{1, 0, 3, 2},
			{0, 1, 3, 2},
			{0, 2, 3, 1},
			{2, 0, 3, 1},
			{3, 0, 2, 1},
			{0, 3, 2, 1},
			{2, 3, 0, 1},
			{3, 2, 0, 1},
			{3, 2, 1, 0},
			{2, 3, 1, 0},
			{1, 3, 2, 0},
			{3, 1, 2, 0},
			{2, 1, 3, 0},
			{1, 2, 3, 0},
		},
		r,
	)
}
