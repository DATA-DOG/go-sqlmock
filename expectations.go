package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"
)

// an expectation interface
type expectation interface {
	fulfilled() bool
	Lock()
	Unlock()
	String() string
}

// common expectation struct
// satisfies the expectation interface
type commonExpectation struct {
	sync.Mutex
	triggered bool
	err       error
}

func (e *commonExpectation) fulfilled() bool {
	return e.triggered
}

// ExpectedClose is used to manage *sql.DB.Close expectation
// returned by *Sqlmock.ExpectClose.
type ExpectedClose struct {
	commonExpectation
}

// WillReturnError allows to set an error for *sql.DB.Close action
func (e *ExpectedClose) WillReturnError(err error) *ExpectedClose {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedClose) String() string {
	msg := "ExpectedClose => expecting database Close"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}

// ExpectedBegin is used to manage *sql.DB.Begin expectation
// returned by *Sqlmock.ExpectBegin.
type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}

// WillReturnError allows to set an error for *sql.DB.Begin action
func (e *ExpectedBegin) WillReturnError(err error) *ExpectedBegin {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedBegin) String() string {
	msg := "ExpectedBegin => expecting database transaction Begin"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}

// WillDelayFor allows to specify duration for which it will delay
// result. May be used together with Context
func (e *ExpectedBegin) WillDelayFor(duration time.Duration) *ExpectedBegin {
	e.delay = duration
	return e
}

// ExpectedCommit is used to manage *sql.Tx.Commit expectation
// returned by *Sqlmock.ExpectCommit.
type ExpectedCommit struct {
	commonExpectation
}

// WillReturnError allows to set an error for *sql.Tx.Close action
func (e *ExpectedCommit) WillReturnError(err error) *ExpectedCommit {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedCommit) String() string {
	msg := "ExpectedCommit => expecting transaction Commit"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}

// ExpectedRollback is used to manage *sql.Tx.Rollback expectation
// returned by *Sqlmock.ExpectRollback.
type ExpectedRollback struct {
	commonExpectation
}

// WillReturnError allows to set an error for *sql.Tx.Rollback action
func (e *ExpectedRollback) WillReturnError(err error) *ExpectedRollback {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedRollback) String() string {
	msg := "ExpectedRollback => expecting transaction Rollback"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}

// ExpectedSql is used to manage *sql.DB.Query, *dql.DB.QueryRow, *sql.Tx.Query,
// *sql.Tx.QueryRow, *sql.Stmt.Query or *sql.Stmt.QueryRow expectations.
// Returned by *Sqlmock.ExpectQuery.
type ExpectedSql struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
	result           driver.Result
	expectedOpt      Matcher
}

// WithArgsCheck match sql args
func (e *ExpectedSql) WithArgsCheck(checkArgs func(opt string, sql string, args []driver.NamedValue) error) *ExpectedSql {
	e.checkArgs = checkArgs
	return e
}

// WithArgs will match given expected args to actual database query arguments.
// if at least one argument does not match, it will return an error. For specific
// arguments an sqlmock.Matcher interface can be used to match an argument.
func (e *ExpectedSql) WithArgs(args ...driver.Value) *ExpectedSql {
	e.args = args
	return e
}

// RowsWillBeClosed expects this query rows to be closed.
func (e *ExpectedSql) RowsWillBeClosed() *ExpectedSql {
	e.rowsMustBeClosed = true
	return e
}

// WillReturnError allows to set an error for expected database query
func (e *ExpectedSql) WillReturnError(err error) *ExpectedSql {
	e.err = err
	return e
}

// WillDelayFor allows to specify duration for which it will delay
// result. May be used together with Context
func (e *ExpectedSql) WillDelayFor(duration time.Duration) *ExpectedSql {
	e.delay = duration
	return e
}

func (e *ExpectedSql) WillReturnResult(result driver.Result) *ExpectedSql {
	e.result = result
	return e
}

