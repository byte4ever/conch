package dirty

import (
	"fmt"
	"testing"
	"time"
)

func TestLeakyBucket_Incr(t *testing.T) {
	lb := NewLeakyBucket(10, time.Second)

	for i := 0; i < 1000; i++ {
		fmt.Println(lb.Incr())
		time.Sleep(time.Second / 11)
	}
}
