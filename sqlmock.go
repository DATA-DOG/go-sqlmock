/*
Package sqlmock provides sql driver mock connecection, which allows to test database,
create expectations and ensure the correct execution flow of any database operations.
It hooks into Go standard library's database/sql package.

The package provides convenient methods to mock database queries, transactions and
expect the right execution flow, compare query arguments or even return error instead
to simulate failures. See the example bellow, which illustrates how convenient it is
to work with.
*/
package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
)

// Sqlmock type satisfies required sql.driver interfaces
// to simulate actual database and also serves to
// create expectations for any kind of database action
// in order to mock and test real database behavior.
type Sqlmock struct {
	dsn    string
	opened int
	drv    *mockDriver

	expected []expectation
}

func (c *Sqlmock) next() (e expectation) {
	for _, e = range c.expected {
		if !e.fulfilled() {
			return
		}
	}
	return nil // all expectations were fulfilled
}

// ExpectClose queues an expectation for this database
// action to be triggered. the *ExpectedClose allows
// to mock database response
func (c *Sqlmock) ExpectClose() *ExpectedClose {
	e := &ExpectedClose{}
	c.expected = append(c.expected, e)
	return e
}

// Close a mock database driver connection. It may or may not
// be called depending on the sircumstances, but if it is called
// there must be an *ExpectedClose expectation satisfied.
// meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *Sqlmock) Close() error {
	c.drv.Lock()
	defer c.drv.Unlock()

	c.opened--
	if c.opened == 0 {
		delete(c.drv.conns, c.dsn)
	}
	e := c.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to database Close was not expected")
	}

	t, ok := e.(*ExpectedClose)
	if !ok {
		return fmt.Errorf("call to database Close, was not expected, next expectation is %T as %+v", e, e)
	}
	t.triggered = true
	return t.err
}

// ExpectationsWereMet checks whether all queued expectations
// were met in order. If any of them was not met - an error is returned.
func (c *Sqlmock) ExpectationsWereMet() error {
	for _, e := range c.expected {
		if !e.fulfilled() {
			return fmt.Errorf("there is a remaining expectation %T which was not matched yet", e)
		}
	}
	return nil
}

// Begin meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *Sqlmock) Begin() (driver.Tx, error) {
	e := c.next()
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to begin transaction was not expected")
	}

	t, ok := e.(*ExpectedBegin)
	if !ok {
		return nil, fmt.Errorf("call to begin transaction, was not expected, next expectation is %T as %+v", e, e)
	}
	t.triggered = true
	return c, t.err
}

// ExpectBegin expects *sql.DB.Begin to be called.
// the *ExpectedBegin allows to mock database response
func (c *Sqlmock) ExpectBegin() *ExpectedBegin {
	e := &ExpectedBegin{}
	c.expected = append(c.expected, e)
	return e
}

// Exec meets http://golang.org/pkg/database/sql/driver/#Execer
func (c *Sqlmock) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to exec '%s' query with args %+v was not expected", query, args)
	}

	t, ok := e.(*ExpectedExec)
	if !ok {
		return nil, fmt.Errorf("call to exec query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	t.triggered = true
	if t.err != nil {
		return nil, t.err // mocked to return error
	}

	if t.result == nil {
		return nil, fmt.Errorf("exec query '%s' with args %+v, must return a database/sql/driver.result, but it was not set for expectation %T as %+v", query, args, t, t)
	}

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !t.queryMatches(query) {
		return nil, fmt.Errorf("exec query '%s', does not match regex '%s'", query, t.sqlRegex.String())
	}

	if !t.argsMatches(args) {
		return nil, fmt.Errorf("exec query '%s', args %+v does not match expected %+v", query, args, t.args)
	}

	return t.result, err
}

// ExpectExec expects Exec() to be called with sql query
// which match sqlRegexStr given regexp.
// the *ExpectedExec allows to mock database response
func (c *Sqlmock) ExpectExec(sqlRegexStr string) *ExpectedExec {
	e := &ExpectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	c.expected = append(c.expected, e)
	return e
}

// Prepare meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *Sqlmock) Prepare(query string) (driver.Stmt, error) {
	e := c.next()

	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to Prepare '%s' query was not expected", query)
	}
	t, ok := e.(*ExpectedPrepare)
	if !ok {
		return nil, fmt.Errorf("call to Prepare stetement with query '%s', was not expected, next expectation is %T as %+v", query, e, e)
	}

	t.triggered = true
	if t.err != nil {
		return nil, t.err // mocked to return error
	}

	return &statement{c, query, t.closeErr}, nil
}

