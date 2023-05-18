package conch

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCombinationsIndexes(t *testing.T) {
	const (
		n = 23
		r = 10
	)

	var cnt int64

	for range CombinationsIndexes(context.Background(), n, r) {
		cnt++
	}

	require.Equal(
		t, comb(
			big.NewInt(n),
			big.NewInt(r),
		).Int64(), cnt,
	)
}

func TestCombinationsIndexes2(t *testing.T) {
	const (
		n = 11
		r = 5
	)

	var cnt int64

	for range CombinationsIndexesWithReplacement(
		context.Background(),
		uint(n),
		uint(r),
	) {
		cnt++
	}

	fmt.Println(cnt)

	require.Equal(
		t, combWithRep(n, r).Int64(), cnt,
	)
}

func TestCombinationsIndexesWithReplacement(t *testing.T) {
	var l [][]uint

	for p := range CombinationsIndexesWithReplacement(
		context.Background(),
		uint(5),
		uint(3),
	) {
		l = append(l, p)
	}

	expected := [][]uint{
		{0, 0, 0},
		{0, 0, 1},
		{0, 0, 2},
		{0, 0, 3},
		{0, 0, 4},
		{0, 1, 1},
		{0, 1, 2},
		{0, 1, 3},
		{0, 1, 4},
		{0, 2, 2},
		{0, 2, 3},
		{0, 2, 4},
		{0, 3, 3},
		{0, 3, 4},
		{0, 4, 4},
		{1, 1, 1},
		{1, 1, 2},
		{1, 1, 3},
		{1, 1, 4},
		{1, 2, 2},
		{1, 2, 3},
		{1, 2, 4},
		{1, 3, 3},
		{1, 3, 4},
		{1, 4, 4},
		{2, 2, 2},
		{2, 2, 3},
		{2, 2, 4},
		{2, 3, 3},
		{2, 3, 4},
		{2, 4, 4},
		{3, 3, 3},
		{3, 3, 4},
		{3, 4, 4},
		{4, 4, 4},
	}

	require.Equal(t, expected, l)
}

func TestCombinations(t *testing.T) {
	var l [][]string

	for p := range Combinations(
		context.Background(),
		[]string{"a", "b", "c", "d", "e", "f"},
		3,
	) {
		l = append(l, p)
	}

	expected := [][]string{
		{"a", "b", "c"},
		{"a", "b", "d"},
		{"a", "b", "e"},
		{"a", "b", "f"},
		{"a", "c", "d"},
		{"a", "c", "e"},
		{"a", "c", "f"},
		{"a", "d", "e"},
		{"a", "d", "f"},
		{"a", "e", "f"},
		{"b", "c", "d"},
		{"b", "c", "e"},
		{"b", "c", "f"},
		{"b", "d", "e"},
		{"b", "d", "f"},
		{"b", "e", "f"},
		{"c", "d", "e"},
		{"c", "d", "f"},
		{"c", "e", "f"},
		{"d", "e", "f"},
	}

	require.Equal(t, expected, l)
}
