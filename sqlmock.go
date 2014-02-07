/*
Package sqlmock provides sql driver mock connecection, which allows to test database,
create expectations and ensure the correct execution flow of any database operations.
It hooks into Go standard library's database/sql package.

The package provides convenient methods to mock database queries, transactions and
expect the right execution flow, compare query arguments or even return error instead
to simulate failures. See the example bellow, which illustrates how convenient it is
to work with:


    package main

    import (
        "database/sql"
        "github.com/DATA-DOG/go-sqlmock"
        "testing"
        "fmt"
    )

    // will test that order with a different status, cannot be cancelled
    func TestShouldNotCancelOrderWithNonPendingStatus(t *testing.T) {
        // open database stub
        db, err := sql.Open("mock", "")
        if err != nil {
            t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
        }

        // columns to be used for result
        columns := []string{"id", "status"}
        // expect transaction begin
        sqlmock.ExpectBegin()
        // expect query to fetch order, match it with regexp
        sqlmock.ExpectQuery("SELECT (.+) FROM orders (.+) FOR UPDATE").
            WithArgs(1).
            WillReturnRows(sqlmock.RowsFromCSVString(columns, "1,1"))
        // expect transaction rollback, since order status is "cancelled"
        sqlmock.ExpectRollback()

        // run the cancel order function
        someOrderId := 1
        // call a function which executes expected database operations
        err = cancelOrder(someOrderId, db)
        if err != nil {
            t.Errorf("Expected no error, but got %s instead", err)
        }
        // db.Close() ensures that all expectations have been met
        if err = db.Close(); err != nil {
            t.Errorf("Error '%s' was not expected while closing the database", err)
        }
    }

*/
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
// with the methods this interface provides
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

// Expect transaction to be started
func ExpectBegin() Mock {
	e := &expectedBegin{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// Expect transaction to be commited
func ExpectCommit() Mock {
	e := &expectedCommit{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// Expect transaction to be rolled back
func ExpectRollback() Mock {
	e := &expectedRollback{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// The expectation will return an error
func (c *conn) WillReturnError(err error) Mock {
	c.active.setError(err)
	return c
}

// Expect database Exec to be triggered, which will match
// the given query string as a regular expression
func ExpectExec(sqlRegexStr string) Mock {
	e := &expectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// Expect database Query to be triggered, which will match
// the given query string as a regular expression
func ExpectQuery(sqlRegexStr string) Mock {
	e := &expectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)

	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// The expectation should be called with given arguments.
// Works with Exec and Query expectations
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

// The expectation will return a Result.
// Works only with Exec expectations
func (c *conn) WillReturnResult(result driver.Result) Mock {
	eq, ok := c.active.(*expectedExec)
	if !ok {
		panic(fmt.Sprintf("driver.Result may be returned only by Exec expectations, current is %T", c.active))
	}
	eq.result = result
	return c
}

// The expectation will return Rows.
// Works only with Query expectations
func (c *conn) WillReturnRows(rows driver.Rows) Mock {
	eq, ok := c.active.(*expectedQuery)
	if !ok {
		panic(fmt.Sprintf("driver.Rows may be returned only by Query expectations, current is %T", c.active))
	}
	eq.rows = rows
	return c
}
