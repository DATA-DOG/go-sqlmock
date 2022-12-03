//go:build go1.8
// +build go1.8

package sqlmock

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

var _ driver.QueryerContext = (*sqlmock)(nil)
var _ driver.ConnPrepareContext = (*sqlmock)(nil)
var _ driver.ExecerContext = (*sqlmock)(nil)
var _ driver.ConnBeginTx = (*sqlmock)(nil)

// Sqlmock interface for Go 1.8+
type Sqlmock interface {
	// Common Embed common methods
	Common
}

// ErrCancelled defines an error value, which can be expected in case of
// such cancellation error.
var ErrCancelled = errors.New("canceling query due to user request")

// QueryContext Implement the "QueryerContext" interface
func (c *sqlmock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	ex, err := c.query(query, args)
	if ex == nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		if err != nil {
			return nil, err
		}
		return ex.rows, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// ExecContext Implement the "ExecerContext" interface
func (c *sqlmock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ex, err := c.exec(query, args)
	if ex == nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		if err != nil {
			return nil, err
		}
		return ex.result, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// BeginTx Implement the "ConnBeginTx" interface
func (c *sqlmock) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	ex, err := c.begin()
	if ex == nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		if err != nil {
			return nil, err
		}
		return c, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// PrepareContext Implement the "ConnPrepareContext" interface
func (c *sqlmock) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	ex, err := c.prepare(query)
	if ex == nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		if err != nil {
			return nil, err
		}
		return &statement{c, ex, query}, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// Ping Implement the "Pinger" interface - the explicit DB driver ping was only added to database/sql in Go 1.8
func (c *sqlmock) Ping(ctx context.Context) error {
	if !c.monitorPings {
		return nil
	}

	ex, err := c.ping()
	if ex == nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ErrCancelled
	case <-time.After(ex.delay):
		return err
	}
}

func (c *sqlmock) ping() (*ExpectedPing, error) {
	var expected *ExpectedPing
	var fulfilled int
	var ok bool
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok = next.(*ExpectedPing); ok {
			break
		}

		next.Unlock()
		if c.ordered {
			return nil, fmt.Errorf("call to database Ping, was not expected, next expectation is: %s", next)
		}
	}

	if expected == nil {
		msg := "call to database Ping was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg)
	}

	expected.triggered = true
	expected.Unlock()
	return expected, expected.err
}

// Query meets http://golang.org/pkg/database/sql/driver/#Queryer
// Deprecated: Drivers should implement QueryerContext instead.
func (c *sqlmock) Query(query string, args []driver.Value) (driver.Rows, error) {
	ex, err := c.query(query, convNameValue(args))
	if ex != nil {
		time.Sleep(ex.delay)
	}
	if err != nil {
		return nil, err
	}

	return ex.rows, nil
}

func (c *sqlmock) doSql(opt string, query string, args []driver.NamedValue) (*ExpectedSql, error) {
	var expected *ExpectedSql
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
			if expected, ok = next.(*ExpectedSql); ok {
				break
			}
			next.Unlock()
			return nil, fmt.Errorf("call to Query '%s' with args %+v, was not expected, next expectation is: %s", query, args, next)
		}

		if qr, ok := next.(*ExpectedSql); ok {
			if err := c.queryMatcher.Match(qr.expectSQL, query); err != nil {
				next.Unlock()
				continue
			}

			if qr.checkArgs != nil {
				if err := qr.checkArgs(query, args); err == nil {
					expected = qr
					break
				}
			} else {
				if err := qr.attemptArgMatch(args); err == nil {
					expected = qr
					break
				}
			}
		}
		next.Unlock()
	}

	if expected == nil {
		msg := "call to Query '%s' with args %+v was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg, query, args)
	}

	defer expected.Unlock()

	if err := c.queryMatcher.Match(expected.expectSQL, query); err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	if expected.checkArgs != nil {
		if err := expected.checkArgs(query, args); err != nil {
			return nil, fmt.Errorf("query '%s', arguments do not match: %s", query, err)
		}
	} else {
		if err := expected.argsMatches(args); err != nil {
			return nil, fmt.Errorf("query '%s', arguments do not match: %s", query, err)
		}
	}

	expected.triggered = true
	if expected.err != nil {
		return expected, expected.err // mocked to return error
	}

	if expected.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.Rows, but it was not set for expectation %T as %+v", query, args, expected, expected)
	}
	return expected, nil
}

