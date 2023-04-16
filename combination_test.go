package conch

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

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
	t.Log(cnt, " combinations generated")
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

func TestCombinationsIndexesWithReplacement(t *testing.T) {
	var cnt int

	s := time.Now()

	for range CombinationsIndexesWithReplacement(
		context.Background(), 17, 8,
	) {
		cnt++
	}

	fmt.Println(time.Since(s), cnt)
}

func TestCombinations(t *testing.T) {
	for p := range Combinations(
		context.Background(),
		[]string{"a", "b", "c", "d", "e", "f"},
		3,
	) {
		fmt.Println(p)
	}
}

func p(x []int, k, n, r int, f func([]int), depth int) {
	fmt.Println(" - ", depth, k, x, n, r)
	// if depth == 3 {
	// 	return
	// }

	if r == 0 {
		f(x)
		return
	}

	for i := k; i < n-r+1; i++ {
		x[depth] = i
		p(x, i, n, r-1, f, depth+1)
	}
}

func Test2(t *testing.T) {

	// n := 7
	//
	// for i := 0; i < n-2; i++ {
	// 	for j := i; j < n-1; j++ {
	// 		for k := j; k < n; k++ {
	// 			fmt.Println(i, j, k)
	// 		}
	// 	}
	// }

	x := make([]int, 3)
	p(
		x, 0, 7, 3, func(x []int) {
			fmt.Println(x)
		}, 0,
	)
}

// idea: custom number system with 2s complement like 0b10...0==MIN stop case
// void combination_with_repetiton(int n, int k) {
//  while (1) {
//    for (int i = k; i > 0; i -= 1) {
//      if (pos[i] > n - 1) // if number spilled over: xx0(n-1)xx
//      {
//        pos[i - 1] += 1; // set xx1(n-1)xx
//        for (int j = i; j <= k; j += 1)
//          pos[j] = pos[j - 1]; // set xx11..1
//      }
//    }
//    if (pos[0] > 0) // stop condition: 1xxxx
//      break;
//    printDonuts(k);
//    pos[k] += 1; // xxxxN -> xxxxN+1
//  }
// }

// void comb(int m, int n, unsigned char *c)
// {
//	int i;
//	for (i = 0; i < n; i++) c[i] = n - i;
//
//	while (1) {
//		for (i = n; i--;)
//			printf("%d%c", c[i], i ? ' ': '\n');
//
//		/* this check is not strictly necessary, but if m is not close to n,
//		   it makes the whole thing quite a bit faster */
//		i = 0;
//		if (c[i]++ < m) continue;
//
//		for (; c[i] >= m - i;) if (++i >= n) return;
//		for (c[i]++; i; i--) c[i-1] = c[i] + 1;
//	}
// }

func combination_with_repetition(n, k uint8) {

	var cnt int
	defer func() {
		fmt.Println(cnt)
	}()

	indexes := make([]uint8, k+1)
	// pt := indexes[1:]
	for {
		for i := k; i > 0; i-- {
			if indexes[i] > n-1 {
				indexes[i-1]++
				for j := i; j <= k; j++ {
					indexes[j] = indexes[j-1]
				}
			}
		}

		if indexes[0] > 0 {
			break
		}

		cnt++
		// fmt.Printf("%8d ", cnt)

		// for x := uint8(1); x < k+1; x++ {
		// 	fmt.Printf("%d ", indexes[x])
		// }
		//
		// fmt.Println(
		// 	"",
		// )
		// fmt.Println(pt)
		indexes[k]++
	}
}

func TestShit(t *testing.T) {
	combination_with_repetition(8, 4)
	// combination(8, 3)
}
