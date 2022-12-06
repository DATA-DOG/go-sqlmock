//go:build go1.8
// +build go1.8

package sqlmock

import (
	"context"
	"database/sql/driver"
)

var _ driver.Stmt = (*statement)(nil)
var _ driver.StmtExecContext = (*statement)(nil)
var _ driver.StmtQueryContext = (*statement)(nil)

// ExecContext Implement the "StmtExecContext" interface
func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return stmt.conn.ExecContext(ctx, stmt.query, args)
}

// QueryContext Implement the "StmtQueryContext" interface
func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return stmt.conn.QueryContext(ctx, stmt.query, args)
}

// Deprecated: Drivers should implement ExecerContext instead.
func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.conn.ExecContext(context.Background(), stmt.query, convertValueToNamedValue(args))
}

// Deprecated: Drivers should implement StmtQueryContext instead (or additionally).
func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.conn.QueryContext(context.Background(), stmt.query, convertValueToNamedValue(args))
}

func convertValueToNamedValue(args []driver.Value) []driver.NamedValue {
	namedArgs := make([]driver.NamedValue, len(args))
	for i, v := range args {
		namedArgs[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return namedArgs
}
