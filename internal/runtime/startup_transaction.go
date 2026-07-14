package runtime

import (
	"context"
	"errors"
)

type startupRollback func(context.Context) error

type startupTransaction struct {
	rollbacks []startupRollback
}

func (transaction *startupTransaction) acquire(
	acquire func() (startupRollback, error),
) error {
	rollback, err := acquire()
	if err != nil {
		return err
	}
	if rollback != nil {
		transaction.rollbacks = append(transaction.rollbacks, rollback)
	}
	return nil
}

func (transaction *startupTransaction) commit() {
	transaction.rollbacks = nil
}

func (transaction *startupTransaction) rollback(ctx context.Context) error {
	errorsList := make([]error, 0, len(transaction.rollbacks))
	for index := len(transaction.rollbacks) - 1; index >= 0; index-- {
		if err := transaction.rollbacks[index](ctx); err != nil {
			errorsList = append(errorsList, err)
		}
	}
	transaction.rollbacks = nil
	return errors.Join(errorsList...)
}
