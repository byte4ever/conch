package conch

var defaultMultiplier64 = uint64(6364136223846793005)
var defaultStream64 = uint64(1442695040888963407)

func lcg(x uint64) (uint64, error) {
	res := x*defaultMultiplier64 + defaultStream64
	if res%2 == 0 {
		return 0, ErrMocked
	}
	return res, nil
}
