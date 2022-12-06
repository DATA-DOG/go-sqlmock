package sqlmock

import "database/sql/driver"

var _ driver.Stmt = (*statement)(nil)

type statement struct {
	conn  *sqlmock
	ex    *ExpectedPrepare
	query string
}

func (stmt *statement) Close() error {
	stmt.ex.wasClosed = true
	return stmt.ex.closeErr
}

func (stmt *statement) NumInput() int {
	return -1
}
