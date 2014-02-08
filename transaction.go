package sqlmock

import (
	"fmt"
)

type transaction struct {
	conn *conn
}

func (tx *transaction) Commit() error {
	e := tx.conn.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to commit transaction was not expected")
	}

	etc, ok := e.(*expectedCommit)
	if !ok {
		return fmt.Errorf("call to commit transaction, was not expected, next expectation was %v", e)
	}
	etc.triggered = true
	return etc.err
}

func (tx *transaction) Rollback() error {
	e := tx.conn.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to rollback transaction was not expected")
	}

	etr, ok := e.(*expectedRollback)
	if !ok {
		return fmt.Errorf("call to rollback transaction, was not expected, next expectation was %v", e)
	}
	etr.triggered = true
	return etr.err
}
