package ports

import "context"

type FAtomicCallback func(d IUnitOfWork) error

type IUnitOfWork interface {
	Job() IJobRepository
	Attempt() IAttemptRepository
	Event() IEventRepository
	// DeadLetter() IDeadLetterRepository
	Atomic(ctx context.Context, fn FAtomicCallback) error
}
