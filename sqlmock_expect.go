/*
Package sqlmock is a mock library implementing sql driver. Which has one and only
purpose - to simulate any sql driver behavior in tests, without needing a real
database connection. It helps to maintain correct **TDD** workflow.

It does not require any modifications to your source code in order to test
and mock database operations. Supports concurrency and multiple database mocking.

The driver allows to mock any sql driver method behavior.
*/
package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

func (c *sqlmock) ExpectOperation(arg Argument) *ExpectedOperation {
	var match = AnyArg()
	if arg != nil {
		match = arg
	}

	e := &ExpectedOperation{arg: match}
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) open(options []func(*sqlmock) error) (*sql.DB, Sqlmock, error) {
	db, err := sql.Open("sqlmock", c.dsn)
	if err != nil {
		return db, c, err
	}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return db, c, err
		}
	}
	if c.converter == nil {
		c.converter = driver.DefaultParameterConverter
	}
	if c.queryMatcher == nil {
		c.queryMatcher = QueryMatcherRegexp
	}

	if c.monitorPings {
		// We call Ping on the driver shortly to verify startup assertions by
		// driving internal behaviour of the sql standard library. We don't
		// want this call to ping to be monitored for expectation purposes so
		// temporarily disable.
		c.monitorPings = false
		defer func() { c.monitorPings = true }()
	}
	return db, c, db.Ping()
}

func (c *sqlmock) ExpectClose() *ExpectedClose {
	e := &ExpectedClose{}
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) MatchExpectationsInOrder(b bool) {
	c.ordered = b
}

func (c *sqlmock) ExpectationsWereMet() error {
	for _, e := range c.expected {
		e.Lock()
		fulfilled := e.fulfilled()
		e.Unlock()

		if !fulfilled {
			return fmt.Errorf("there is a remaining expectation which was not matched: %s", e)
		}

		// for expected prepared statement check whether it was closed if expected
		if prep, ok := e.(*ExpectedPrepare); ok {
			if prep.mustBeClosed && !prep.wasClosed {
				return fmt.Errorf("expected prepared statement to be closed, but it was not: %s", prep)
			}
		}

		// must check whether all expected queried rows are closed
		if query, ok := e.(*ExpectedQuery); ok {
			if query.rowsMustBeClosed && !query.rowsWereClosed {
				return fmt.Errorf("expected query rows to be closed, but it was not: %s", query)
			}
		}
	}
	return nil
}

func (c *sqlmock) ExpectBegin() *ExpectedBegin {
	e := &ExpectedBegin{}
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) ExpectExec(expectedSQL string) *ExpectedExec {
	e := &ExpectedExec{}
	e.expectSQL = expectedSQL
	e.converter = c.converter
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) ExpectPrepare(expectedSQL string) *ExpectedPrepare {
	e := &ExpectedPrepare{expectSQL: expectedSQL, mock: c}
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) ExpectQuery(expectedSQL string) *ExpectedQuery {
	e := &ExpectedQuery{}
	e.expectSQL = expectedSQL
	e.converter = c.converter
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) ExpectCommit() *ExpectedCommit {
	e := &ExpectedCommit{}
	c.expected = append(c.expected, e)
	return e
}

func (c *sqlmock) ExpectRollback() *ExpectedRollback {
	e := &ExpectedRollback{}
	c.expected = append(c.expected, e)
	return e
}

// NewRows allows Rows to be created from a
// sql driver.Value slice or from the CSV string and
// to be used as sql driver.Rows.
func (c *sqlmock) NewRows(columns []string) *Rows {
	r := NewRows(columns)
	r.converter = c.converter
	return r
}
