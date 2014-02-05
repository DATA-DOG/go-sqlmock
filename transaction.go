package sqlmock

import (
	"errors"
	"fmt"
)

type transaction struct {
	conn *conn
}

func (tx *transaction) Commit() error {
	e := tx.conn.next()
	if e == nil {
		return errors.New("All expectations were already fulfilled, call to Commit transaction was not expected")
	}

	etc, ok := e.(*expectedCommit)
	if !ok {
		return errors.New(fmt.Sprintf("Call to Commit transaction, was not expected, next expectation was %v", e))
	}
	etc.triggered = true
	return etc.err
}

func (tx *transaction) Rollback() error {
	e := tx.conn.next()
	if e == nil {
		return errors.New("All expectations were already fulfilled, call to Rollback transaction was not expected")
	}

	etr, ok := e.(*expectedRollback)
	if !ok {
		return errors.New(fmt.Sprintf("Call to Rollback transaction, was not expected, next expectation was %v", e))
	}
	etr.triggered = true
	return etr.err
}
