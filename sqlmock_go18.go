// +build go1.8

package sqlmock

import (
	"context"
	"database/sql/driver"
	"fmt"
)

var CancelledStatementErr = fmt.Errorf("canceling query due to user request")

// Implement the "QueryerContext" interface
func (c *sqlmock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.queryExpectation(query, namedArgs)
	if err != nil {
		return nil, err
	}

	type result struct {
		rows driver.Rows
		err  error
	}

	exec := make(chan result)
	defer func() {
		close(exec)
	}()

	go func() {
		rows, err := c.query(ex)
		exec <- result{rows, err}
	}()

	select {
	case res := <-exec:
		return res.rows, res.err
	case <-ctx.Done():
		return nil, CancelledStatementErr
	}
}

// Implement the "ExecerContext" interface
func (c *sqlmock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	namedArgs := make([]namedValue, len(args))
	for i, nv := range args {
		namedArgs[i] = namedValue(nv)
	}

	ex, err := c.execExpectation(query, namedArgs)
	if err != nil {
		return nil, err
	}

	type result struct {
		rs  driver.Result
		err error
	}

	exec := make(chan result)
	defer func() {
		close(exec)
	}()

	go func() {
		rs, err := c.exec(ex)
		exec <- result{rs, err}
	}()

	select {
	case res := <-exec:
		return res.rs, res.err
	case <-ctx.Done():
		return nil, CancelledStatementErr
	}
}

// Implement the "ConnBeginTx" interface
func (c *sqlmock) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	ex, err := c.beginExpectation()
	if err != nil {
		return nil, err
	}

	type result struct {
		tx  driver.Tx
		err error
	}

	exec := make(chan result)
	defer func() {
		close(exec)
	}()

	go func() {
		tx, err := c.begin(ex)
		exec <- result{tx, err}
	}()

	select {
	case res := <-exec:
		return res.tx, res.err
	case <-ctx.Done():
		return nil, CancelledStatementErr
	}
}

// Implement the "ConnPrepareContext" interface
func (c *sqlmock) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	ex, err := c.prepareExpectation(query)
	if err != nil {
		return nil, err
	}

	type result struct {
		stmt driver.Stmt
		err  error
	}

	exec := make(chan result)
	defer func() {
		close(exec)
	}()

	go func() {
		stmt, err := c.prepare(ex, query)
		exec <- result{stmt, err}
	}()

	select {
	case res := <-exec:
		return res.stmt, res.err
	case <-ctx.Done():
		return nil, CancelledStatementErr
	}
}

// @TODO maybe add ExpectedBegin.WithOptions(driver.TxOptions)
