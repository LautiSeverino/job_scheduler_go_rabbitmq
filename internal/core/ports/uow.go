package ports

import "context"

type FAtomicCallback func(d IUnitOfWork) error

type IUnitOfWork interface {
	Atomic(ctx context.Context, fn FAtomicCallback) error
}
