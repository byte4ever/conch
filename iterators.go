package conch

import (
	"math"
	"math/rand"
	"time"
)

type Gen struct {
	n float64
	// initialDuration time.Duration
	k float64
	t time.Duration
	s float64
	m time.Duration
}

func (g *Gen) Next() time.Duration {
	spread := rand.Float64() * g.s
	pp := (math.Pow(g.k, g.n) + spread) * float64(time.Millisecond)

	// fmt.Println(spread)

	d := time.Duration(pp) * (g.t / time.Millisecond)
	if g.m != 0 {
		if d > g.m {
			return d
		}
	}

	g.n += 1.0

	return d
}
