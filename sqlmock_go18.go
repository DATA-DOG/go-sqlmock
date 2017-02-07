// +build go1.8

package sqlmock

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"
)

var ErrCancelled = errors.New("canceling query due to user request")

// Implement the "QueryerContext" interface
func (c *sqlmock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.query(query, namedArgs)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		return ex.rows, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// Implement the "ExecerContext" interface
func (c *sqlmock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.exec(query, namedArgs)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		return ex.result, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// Implement the "ConnBeginTx" interface
func (c *sqlmock) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	ex, err := c.begin()
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		return c, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// Implement the "ConnPrepareContext" interface
func (c *sqlmock) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	ex, err := c.prepare(query)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(ex.delay):
		return &statement{c, query, ex.closeErr}, nil
	case <-ctx.Done():
		return nil, ErrCancelled
	}
}

// @TODO maybe add ExpectedBegin.WithOptions(driver.TxOptions)
