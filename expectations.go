package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// Argument interface allows to match
// any argument in specific way when used with
// ExpectedQuery and ExpectedExec expectations.
type Argument interface {
	Match(driver.Value) bool
}

// an expectation interface
type expectation interface {
	fulfilled() bool
	check() bool
	Lock()
	Unlock()
	String() string
	Optional()
}

// common expectation struct
// satisfies the expectation interface
type commonExpectation struct {
	sync.Mutex
	triggered bool
	err       error
	optional bool
}

func (e *commonExpectation) fulfilled() bool {
	return e.triggered
}

func (e *commonExpectation) check() bool {
	return e.triggered || e.optional
}

func (e *commonExpectation) Optional() {
	e.optional = true
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

// ExpectedQuery is used to manage *sql.DB.Query, *dql.DB.QueryRow, *sql.Tx.Query,
// *sql.Tx.QueryRow, *sql.Stmt.Query or *sql.Stmt.QueryRow expectations.
// Returned by *Sqlmock.ExpectQuery.
type ExpectedQuery struct {
	queryBasedExpectation
	rows driver.Rows
}

// WithArgs will match given expected args to actual database query arguments.
// if at least one argument does not match, it will return an error. For specific
// arguments an sqlmock.Argument interface can be used to match an argument.
func (e *ExpectedQuery) WithArgs(args ...driver.Value) *ExpectedQuery {
	e.args = args
	return e
}

// WillReturnError allows to set an error for expected database query
func (e *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	e.err = err
	return e
}

// WillReturnRows specifies the set of resulting rows that will be returned
// by the triggered query
func (e *ExpectedQuery) WillReturnRows(rows driver.Rows) *ExpectedQuery {
	e.rows = rows
	return e
}

// String returns string representation
func (e *ExpectedQuery) String() string {
	msg := "ExpectedQuery => expecting Query or QueryRow which:"
	msg += "\n  - matches sql: '" + e.sqlRegex.String() + "'"

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
		msg += "\n  - should return rows:\n"
		rs, _ := e.rows.(*rows)
		for i, row := range rs.rows {
			msg += fmt.Sprintf("    %d - %+v\n", i, row)
		}
		msg = strings.TrimSpace(msg)
	}

	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}

	return msg
}

// ExpectedExec is used to manage *sql.DB.Exec, *sql.Tx.Exec or *sql.Stmt.Exec expectations.
// Returned by *Sqlmock.ExpectExec.
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
}

// WithArgs will match given expected args to actual database exec operation arguments.
// if at least one argument does not match, it will return an error. For specific
// arguments an sqlmock.Argument interface can be used to match an argument.
func (e *ExpectedExec) WithArgs(args ...driver.Value) *ExpectedExec {
	e.args = args
	return e
}

// WillReturnError allows to set an error for expected database exec action
func (e *ExpectedExec) WillReturnError(err error) *ExpectedExec {
	e.err = err
	return e
}

// String returns string representation
func (e *ExpectedExec) String() string {
	msg := "ExpectedExec => expecting Exec which:"
	msg += "\n  - matches sql: '" + e.sqlRegex.String() + "'"

	if len(e.args) == 0 {
		msg += "\n  - is without arguments"
	} else {
		msg += "\n  - is with arguments:\n"
		var margs []string
		for i, arg := range e.args {
			margs = append(margs, fmt.Sprintf("    %d - %+v", i, arg))
		}
		msg += strings.Join(margs, "\n")
	}

	if e.result != nil {
		res, _ := e.result.(*result)
		msg += "\n  - should return Result having:"
		msg += fmt.Sprintf("\n      LastInsertId: %d", res.insertID)
		msg += fmt.Sprintf("\n      RowsAffected: %d", res.rowsAffected)
		if res.err != nil {
			msg += fmt.Sprintf("\n      Error: %s", res.err)
		}
	}

	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}

	return msg
}

// WillReturnResult arranges for an expected Exec() to return a particular
// result, there is sqlmock.NewResult(lastInsertID int64, affectedRows int64) method
// to build a corresponding result. Or if actions needs to be tested against errors
// sqlmock.NewErrorResult(err error) to return a given error.
func (e *ExpectedExec) WillReturnResult(result driver.Result) *ExpectedExec {
	e.result = result
	return e
}

// ExpectedPrepare is used to manage *sql.DB.Prepare or *sql.Tx.Prepare expectations.
// Returned by *Sqlmock.ExpectPrepare.
type ExpectedPrepare struct {
	commonExpectation
	mock      *sqlmock
	sqlRegex  *regexp.Regexp
	statement driver.Stmt
	closeErr  error
}

// WillReturnError allows to set an error for the expected *sql.DB.Prepare or *sql.Tx.Prepare action.
func (e *ExpectedPrepare) WillReturnError(err error) *ExpectedPrepare {
	e.err = err
	return e
}

// WillReturnCloseError allows to set an error for this prapared statement Close action
func (e *ExpectedPrepare) WillReturnCloseError(err error) *ExpectedPrepare {
	e.closeErr = err
	return e
}

// ExpectQuery allows to expect Query() or QueryRow() on this prepared statement.
// this method is convenient in order to prevent duplicating sql query string matching.
func (e *ExpectedPrepare) ExpectQuery() *ExpectedQuery {
	eq := &ExpectedQuery{}
	eq.sqlRegex = e.sqlRegex
	e.mock.expected = append(e.mock.expected, eq)
	return eq
}

// ExpectExec allows to expect Exec() on this prepared statement.
// this method is convenient in order to prevent duplicating sql query string matching.
func (e *ExpectedPrepare) ExpectExec() *ExpectedExec {
	eq := &ExpectedExec{}
	eq.sqlRegex = e.sqlRegex
	e.mock.expected = append(e.mock.expected, eq)
	return eq
}

// String returns string representation
func (e *ExpectedPrepare) String() string {
	msg := "ExpectedPrepare => expecting Prepare statement which:"
	msg += "\n  - matches sql: '" + e.sqlRegex.String() + "'"

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
	sqlRegex *regexp.Regexp
	args     []driver.Value
}

func (e *queryBasedExpectation) attemptMatch(sql string, args []driver.Value) (ret bool) {
	if !e.queryMatches(sql) {
		return
	}

	defer recover() // ignore panic since we attempt a match

	if e.argsMatches(args) {
		return true
	}
	return
}

func (e *queryBasedExpectation) queryMatches(sql string) bool {
	return e.sqlRegex.MatchString(sql)
}

func (e *queryBasedExpectation) argsMatches(args []driver.Value) bool {
	if nil == e.args {
		return true
	}
	if len(args) != len(e.args) {
		return false
	}
	for k, v := range args {
		matcher, ok := e.args[k].(Argument)
		if ok {
			if !matcher.Match(v) {
				return false
			}
			continue
		}
		vi := reflect.ValueOf(v)
		ai := reflect.ValueOf(e.args[k])
		switch vi.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if vi.Int() != ai.Int() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if vi.Float() != ai.Float() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if vi.Uint() != ai.Uint() {
				return false
			}
		case reflect.String:
			if vi.String() != ai.String() {
				return false
			}
		default:
			// compare types like time.Time based on type only
			if vi.Kind() != ai.Kind() {
				return false
			}
		}
	}
	return true
}
