package sqlmock

import (
	"database/sql/driver"
	"fmt"
)

type conn struct {
	expectations []expectation
	active       expectation
}

// closes a mock database driver connection. It should
// be always called to ensure that all expectations
// were met successfully
func (c *conn) Close() (err error) {
	for _, e := range mock.conn.expectations {
		if !e.fulfilled() {
			err = fmt.Errorf("There is a remaining expectation %T which was not matched yet", e)
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
		return nil, fmt.Errorf("All expectations were already fulfilled, call to Begin transaction was not expected")
	}

	etb, ok := e.(*expectedBegin)
	if !ok {
		return nil, fmt.Errorf("Call to Begin transaction, was not expected, next expectation is %T as %+v", e, e)
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

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("All expectations were already fulfilled, call to Exec '%s' query with args %+v was not expected", query, args)
	}

	eq, ok := e.(*expectedExec)
	if !ok {
		return nil, fmt.Errorf("Call to Exec query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.result == nil {
		return nil, fmt.Errorf("Exec query '%s' with args %+v, must return a database/sql/driver.Result, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("Exec query '%s', does not match regex '%s'", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("Exec query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	return eq.result, nil
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &statement{mock.conn, stripQuery(query)}, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, fmt.Errorf("All expectations were already fulfilled, call to Query '%s' with args %+v was not expected", query, args)
	}

	eq, ok := e.(*expectedQuery)
	if !ok {
		return nil, fmt.Errorf("Call to Query '%s' with args %+v, was not expected, next expectation is %T as %+v", query, args, e, e)
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.rows == nil {
		return nil, fmt.Errorf("Query '%s' with args %+v, must return a database/sql/driver.Rows, but it was not set for expectation %T as %+v", query, args, eq, eq)
	}

	if !eq.queryMatches(query) {
		return nil, fmt.Errorf("Query '%s', does not match regex [%s]", query, eq.sqlRegex.String())
	}

	if !eq.argsMatches(args) {
		return nil, fmt.Errorf("Query '%s', args %+v does not match expected %+v", query, args, eq.args)
	}

	return eq.rows, nil
}

