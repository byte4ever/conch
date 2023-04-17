package conch

import (
	"fmt"
	"time"
)

type LeakyBucket struct {
	current           float64
	lastUpdate        time.Time
	leakRatePerSecond float64
	max               float64
}

func NewLeakyBucket(
	count int,
	interval time.Duration,
) *LeakyBucket {
	leakRatePerSecond := -float64(count) / interval.Seconds()

	return &LeakyBucket{
		max:               float64(count),
		lastUpdate:        time.Now(),
		leakRatePerSecond: leakRatePerSecond,
	}
}

func (b *LeakyBucket) Incr() bool {
	// evaluate leak quantity
	now := time.Now()
	deltaTime := now.Sub(b.lastUpdate).Seconds()
	leak := b.leakRatePerSecond * deltaTime
	b.current += leak

	if b.current < 0 {
		b.current = 0
	}

	fmt.Println(b.max, b.current, leak)
	b.lastUpdate = now
	if b.current+1 > b.max {
		return true
	}

	b.current += 1

	return false
}
