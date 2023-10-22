package conch

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func fakeStream2[T any](f func(int) T, count int) <-chan T {
	outStream := make(chan T, count)

	for i := 0; i < count; i++ {
		outStream <- f(i)
	}

	close(outStream)

	return outStream
}

func fact(n *big.Int) *big.Int {
	if n.Sign() < 1 {
		return big.NewInt(0)
	}

	r := big.NewInt(1)

	i := big.NewInt(2)

	for i.Cmp(n) < 1 {
		r.Mul(r, i)
		i.Add(i, big.NewInt(1))
	}

	return r
}

func comb(n, r *big.Int) *big.Int {
	if r.Cmp(n) == 1 {
		return big.NewInt(0)
	}

	if r.Cmp(n) == 0 {
		return big.NewInt(1)
	}

	c := fact(n)

	den := fact(n.Sub(n, r))
	den.Mul(den, fact(r))
	c.Div(c, den)

	return c
}

// combWithRep computes the combination with replacement of n and r
func combWithRep(n, r int64) *big.Int {
	// Compute (n+r-1)! / (r! * (n-1)!)
	numerator := big.NewInt(0).MulRange(1, n+r-1)
	p1 := big.NewInt(0).MulRange(1, r)
	p2 := big.NewInt(0).MulRange(1, n-1)
	m := p1.Mul(p1, p2)

	return numerator.Div(numerator, m)
}

func Test_comWithRep(t *testing.T) {
	require.Equal(t, uint64(3003), combWithRep(11, 5).Uint64())
	require.Equal(t, uint64(220), combWithRep(10, 3).Uint64())
	require.Equal(t, uint64(13037895), combWithRep(17, 11).Uint64())
}

var ErrMocked = errors.New("mocked")
