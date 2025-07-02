package domain

import (
	"context"
)

type Cache[P Hashable, R any] interface {
	Get(ctx context.Context, key P) (R, bool)
	Store(ctx context.Context, key P, value R)
}

type Hashable interface {
	Hash() Key
}

type Hashable2[T any] interface {
	Hash() Key
	SetValue(v T)
}
