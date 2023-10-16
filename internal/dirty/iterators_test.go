package dirty

import (
	"fmt"
	"testing"
	"time"
)

func TestGen_Next(t *testing.T) {
	g := Gen{
		n: 0.0,
		k: 1.61803398875,
		t: 10 * time.Millisecond,
		s: 0.61803398875,
		m: 250 * time.Millisecond,
	}

	var accum time.Duration

	for i := 0; i < 30; i++ {
		next := g.Next()
		accum += next
		fmt.Println(accum, next)
	}
}
