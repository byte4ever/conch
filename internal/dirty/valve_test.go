package dirty

/*func TestOpenedValve(t *testing.T) {
defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

t.Parallel()

t.Run(
	"stream are connected", func(t *testing.T) {
		t.Parallel()
		const nbTest = 100

		for i := 0; i < nbTest; i++ {
			v := valve.New()
			t.Log("running test", i)
			ctx, cancel := context.WithTimeout(
				context.Background(),
				200*time.Millisecond,
			)
			defer cancel()

			g := make(chan time.Time)
			outStream := Valve(ctx, v.GetChannel, true, g)

			var sg sync.WaitGroup
			sg.Add(3)

			go func() {
				defer sg.Done()
				defer fmt.Println("switcher off")
				timer := time.NewTimer(
					time.Duration(rand.Int63n(333)) *
						time.Microsecond,
				)
				isOpen := true
				for {
					select {
					case <-timer.C:
						timer.Reset(
							time.Duration(rand.Int63n(333)) *
								time.Microsecond,
						)
						if isOpen {
							isOpen = false

							v.Close()
							fmt.Println("closed")
							continue
						}
						isOpen = true
						fmt.Println("open")
						v.Open()
						continue
					case <-ctx.Done():
						if !timer.Stop() {
							<-timer.C
						}
						return
					}
				}
			}()

			go func() {
				defer sg.Done()
				defer fmt.Println("generator off")
				for {
					select {
					case <-ctx.Done():
						return
					case g <- time.Now():
					}
				}
			}()

			go func() {
				defer sg.Done()
				defer fmt.Println("reader off")
				var received int
				for range outStream {
					received++
				}
				require.NotZero(t, received)
			}()

			close(g)

			require.Eventually(
				t, func() bool {
					sg.Wait()
					return true
				},
				2000*time.Millisecond,
				1*time.Millisecond,
			)
		}
	},
)

/*
	t.Run(
		"close input no value", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int)
			_, _, outStream := Valve(ctx, g, true)

			close(g)
			isClosing(t, outStream, time.Second, time.Millisecond)
		},
	)

	t.Run(
		"cancel ctx no val", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int)
			_, _, outStream := Valve(ctx, g, true)

			cancel()
			isClosing(t, outStream, time.Second, time.Millisecond)
		},
	)

	t.Run(
		"open multiple time", func(t *testing.T) {
			t.Parallel()

			const (
				testValue = 10101
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int, 1)
			g <- testValue
			openIt, _, outStream := Valve(ctx, g, true)

			for i := 0; i < 3; i++ {
				t.Log(i)
				require.Eventually(
					t, func() bool {
						openIt()
						return true
					},
					1000*time.Millisecond,
					1*time.Millisecond,
				)
			}

			require.Eventually(
				t, func() bool {
					require.Equal(t, testValue, <-outStream)
					return true
				},
				100*time.Millisecond,
				1*time.Millisecond,
			)
		},
	)

	t.Run(
		"open multiple time after cancel ctx", func(t *testing.T) {
			t.Parallel()

			const (
				testValue = 10101
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int, 1)
			g <- testValue
			openIt, _, outStream := Valve(ctx, g, true)

			require.Eventually(
				t, func() bool {
					require.Equal(t, testValue, <-outStream)
					return true
				},
				100*time.Millisecond,
				10*time.Millisecond,
			)

			cancel()

			isClosing(t, outStream, time.Second, time.Millisecond)

			for i := 0; i < 3; i++ {
				require.Eventually(
					t, func() bool {
						openIt()
						return true
					},
					1000*time.Millisecond,
					1*time.Millisecond,
				)
			}
		},
	)

	t.Run(
		"multiple calls to openIt after close input", func(t *testing.T) {
			t.Parallel()

			const (
				testValue = 10101
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int, 1)
			g <- testValue
			openIt, _, outStream := Valve(ctx, g, true)

			require.Eventually(
				t, func() bool {
					require.Equal(t, testValue, <-outStream)
					return true
				},
				100*time.Millisecond,
				10*time.Millisecond,
			)

			close(g)
			isClosing(t, g, time.Second, time.Millisecond)

			for i := 0; i < 3; i++ {
				require.Eventually(
					t, func() bool {
						openIt()
						return true
					},
					1000*time.Millisecond,
					1*time.Millisecond,
				)
			}
		},
	)

	t.Run(
		"close input before creation", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int)
			close(g)
			_, _, outStream := Valve(ctx, g, true)
			isClosing(t, outStream, time.Second, time.Millisecond)
		},
	)

	t.Run(
		"cancel ctx before creation", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cancel()
			g := make(chan int)
			_, _, outStream := Valve(ctx, g, true)
			isClosing(t, outStream, time.Second, time.Millisecond)
		},
	)

	t.Run(
		"cancel ctx value not read", func(t *testing.T) {
			t.Parallel()

			const (
				testValue = 10101
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			g := make(chan int, 1)
			g <- testValue
			_, _, outStream := Valve(ctx, g, true)
			time.Sleep(10 * time.Millisecond)
			cancel()
			isClosed(t, outStream)
		},
	)*/
//}
//*/
