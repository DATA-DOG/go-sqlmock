package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
)

var mock *mockDriver

// Mock interface defines a mock which is returned
// by any expectation and can be detailed further
// with the following methods
type Mock interface {
	WithArgs(...driver.Value) Mock
	WillReturnError(error) Mock
	WillReturnRows(driver.Rows) Mock
	WillReturnResult(driver.Result) Mock
}

type mockDriver struct {
	conn *conn
}

// opens a mock driver database connection
func (d *mockDriver) Open(dsn string) (driver.Conn, error) {
	return mock.conn, nil
}

func init() {
	mock = &mockDriver{&conn{}}
	sql.Register("mock", mock)
}

// expect transaction to be started
func ExpectBegin() Mock {
	e := &expectedBegin{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// expect transaction to be commited
func ExpectCommit() Mock {
	e := &expectedCommit{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// expect transaction to be rolled back
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
		panic(fmt.Sprintf("driver.Result may be returned only by Exec expectations, current is %T", c.active))
	}
	eq.result = result
	return c
}

func (c *conn) WillReturnRows(rows driver.Rows) Mock {
	eq, ok := c.active.(*expectedQuery)
	if !ok {
		panic(fmt.Sprintf("driver.Rows may be returned only by Query expectations, current is %T", c.active))
	}
	eq.rows = rows
	return c
}
