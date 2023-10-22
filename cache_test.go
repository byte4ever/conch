package conch

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type PP uint64

func (p PP) Hash() Key {
	return Key{
		A: uint64(p),
		B: uint64(p),
	}
}

var matchAnyCtx = mock.MatchedBy(
	func(ctx context.Context) bool {
		return true
	},
)

var errMocked = errors.New("mocked")

func TestCacheInterceptorC(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cacheMock := NewMockCache[PP, uint64](t)

	cacheMock.
		On(
			"Get",
			matchAnyCtx,
			PP(0),
		).
		Return(
			uint64(1000),
			true,
		).
		Once().
		On(
			"Get",
			matchAnyCtx,
			PP(1),
		).
		Return(
			uint64(0),
			false,
		).
		Once().
		On(
			"Store",
			matchAnyCtx,
			PP(1),
			uint64(100),
		).
		Return().
		Once().
		On(
			"Get",
			matchAnyCtx,
			PP(2),
		).
		Return(
			uint64(0),
			false,
		)

	requesterFunc := NewMockRequestFunc[PP, uint64](t)
	requesterFunc.
		On(
			"Execute",
			matchAnyCtx,
			PP(1),
		).
		Return(
			uint64(100),
			nil,
		).Once().
		On(
			"Execute",
			matchAnyCtx,
			PP(2),
		).
		Return(
			uint64(0),
			errMocked,
		).Once()

	requester := RequesterC(
		ctx,
		&wg,
		CacheReadInterceptorsC[PP, uint64](
			cacheMock,
			CacheWriteInterceptorsC[PP, uint64](
				cacheMock,
				RequestConsumersC(
					requesterFunc.Execute,
				),
			),
		),
	)

	// result is in cache
	v, err := requester(ctx, PP(0))
	require.Equal(t, uint64(1000), v)
	require.NoError(t, err)

	// not in cache so it's stored evaluated and stored in cache
	v, err = requester(ctx, PP(1))
	require.Equal(t, uint64(100), v)
	require.NoError(t, err)

	// not in cache so it's stored evaluated and not stored in cache because it
	// returns an error.
	v, err = requester(ctx, PP(2))
	require.Equal(t, uint64(0), v)
	require.ErrorIs(t, err, errMocked)

	cancel()
	wg.Wait()
}

func TestCacheInterceptorsC(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cacheMock := NewMockCache[PP, uint64](t)

	cacheMock.
		On(
			"Get",
			matchAnyCtx,
			PP(0),
		).
		Return(
			uint64(1000),
			true,
		).
		Once().
		On(
			"Get",
			matchAnyCtx,
			PP(1),
		).
		Return(
			uint64(0),
			false,
		).
		Once().
		On(
			"Store",
			matchAnyCtx,
			PP(1),
			uint64(100),
		).
		Return().
		Once().
		On(
			"Get",
			matchAnyCtx,
			PP(2),
		).
		Return(
			uint64(0),
			false,
		)

	requesterFunc := NewMockRequestFunc[PP, uint64](t)
	requesterFunc.
		On(
			"Execute",
			matchAnyCtx,
			PP(1),
		).
		Return(
			uint64(100),
			nil,
		).Once().
		On(
			"Execute",
			matchAnyCtx,
			PP(2),
		).
		Return(
			uint64(0),
			errMocked,
		).Once()

	requester := RequesterC(
		ctx,
		&wg,
		MultiplexC(
			10,
			CacheReadInterceptorsC[PP, uint64](
				cacheMock,
				CacheWriteInterceptorsC[PP, uint64](
					cacheMock,
					RequestConsumersC(
						requesterFunc.Execute,
					),
				),
			),
		),
	)

	// result is in cache
	v, err := requester(ctx, PP(0))
	require.Equal(t, uint64(1000), v)
	require.NoError(t, err)

	// not in cache so it's stored evaluated and stored in cache
	v, err = requester(ctx, PP(1))
	require.Equal(t, uint64(100), v)
	require.NoError(t, err)

	// not in cache so it's stored evaluated and not stored in cache because it
	// returns an error.
	v, err = requester(ctx, PP(2))
	require.Equal(t, uint64(0), v)
	require.ErrorIs(t, err, errMocked)

	cancel()
	wg.Wait()
}
