// +build go1.8

package sqlmock

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"time"
)

// ErrCancelled defines an error value, which can be expected in case of
// such cancellation error.
var ErrCancelled = errors.New("canceling query due to user request")

// Implement the "QueryerContext" interface
func (c *sqlmock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.query(query, namedArgs)
	if ex != nil {
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

	return nil, err
}

// Implement the "ExecerContext" interface
func (c *sqlmock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.exec(query, namedArgs)
	if ex != nil {
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

	return nil, err
}

// Implement the "ConnBeginTx" interface
func (c *sqlmock) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	ex, err := c.begin()
	if ex != nil {
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

	return nil, err
}

// Implement the "ConnPrepareContext" interface
func (c *sqlmock) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	ex, err := c.prepare(query)
	if ex != nil {
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

	return nil, err
}

// Implement the "Pinger" interface - the explicit DB driver ping was only added to database/sql in Go 1.8
func (c *sqlmock) Ping(ctx context.Context) error {
	if !c.monitorPings {
		return nil
	}

	ex, err := c.ping()
	if ex != nil {
		select {
		case <-ctx.Done():
			return ErrCancelled
		case <-time.After(ex.delay):
		}
	}

	return err
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

// Implement the "StmtExecContext" interface
func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return stmt.conn.ExecContext(ctx, stmt.query, args)
}

// Implement the "StmtQueryContext" interface
func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return stmt.conn.QueryContext(ctx, stmt.query, args)
}

func (c *sqlmock) ExpectPing() *ExpectedPing {
	if !c.monitorPings {
		log.Println("ExpectPing will have no effect as monitoring pings is disabled. Use MonitorPingsOption to enable.")
		return nil
	}
	e := &ExpectedPing{}
	c.expected = append(c.expected, e)
	return e
}

// @TODO maybe add ExpectedBegin.WithOptions(driver.TxOptions)
