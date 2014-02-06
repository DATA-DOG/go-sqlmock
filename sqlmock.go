package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var mock *mockDriver

type Mock interface {
	WithArgs(...driver.Value) Mock
	WillReturnError(error) Mock
	WillReturnRows(driver.Rows) Mock
	WillReturnResult(driver.Result) Mock
}

type mockDriver struct {
	conn *conn
}

func (d *mockDriver) Open(dsn string) (driver.Conn, error) {
	return mock.conn, nil
}

func init() {
	mock = &mockDriver{&conn{}}
	sql.Register("mock", mock)
}

type conn struct {
	expectations []expectation
	active       expectation
}

func stripQuery(q string) (s string) {
	s = strings.Replace(q, "\n", " ", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.TrimSpace(s)
	return
}

func (c *conn) Close() (err error) {
	for _, e := range mock.conn.expectations {
		if !e.fulfilled() {
			err = errors.New(fmt.Sprintf("There is a remaining expectation %T which was not matched yet", e))
			break
		}
	}
	mock.conn.expectations = []expectation{}
	mock.conn.active = nil
	return err
}

func ExpectBegin() Mock {
	e := &expectedBegin{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

func ExpectCommit() Mock {
	e := &expectedCommit{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

func ExpectRollback() Mock {
	e := &expectedRollback{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

func (c *conn) WillReturnError(err error) Mock {
	c.active.setError(err)
	return c
}

func (c *conn) Begin() (driver.Tx, error) {
	e := c.next()
	if e == nil {
		return nil, errors.New("All expectations were already fulfilled, call to Begin transaction was not expected")
	}

	etb, ok := e.(*expectedBegin)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Call to Begin transaction, was not expected, next expectation is %+v", e))
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
		return nil, errors.New(fmt.Sprintf("All expectations were already fulfilled, call to Exec '%s' query with args %+v was not expected", query, args))
	}

	eq, ok := e.(*expectedExec)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Call to Exec query '%s' with args %+v, was not expected, next expectation is %+v", query, args, e))
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.result == nil {
		return nil, errors.New(fmt.Sprintf("Exec query '%s' with args %+v, must return a database/sql/driver.Result, but it was not set for expectation %+v", query, args, eq))
	}

	if !eq.queryMatches(query) {
		return nil, errors.New(fmt.Sprintf("Exec query '%s', does not match regex '%s'", query, eq.sqlRegex.String()))
	}

	if !eq.argsMatches(args) {
		return nil, errors.New(fmt.Sprintf("Exec query '%s', args %+v does not match expected %+v", query, args, eq.args))
	}

	return eq.result, nil
}

func ExpectExec(sqlRegexStr string) Mock {
	e := &expectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

func ExpectQuery(sqlRegexStr string) Mock {
	e := &expectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)

	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

func (c *conn) WithArgs(args ...driver.Value) Mock {
	eq, ok := c.active.(*expectedQuery)
	if !ok {
		ee, ok := c.active.(*expectedExec)
		if !ok {
			panic(fmt.Sprintf("Arguments may be expected only with query based expectations, current is %T", c.active))
		}
		ee.args = args
	} else {
		eq.args = args
	}
	return c
}

func (c *conn) WillReturnResult(result driver.Result) Mock {
	eq, ok := c.active.(*expectedExec)
	if !ok {
		panic(fmt.Sprintf("driver.Result may be returned only by Exec expectations, current is %+v", c.active))
	}
	eq.result = result
	return c
}

func (c *conn) WillReturnRows(rows driver.Rows) Mock {
	eq, ok := c.active.(*expectedQuery)
	if !ok {
		panic(fmt.Sprintf("driver.Rows may be returned only by Query expectations, current is %+v", c.active))
	}
	eq.rows = rows
	return c
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &statement{mock.conn, stripQuery(query)}, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	e := c.next()
	query = stripQuery(query)
	if e == nil {
		return nil, errors.New(fmt.Sprintf("All expectations were already fulfilled, call to Query '%s' with args %+v was not expected", query, args))
	}

	eq, ok := e.(*expectedQuery)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Call to Query '%s' with args %+v, was not expected, next expectation is %+v", query, args, e))
	}

	eq.triggered = true
	if eq.err != nil {
		return nil, eq.err // mocked to return error
	}

	if eq.rows == nil {
		return nil, errors.New(fmt.Sprintf("Query '%s' with args %+v, must return a database/sql/driver.Rows, but it was not set for expectation %+v", query, args, eq))
	}

	if !eq.queryMatches(query) {
		return nil, errors.New(fmt.Sprintf("Query '%s', does not match regex [%s]", query, eq.sqlRegex.String()))
	}

	if !eq.argsMatches(args) {
		return nil, errors.New(fmt.Sprintf("Query '%s', args %+v does not match expected %+v", query, args, eq.args))
	}

	return eq.rows, nil
}
