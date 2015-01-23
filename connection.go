package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

type conn struct {
	expectations []expectation
	active       expectation
}

// Close a mock database driver connection. It should
// be always called to ensure that all expectations
// were met successfully. Returns error if there is any
func (c *conn) Close() (err error) {
	for _, e := range mock.conn.expectations {
		if !e.fulfilled() {
			err = fmt.Errorf("there is a remaining expectation %T which was not matched yet", e)
			break
		}
	}
	mock.conn.expectations = []expectation{}
	mock.conn.active = nil
	return err
}

func (c *conn) Begin() (driver.Tx, error) {
	e := c.next()
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to begin transaction was not expected")
	}

	etb, ok := e.(*expectedBegin)
	if !ok {
		return nil, fmt.Errorf("call to begin transaction, was not expected, next expectation is %T as %+v", e, e)
	}
	etb.triggered = true
	return &transaction{c}, etb.err
}

// get next unfulfilled expectation
func (c *conn) next() (e expectation) {
	for _, e = range c.expectations {
		if !e.fulfilled() {
			return
		}
	}
	return nil // all expectations were fulfilled
}

func (c *conn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to exec '%s' query with args %+v was not expected", query, args)
	}

	eq, ok := e.(*expectedExec)
	if !ok {
		return nil, fmt.Errorf("call to exec query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("exec query '%s', does not match regex '%s'", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("exec query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.result == nil {
		return nil, fmt.Errorf("exec query '%s' with args %+v, must return a database/sql/driver.result, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	return eq.result, err
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	e := c.next()

	// for backwards compatibility, ignore when Prepare not expected
	if e == nil {
		return &statement{mock.conn, stripQuery(query)}, nil
	}
	eq, ok := e.(*expectedPrepare)
	if !ok {
		return &statement{mock.conn, stripQuery(query)}, nil
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	return &statement{mock.conn, stripQuery(query)}, nil
}

func (c *conn) Query(query string, args []driver.Value) (rw driver.Rows, err error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("all expectations were already fulfilled, call to query '%s' with args %+v was not expected", query, args)
	}

	eq, ok := e.(*expectedQuery)
	if !ok {
		return nil, fmt.Errorf("call to query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true

	defer argMatcherErrorHandler(&err) // converts panic to error in case of reflect value type mismatch

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("query '%s', does not match regex [%s]", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.rows, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	return eq.rows, err
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
