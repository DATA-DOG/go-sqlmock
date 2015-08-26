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

	// MatchExpectationsInOrder gives an option whether to match all
	// expectations in the order they were set or not.
	//
	// By default it is set to - true. But if you use goroutines
	// to parallelize your query executation, that option may
	// be handy.
	MatchExpectationsInOrder bool

	dsn    string
	opened int
	drv    *mockDriver

	expected []expectation
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

	var expected *ExpectedClose
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if expected, ok = next.(*ExpectedClose); ok {
			break
		}

		next.Unlock()
		if c.MatchExpectationsInOrder {
			return fmt.Errorf("call to database Close, was not expected, next expectation is %T as %+v", next, next)
		}
	}
	if expected == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to database Close was not expected")
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
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
	var expected *ExpectedBegin
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if expected, ok = next.(*ExpectedBegin); ok {
			break
		}

		next.Unlock()
		if c.MatchExpectationsInOrder {
			return nil, fmt.Errorf("call to begin transaction, was not expected, next expectation is %T as %+v", next, next)
		}
	}
	if expected == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to begin transaction was not expected")
	}

	expected.triggered = true
	expected.Unlock()
	return c, expected.err
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
	query = stripQuery(query)
	var expected *ExpectedExec
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if c.MatchExpectationsInOrder {
			if expected, ok = next.(*ExpectedExec); ok {
				break
			}
			next.Unlock()
			return nil, fmt.Errorf("call to exec query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, next, next)
		}
		if exec, ok := next.(*ExpectedExec); ok {
			if exec.attemptMatch(query, args) {
				expected = exec
				break
			}
		}
		next.Unlock()
	}
	if expected == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to exec '%s' query with args %+v was not expected", query, args)
	}

	defer expected.Unlock()
	expected.triggered = true
	// converts panic to error in case of reflect value type mismatch
	defer func(errp *error, exp *ExpectedExec, q string, a []driver.Value) {
		if e := recover(); e != nil {
			if se, ok := e.(*reflect.ValueError); ok { // catch reflect error, failed type conversion
				msg := "exec query \"%s\", args \"%+v\" failed to match expected arguments \"%+v\", reason %s"
				*errp = fmt.Errorf(msg, q, a, exp.args, se)
			} else {
				panic(e) // overwise if unknown error panic
			}
		}
	}(&err, expected, query, args)

	if !expected.queryMatches(query) {
		return nil, fmt.Errorf("exec query '%s', does not match regex '%s'", query, expected.sqlRegex.String())
	}

	if !expected.argsMatches(args) {
		return nil, fmt.Errorf("exec query '%s', args %+v does not match expected %+v", query, args, expected.args)
	}

	if expected.err != nil {
		return nil, expected.err // mocked to return error
	}

	if expected.result == nil {
		return nil, fmt.Errorf("exec query '%s' with args %+v, must return a database/sql/driver.result, but it was not set for expectation %T as %+v", query, args, expected, expected)
	}
	return expected.result, err
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
	var expected *ExpectedPrepare
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if expected, ok = next.(*ExpectedPrepare); ok {
			break
		}

		next.Unlock()
		if c.MatchExpectationsInOrder {
			return nil, fmt.Errorf("call to Prepare stetement with query '%s', was not expected, next expectation is %T as %+v", query, next, next)
		}
	}

	query = stripQuery(query)
	if expected == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to Prepare '%s' query was not expected", query)
	}

	expected.triggered = true
	expected.Unlock()
	return &statement{c, query, expected.closeErr}, expected.err
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
	query = stripQuery(query)
	var expected *ExpectedQuery
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if c.MatchExpectationsInOrder {
			if expected, ok = next.(*ExpectedQuery); ok {
				break
			}
			next.Unlock()
			return nil, fmt.Errorf("call to query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, next, next)
		}
		if qr, ok := next.(*ExpectedQuery); ok {
			if qr.attemptMatch(query, args) {
				expected = qr
				break
			}
		}
		next.Unlock()
	}
	if expected == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to query '%s' with args %+v was not expected", query, args)
	}

	defer expected.Unlock()
	expected.triggered = true
	// converts panic to error in case of reflect value type mismatch
	defer func(errp *error, exp *ExpectedQuery, q string, a []driver.Value) {
		if e := recover(); e != nil {
			if se, ok := e.(*reflect.ValueError); ok { // catch reflect error, failed type conversion
				msg := "query \"%s\", args \"%+v\" failed to match expected arguments \"%+v\", reason %s"
				*errp = fmt.Errorf(msg, q, a, exp.args, se)
			} else {
				panic(e) // overwise if unknown error panic
			}
		}
	}(&err, expected, query, args)

	if !expected.queryMatches(query) {
		return nil, fmt.Errorf("query '%s', does not match regex [%s]", query, expected.sqlRegex.String())
	}

	if !expected.argsMatches(args) {
		return nil, fmt.Errorf("query '%s', args %+v does not match expected %+v", query, args, expected.args)
	}

	if expected.err != nil {
		return nil, expected.err // mocked to return error
	}

	if expected.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.rows, but it was not set for expectation %T as %+v", query, args, expected, expected)
	}

	return expected.rows, err
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
	var expected *ExpectedCommit
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if expected, ok = next.(*ExpectedCommit); ok {
			break
		}

		next.Unlock()
		if c.MatchExpectationsInOrder {
			return fmt.Errorf("call to commit transaction, was not expected, next expectation is %T as %+v", next, next)
		}
	}
	if expected == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to commit transaction was not expected")
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
}

// Rollback meets http://golang.org/pkg/database/sql/driver/#Tx
func (c *Sqlmock) Rollback() error {
	var expected *ExpectedRollback
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			continue
		}

		if expected, ok = next.(*ExpectedRollback); ok {
			break
		}

		next.Unlock()
		if c.MatchExpectationsInOrder {
			return fmt.Errorf("call to rollback transaction, was not expected, next expectation is %T as %+v", next, next)
		}
	}
	if expected == nil {
		return fmt.Errorf("all expectations were already fulfilled, call to rollback transaction was not expected")
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
}
