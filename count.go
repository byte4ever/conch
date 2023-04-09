package conch

func Count[T any](
	inStream <-chan T,
) (count <-chan int) {
	var c int

	iCount := make(chan int)

	go func() {
		defer close(iCount)

		for range inStream {
			c++
		}

		iCount <- c
	}()

	return iCount
}
