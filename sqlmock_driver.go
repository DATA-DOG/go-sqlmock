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
	"database/sql/driver"
	"fmt"
	"time"
)

var _ driver.Conn = (*sqlmock)(nil)
var _ driver.Tx = (*sqlmock)(nil)

// Close a mock database driver connection. It may or may not
// be called depending on the circumstances, but if it is called
// there must be an *ExpectedClose expectation satisfied.
// meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *sqlmock) Close() error {
	c.drv.Lock()
	defer c.drv.Unlock()

	c.opened--
	if c.opened == 0 {
		delete(c.drv.connMap, c.dsn)
	}

	var expected *ExpectedClose
	var fulfilled int
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok = next.(*ExpectedClose); ok {
			break
		}

		next.Unlock()
		if c.ordered {
			return fmt.Errorf("call to database Close, was not expected, next expectation is: %s", next)
		}
	}

	if expected == nil {
		msg := "call to database Close was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return fmt.Errorf(msg)
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
}

// Begin meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *sqlmock) Begin() (driver.Tx, error) {
	ex, err := c.begin()
	if ex != nil {
		time.Sleep(ex.delay)
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *sqlmock) begin() (*ExpectedBegin, error) {
	var expected *ExpectedBegin
	var ok bool
	var fulfilled int
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok = next.(*ExpectedBegin); ok {
			break
		}

		next.Unlock()
		if c.ordered {
			return nil, fmt.Errorf("call to database transaction Begin, was not expected, next expectation is: %s", next)
		}
	}
	if expected == nil {
		msg := "call to database transaction Begin was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg)
	}

	expected.triggered = true
	expected.Unlock()

	return expected, expected.err
}

// Prepare meets http://golang.org/pkg/database/sql/driver/#Conn interface
func (c *sqlmock) Prepare(query string) (driver.Stmt, error) {
	ex, err := c.prepare(query)
	if ex != nil {
		time.Sleep(ex.delay)
	}
	if err != nil {
		return nil, err
	}

	return &statement{c, ex, query}, nil
}

func (c *sqlmock) prepare(query string) (*ExpectedPrepare, error) {
	var expected *ExpectedPrepare
	var fulfilled int
	var ok bool

	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if c.ordered {
			if expected, ok = next.(*ExpectedPrepare); ok {
				break
			}

			next.Unlock()
			return nil, fmt.Errorf("call to Prepare statement with query '%s', was not expected, next expectation is: %s", query, next)
		}

		if pr, ok := next.(*ExpectedPrepare); ok {
			if err := c.queryMatcher.Match(pr.expectSQL, query); err == nil {
				expected = pr
				break
			}
		}
		next.Unlock()
	}

	if expected == nil {
		msg := "call to Prepare '%s' query was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg, query)
	}
	defer expected.Unlock()
	if err := c.queryMatcher.Match(expected.expectSQL, query); err != nil {
		return nil, fmt.Errorf("Prepare: %v", err)
	}

	expected.triggered = true
	return expected, expected.err
}

// Commit meets http://golang.org/pkg/database/sql/driver/#Tx
func (c *sqlmock) Commit() error {
	var expected *ExpectedCommit
	var fulfilled int
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok = next.(*ExpectedCommit); ok {
			break
		}

		next.Unlock()
		if c.ordered {
			return fmt.Errorf("call to Commit transaction, was not expected, next expectation is: %s", next)
		}
	}
	if expected == nil {
		msg := "call to Commit transaction was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return fmt.Errorf(msg)
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
}

// Rollback meets http://golang.org/pkg/database/sql/driver/#Tx
func (c *sqlmock) Rollback() error {
	var expected *ExpectedRollback
	var fulfilled int
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok = next.(*ExpectedRollback); ok {
			break
		}

		next.Unlock()
		if c.ordered {
			return fmt.Errorf("call to Rollback transaction, was not expected, next expectation is: %s", next)
		}
	}
	if expected == nil {
		msg := "call to Rollback transaction was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return fmt.Errorf(msg)
	}

	expected.triggered = true
	expected.Unlock()
	return expected.err
}
