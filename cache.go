package conch

import (
	"context"

	"github.com/byte4ever/conch/domain"
)

type Cachable[Param domain.Hashable, Result any] interface {
	domain.Processable[Param, Result]
	GetValue() Result
}

type ProcessorCacheFunc[T Cachable[Param, Result], Param domain.Hashable, Result any] func(
	ctx context.Context,
	elem T,
)
