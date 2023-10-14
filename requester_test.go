package conch

/*
func TestRequesterC(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			pp := NewMockRequestFunc[int, int](t)

			requester := RequesterC(
				ctx,
				&wg,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream <-chan Request[int, int],
				) {
					wg.Add(1)

					go func() {
						defer wg.Done()

						for {
							select {
							case <-ctx.Done():
								return
							case r, more := <-inStream:
								fmt.Printf("%#v\n", r)
								if !more {
									return
								}

								r.NakedReturn(
									func(
										pctx context.Context,
									) ValErrorPair[int] {
										select {
										case <-ctx.Done():
											return ValErrorPair[int]{
												Err: ctx.Err(),
											}
										default:
											res, err := pp.Execute(pctx, r.P)
											return ValErrorPair[int]{
												V:   res,
												Err: err,
											}
										}
									},
								)
							}
						}
					}()
				},
			)

			{
				pp.
					On("Execute", ctx, 1).
					Return(100, nil).Once()

				r, err := requester(ctx, 1)

				require.NoError(t, err)
				require.Equal(t, 100, r)
			}

			{
				pp.
					On("Execute", ctx, 2).
					Return(200, nil).Once()

				r, err := requester(ctx, 2)

				require.NoError(t, err)
				require.Equal(t, 200, r)
			}

			{
				pp.
					On("Execute", ctx, 2).
					Return(0, ErrMocked).Once()

				r, err := requester(ctx, 2)

				require.Equal(t, 0, r)
				require.ErrorIs(t, err, ErrMocked)
			}

			{
				ctx2, cancel2 := context.WithCancel(context.Background())
				defer cancel2()

				cancel2()
				r, err := requester(ctx2, 2)

				require.Equal(t, 0, r)
				require.ErrorIs(t, err, context.Canceled)
			}

			cancel()

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			{
				r, err := requester(ctx, 2)

				require.Equal(t, 0, r)
				require.ErrorIs(t, err, context.Canceled)
			}

			{
				ctx2, cancel2 := context.WithCancel(context.Background())
				defer cancel2()

				cancel2()
				r, err := requester(ctx2, 2)

				require.Equal(t, 0, r)
				require.ErrorIs(t, err, context.Canceled)
			}

		},
	)

	t.Run(
		"waiting request are properly drained", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			requester := RequesterC(
				ctx,
				&wg,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream <-chan Request[int, int],
				) {
					wg.Add(1)

					go func() {
						defer wg.Done()

						for {
							select {
							case <-ctx.Done():
								return
							case r, more := <-inStream:
								if !more {
									return
								}

								r.NakedReturn(
									func(
										pctx context.Context,
									) ValErrorPair[int] {
										select {
										case <-ctx.Done():
											return ValErrorPair[int]{
												Err: ctx.Err(),
											}
											// default:
											// 	res, err := pp.Execute(pctx, r.P)
											// 	return ValErrorPair[int]{
											// 		V:   res,
											// 		Err: err,
											// 	}
										}
									},
								)
							}
						}
					}()
				},
			)

			const ConcurrentRequests = 1
			var wg2 sync.WaitGroup
			wg2.Add(ConcurrentRequests)
			for i := 0; i < ConcurrentRequests; i++ {
				go func(i int) {
					defer wg2.Done()
					r, err := requester(ctx, 1)
					fmt.Println(r, err)
					require.Equal(t, 0, r)
					require.ErrorIs(t, err, context.Canceled)
				}(i)
			}

			time.Sleep(time.Millisecond * 500)

			cancel()

			require.Eventually(
				t, func() bool {
					wg2.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"torture test", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			requester := RequesterC(
				ctx,
				&wg,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream <-chan Request[int, int],
				) {
					wg.Add(1)
					go func() {
						defer wg.Done()

						for {
							select {
							case <-ctx.Done():
								return
							case r, more := <-inStream:
								if !more {
									return
								}

								r.NakedReturn(
									func(
										pctx context.Context,
									) ValErrorPair[int] {
										select {
										case <-ctx.Done():
											return ValErrorPair[int]{
												Err: ctx.Err(),
											}
										case <-pctx.Done():
											return ValErrorPair[int]{
												Err: ctx.Err(),
											}
										default:
											time.Sleep(20 * time.Millisecond)
											return ValErrorPair[int]{
												V: r.P*11 + 17,
											}
										}
									},
								)
							}
						}
					}()
				},
			)

			const ConcurrentRequests = 1000
			var wg2 sync.WaitGroup
			wg2.Add(ConcurrentRequests)
			for i := 0; i < ConcurrentRequests; i++ {
				go func(i, startSleep, beforeCall, timeout int) {
					defer wg2.Done()
					time.Sleep(time.Duration(startSleep) * time.Millisecond)
					ctx, cancel := context.WithTimeout(
						context.Background(),
						time.Duration(timeout)*time.Millisecond,
					)
					defer cancel()
					time.Sleep(time.Duration(beforeCall) * time.Millisecond)
					s := time.Now()
					r, err := requester(ctx, i)
					dd := time.Since(s)
					if err == nil {
						fmt.Println(i, r, dd)
						require.Equal(t, i*11+17, r)
					}
				}(
					i,
					rand.Intn(1000),
					rand.Intn(1000),
					rand.Intn(500),
				)
			}

			time.Sleep(time.Millisecond * 500)
			cancel()

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			require.Eventually(
				t, func() bool {
					wg2.Wait()
					return true
				}, 2500*time.Millisecond, time.Millisecond,
			)

		},
	)

	/*	t.Run(
			"cancel requester func ctx", func(t *testing.T) {
				t.Parallel()

				defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
				var wg sync.WaitGroup

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				pp := NewMockRequestFunc[int, int](t)

				requester := RequesterC(
					ctx,
					&wg,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStream <-chan Request[int, int],
					) {
						wg.Add(1)

						go func() {
							defer wg.Done()

							for {
								select {
								case <-ctx.Done():
									return
								case r, more := <-inStream:
									if !more {
										return
									}

									res, err := pp.Execute(ctx, r.P)
									r.Return(
										ctx, ValErrorPair[int]{
											V:   res,
											Err: err,
										},
									)
								}
							}
						}()
					},
				)

				{
					ctx2, cancel2 := context.WithCancel(context.Background())
					defer cancel2()

					cancel2()
					r, err := requester(ctx2, 2)

					require.Equal(t, 0, r)
					require.ErrorIs(t, err, context.Canceled)
				}

				cancel()

				require.Eventually(
					t, func() bool {
						wg.Wait()
						return true
					}, time.Second, time.Millisecond,
				)

			},
		)
*/

/*	t.Run(
			"cancel outer requester func ctx", func(t *testing.T) {
				t.Parallel()

				defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
				var wg sync.WaitGroup

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				pp := NewMockRequestFunc[int, int](t)

				requester := RequesterC(
					ctx,
					&wg,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStream <-chan Request[int, int],
					) {
						wg.Add(1)

						go func() {
							defer wg.Done()

							for {
								select {
								case <-ctx.Done():
									return
								case r, more := <-inStream:
									if !more {
										return
									}

									res, err := pp.Execute(ctx, r.P)
									r.Return(
										ctx, ValErrorPair[int]{
											V:   res,
											Err: err,
										},
									)
								}
							}
						}()
					},
				)

				cancel()
				require.Eventually(
					t, func() bool {
						wg.Wait()
						return true
					}, time.Second, time.Millisecond,
				)

				{
					ctx2, cancel2 := context.WithCancel(context.Background())
					defer cancel2()

					r, err := requester(ctx2, 2)

					require.Equal(t, 0, r)
					require.ErrorIs(t, err, context.Canceled)
					cancel2()
				}
			},
		)
}
*/
