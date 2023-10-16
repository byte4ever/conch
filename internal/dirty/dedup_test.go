package dirty

/*
func Test_pushFunc(t *testing.T) {
	var m sync.Map

	dedupPool := NewDedupPool[int](10)

	f := pushFunc(
		dedupPool,
		Key{1, 2},
		&m,
		func(ctx context.Context, v ValErrorPair[int]) {
			fmt.Println("pushFunc1:", v)
		},
	)

	require.NotNil(t, f)

	require.Nil(
		t, pushFunc(
			dedupPool,
			Key{1, 2},
			&m,
			func(ctx context.Context, v ValErrorPair[int]) {
				fmt.Println("pushFunc2:", v)
			},
		),
	)

	require.Nil(
		t, pushFunc(
			dedupPool,
			Key{1, 2},
			&m,
			func(ctx context.Context, v ValErrorPair[int]) {
				fmt.Println("pushFunc3:", v)
			},
		),
	)

	require.Nil(
		t, pushFunc(
			dedupPool,
			Key{1, 2},
			&m,
			func(ctx context.Context, v ValErrorPair[int]) {
				fmt.Println("pushFunc4:", v)
			},
		),
	)

	f(
		context.Background(), ValErrorPair[int]{
			V:   10,
			Err: nil,
		},
	)

	time.Sleep(time.Second)

}
*/
