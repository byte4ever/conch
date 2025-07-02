package domain

import (
	"context"
)

type Processable[Param any, Result any] interface {
	GetParam() Param
	SetValue(r Result)
	GetValue() Result
	SetError(e error)
	GetError() error
}

type ProcessorFunc[T Processable[Param, Result], Param any, Result any] func(
	ctx context.Context,
	elem T,
)