func (c *sqlmock) query(query string, args []driver.NamedValue) (*ExpectedQuery, error) {
	var expected *ExpectedQuery
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
			if expected, ok = next.(*ExpectedQuery); ok {
				break
			}
			next.Unlock()
			return nil, fmt.Errorf("call to Query '%s' with args %+v, was not expected, next expectation is: %s", query, args, next)
		}

		if qr, ok := next.(*ExpectedQuery); ok {
			if err := c.queryMatcher.Match(qr.expectSQL, query); err != nil {
				next.Unlock()
				continue
			}

			if qr.checkArgs != nil {
				if err := qr.checkArgs(query, args); err == nil {
					expected = qr
					break
				}
			} else {
				if err := qr.attemptArgMatch(args); err == nil {
					expected = qr
					break
				}
			}
		}
		next.Unlock()
	}

	if expected == nil {
		msg := "call to Query '%s' with args %+v was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg, query, args)
	}

	defer expected.Unlock()

	if err := c.queryMatcher.Match(expected.expectSQL, query); err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	if expected.checkArgs != nil {
		if err := expected.checkArgs(query, args); err != nil {
			return nil, fmt.Errorf("query '%s', arguments do not match: %s", query, err)
		}
	} else {
		if err := expected.argsMatches(args); err != nil {
			return nil, fmt.Errorf("query '%s', arguments do not match: %s", query, err)
		}
	}

	expected.triggered = true
	if expected.err != nil {
		return expected, expected.err // mocked to return error
	}

	if expected.rows == nil {
		return nil, fmt.Errorf("query '%s' with args %+v, must return a database/sql/driver.Rows, but it was not set for expectation %T as %+v", query, args, expected, expected)
	}
	return expected, nil
}

// Exec meets http://golang.org/pkg/database/sql/driver/#Execer
// Deprecated: Drivers should implement ExecerContext instead.
func (c *sqlmock) Exec(query string, args []driver.Value) (driver.Result, error) {
	ex, err := c.exec(query, convNameValue(args))
	if ex != nil {
		time.Sleep(ex.delay)
	}
	if err != nil {
		return nil, err
	}

	return ex.result, nil
}

func (c *sqlmock) exec(query string, args []driver.NamedValue) (*ExpectedExec, error) {
	var expected *ExpectedExec
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
			if expected, ok = next.(*ExpectedExec); ok {
				break
			}
			next.Unlock()
			return nil, fmt.Errorf("call to ExecQuery '%s' with args %+v, was not expected, next expectation is: %s", query, args, next)
		}

		if exec, ok := next.(*ExpectedExec); ok {
			if err := c.queryMatcher.Match(exec.expectSQL, query); err != nil {
				next.Unlock()
				continue
			}

			if exec.checkArgs != nil {
				if err := exec.checkArgs(query, args); err == nil {
					expected = exec
					break
				}
			} else {
				if err := exec.attemptArgMatch(args); err == nil {
					expected = exec
					break
				}
			}
		}
		next.Unlock()
	}

	if expected == nil {
		msg := "call to ExecQuery '%s' with args %+v was not expected"
		if fulfilled == len(c.expected) {
			msg = "all expectations were already fulfilled, " + msg
		}
		return nil, fmt.Errorf(msg, query, args)
	}

	defer expected.Unlock()

	if err := c.queryMatcher.Match(expected.expectSQL, query); err != nil {
		return nil, fmt.Errorf("ExecQuery: %v", err)
	}

	if expected.checkArgs != nil {
		if err := expected.checkArgs(query, args); err != nil {
			return nil, fmt.Errorf("ExecQuery '%s', arguments do not match: %s", query, err)
		}
	} else {
		if err := expected.argsMatches(args); err != nil {
			return nil, fmt.Errorf("ExecQuery '%s', arguments do not match: %s", query, err)
		}
	}

	expected.triggered = true
	if expected.err != nil {
		return expected, expected.err // mocked to return error
	}

	if expected.result == nil {
		return nil, fmt.Errorf("ExecQuery '%s' with args %+v, must return a database/sql/driver.Result, but it was not set for expectation %T as %+v", query, args, expected, expected)
	}

	return expected, nil
}
