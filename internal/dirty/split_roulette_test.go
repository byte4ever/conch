package dirty

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRoulette(t *testing.T) {
	coefficients := []float64{
		1.0,
		3.0,
		10.0,
	}
	lc := len(coefficients)

	r, _ := newRoulette(
		coefficients,
	)

	rnd := rand.New(rand.NewSource(1))

	const nbTests = 2_000_000

	stats := make([]float64, lc)
	for i := 0; i < nbTests; i++ {
		stats[r.Pick(rnd)]++
	}

	for i := range stats {
		stats[i] /= nbTests
	}

	var sum float64

	for _, v := range coefficients {
		sum += v
	}

	for i := range coefficients {
		coefficients[i] /= sum
	}

	var dist float64

	for i := 0; i < lc; i++ {
		pp := coefficients[i] - stats[i]
		dist += pp * pp
	}

	require.GreaterOrEqual(
		t,
		0.001,
		math.Sqrt(dist),
	)
}