func (e *ExpectedSql) WillReturnRows(rows ...*Rows) *ExpectedSql {
	sets := make([]*Rows, len(rows))
	for i, r := range rows {
		sets[i] = r
	}
	e.rows = &rowSets{sets: sets, ex: e}
	return e
}

// String returns string representation
func (e *ExpectedSql) String() string {
	msg := "ExpectedSql => expecting Query, QueryContext or QueryRow which:"
	msg += "\n  - matches sql: '" + e.expectSQL + "'"

	if len(e.args) == 0 {
		msg += "\n  - is without arguments"
	} else {
		msg += "\n  - is with arguments:\n"
		for i, arg := range e.args {
			msg += fmt.Sprintf("    %d - %+v\n", i, arg)
		}
		msg = strings.TrimSpace(msg)
	}

	if e.rows != nil {
		msg += fmt.Sprintf("\n  - %s", e.rows)
	}

	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}

	return msg
}

// ExpectedPrepare is used to manage *sql.DB.Prepare or *sql.Tx.Prepare expectations.
// Returned by *Sqlmock.ExpectPrepare.
type ExpectedPrepare struct {
	commonExpectation
	mock         *sqlmock
	expectSQL    string
	statement    driver.Stmt
	closeErr     error
	mustBeClosed bool
	wasClosed    bool
	delay        time.Duration
}

// WillReturnError allows to set an error for the expected *sql.DB.Prepare or *sql.Tx.Prepare action.
func (e *ExpectedPrepare) WillReturnError(err error) *ExpectedPrepare {
	e.err = err
	return e
}

// WillReturnCloseError allows to set an error for this prepared statement Close action
func (e *ExpectedPrepare) WillReturnCloseError(err error) *ExpectedPrepare {
	e.closeErr = err
	return e
}

// WillDelayFor allows to specify duration for which it will delay
// result. May be used together with Context
func (e *ExpectedPrepare) WillDelayFor(duration time.Duration) *ExpectedPrepare {
	e.delay = duration
	return e
}

// WillBeClosed expects this prepared statement to
// be closed.
func (e *ExpectedPrepare) WillBeClosed() *ExpectedPrepare {
	e.mustBeClosed = true
	return e
}

// String returns string representation
func (e *ExpectedPrepare) String() string {
	msg := "ExpectedPrepare => expecting Prepare statement which:"
	msg += "\n  - matches sql: '" + e.expectSQL + "'"

	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}

	if e.closeErr != nil {
		msg += fmt.Sprintf("\n  - should return error on Close: %s", e.closeErr)
	}

	return msg
}

// query based expectation
// adds a query matching logic
type queryBasedExpectation struct {
	commonExpectation
	expectSQL string
	converter driver.ValueConverter
	args      []driver.Value
	checkArgs func(opt string, sql string, args []driver.NamedValue) error
}

// ExpectedPing is used to manage *sql.DB.Ping expectations.
// Returned by *Sqlmock.ExpectPing.
type ExpectedPing struct {
	commonExpectation
	delay time.Duration
}

// WillDelayFor allows to specify duration for which it will delay result. May
// be used together with Context.
func (e *ExpectedPing) WillDelayFor(duration time.Duration) *ExpectedPing {
	e.delay = duration
	return e
}

// WillReturnError allows to set an error for expected database ping
func (e *ExpectedPing) WillReturnError(err error) *ExpectedPing {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedPing) String() string {
	msg := "ExpectedPing => expecting database Ping"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}

type ExpectedOperation struct {
	commonExpectation
	arg Matcher
}

// WillReturnError allows to set an error for *sql.DB.Begin action
func (e *ExpectedOperation) WillReturnError(err error) *ExpectedOperation {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedOperation) String() string {
	msg := "ExpectedOperation => expecting database transaction Begin"
	if e.err != nil {
		msg += fmt.Sprintf(", which should return error: %s", e.err)
	}
	return msg
}