// ExpectPrepare expects Prepare() to be called with sql query
// which match sqlRegexStr given regexp.
// the *ExpectedPrepare allows to mock database response.
// Note that you may expect Query() or Exec() on the *ExpectedPrepare
// statement to prevent repeating sqlRegexStr
func (c *Sqlmock) ExpectPrepare(sqlRegexStr string) *ExpectedPrepare {
	e := &ExpectedPrepare{sqlRegex: regexp.MustCompile(sqlRegexStr), mock: c}
	c.expected = append(c.expected, e)
	return e
}

// Query meets http://golang.org/pkg/database/sql/driver/#Queryer
func (c *Sqlmock) Query(query string, args []driver.Value) (rw driver.Rows, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to query '%s' with args %+v was not expected", query, args)
	}

	t, ok := e.(*ExpectedQuery)
	if !ok {
		return nil, fmt.Errorf("call to query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	t.triggered = true
	if t.err != nil {
		return nil, t.err // mocked to return error
	}

	// remove when rows_test.go:186 is available, won't cause a BC
	if t.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.rows, but it was not set for expectation %T as %+v", query, args, t, t)
	}

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !t.queryMatches(query) {
		return nil, fmt.Errorf("query '%s', does not match regex [%s]", query, t.sqlRegex.String())
	}

	if !t.argsMatches(args) {
		return nil, fmt.Errorf("query '%s', args %+v does not match expected %+v", query, args, t.args)
	}

	return t.rows, err
}

// ExpectQuery expects Query() or QueryRow() to be called with sql query
// which match sqlRegexStr given regexp.
// the *ExpectedQuery allows to mock database response.
func (c *Sqlmock) ExpectQuery(sqlRegexStr string) *ExpectedQuery {
	e := &ExpectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	c.expected = append(c.expected, e)
	return e
}

// ExpectCommit expects *sql.Tx.Commit to be called.
// the *ExpectedCommit allows to mock database response
func (c *Sqlmock) ExpectCommit() *ExpectedCommit {
	e := &ExpectedCommit{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectRollback expects *sql.Tx.Rollback to be called.
// the *ExpectedRollback allows to mock database response
func (c *Sqlmock) ExpectRollback() *ExpectedRollback {
	e := &ExpectedRollback{}
	c.expected = append(c.expected, e)
	return e
}

// Commit meets http://golang.org/pkg/database/sql/driver/#Tx
func (c *Sqlmock) Commit() error {
	e := c.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to commit transaction was not expected")
	}

	t, ok := e.(*ExpectedCommit)
	if !ok {
		return fmt.Errorf("call to commit transaction, was not expected, next expectation was %v", e)
	}
	t.triggered = true
	return t.err
}

// Rollback meets http://golang.org/pkg/database/sql/driver/#Tx
func (c *Sqlmock) Rollback() error {
	e := c.next()
	if e == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to rollback transaction was not expected")
	}

	t, ok := e.(*ExpectedRollback)
	if !ok {
		return fmt.Errorf("call to rollback transaction, was not expected, next expectation was %v", e)
	}
	t.triggered = true
	return t.err
}

func argMatcherErrorHandler(errp *error) {
	if e := recover(); e != nil {
		if se, ok := e.(*reflect.ValueError); ok { // catch reflect error, failed type conversion
			*errp = fmt.Errorf("Failed to compare query arguments: %s", se)
		} else {
			panic(e) // overwise panic
		}
	}
}
